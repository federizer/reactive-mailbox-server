package cmd

import (
	"context"
	"database/sql"
	"fmt"
	pbauth "github.com/federizer/reactive-mailbox/api/generated/auth"
	pbsystem "github.com/federizer/reactive-mailbox/api/generated/system"
	grpcservices "github.com/federizer/reactive-mailbox/grpc_services"
	"github.com/federizer/reactive-mailbox/pkg/config"
	"github.com/federizer/reactive-mailbox/repository"
	services "github.com/federizer/reactive-mailbox/services"
	"github.com/federizer/reactive-mailbox/session"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:     "serve --config [ config file ]",
	Short:   "Start serving requests.",
	Long:    ``,
	Example: "email-services serve --config examples/config-dev.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		if err := serve(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

/*func jsonStringToMap(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if f != reflect.String || t != reflect.Map {
		return data, nil
	}
	raw := data.(string)
	if raw == "" {
		return nil, nil
	}
	var ret map[string]string
	err := json.Unmarshal([]byte(raw), &ret)
	if err != nil {
		log.Printf("ignored error trying to parse %q as json: %v", raw, err)
		return data, nil
	}
	return ret, nil
}*/

/*func floatToDecimal() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Float64 {
			return data, nil
		}
		if t != reflect.TypeOf(decimal.Decimal{}) {
			return data, nil
		}

		// Convert it by parsing
		// dc := decimal.NewFromFloat(data.(float64))

		return nil, nil
	}
}

func decodeHookWithTag(hook mapstructure.DecodeHookFunc, tagName string) viper.DecoderConfigOption {
	return func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = hook
		c.TagName = tagName
	}
}*/

func serve() error {
	// unmarshal config into Struct
	var c config.Config

	/*decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		jsonStringToMap,
	)

	err := viper.Unmarshal(&c, viper.DecodeHook(decodeHook), func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})*/

	/*decodeHookFunc := floatToDecimal()
	decoderConfigOption := decodeHookWithTag(decodeHookFunc, "json")

	err := viper.Unmarshal(&c, decoderConfigOption)*/

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigc
		signal.Stop(sigc)
		cancel()
	}()

	configFile := viper.ConfigFileUsed()
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", configFile, err)
	}

	err = yaml.Unmarshal(configData, &c)
	// err := viper.Unmarshal(&c)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %v", err)
	}

	if err := c.Validate(); err != nil {
		return err
	}

	var (
		grpcServer    *grpc.Server
		wrappedServer *grpcweb.WrappedGrpcServer
	)

	db, err := c.Database.Config.Open()
	if db != nil {
		defer db.Close()
	}

	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	log.Infof("config database: %s", c.Database.Type)

	DB, _ := db.(*sql.DB)

	if c.Web.TLSCert != "" {
		creds, err := credentials.NewServerTLSFromFile(c.Web.TLSCert, c.Web.TLSKey)
		if err != nil {
			log.Fatalf("Failed while obtaining TLS certificates. Error: %+v", err)
		}

		/*opts := []grpcrecovery.Option{
			grpcrecovery.WithRecoveryHandler(GrpcRecoveryHandlerFunc),
		}*/

		grpcServer = grpc.NewServer(
			grpc.Creds(creds),
			grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(
				StreamAuthInterceptor,
				// grpcrecovery.StreamServerInterceptor(opts...),
				grpcrecovery.StreamServerInterceptor(),
			)),
			grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(
				UnaryAuthInterceptor,
				// grpcrecovery.UnaryServerInterceptor(opts...),
				grpcrecovery.UnaryServerInterceptor(),
			)),
		)
	}

	pbsystem.RegisterSystemServiceServer(grpcServer, &grpcservices.SystemStorageImpl{DB})
	pbauth.RegisterAuthServiceServer(grpcServer, &grpcservices.AuthStorageImpl{DB: DB,
		SignUpClients: make(map[string]pbauth.AuthService_SignupServer),
		SignInClients: make(map[string]pbauth.AuthService_SigninServer),
		Mu:            sync.RWMutex{}})
	// pb.RegisterDraftServiceServer(grpcServer, &database.DraftServerImpl{DB})
	// pb.RegisterDraftServiceServer(grpcServer, &database.MessageServerImpl{DB})
	// pb.RegisterAuthServer(grpcServer, &oidc.UserInfoImpl{})
	// pb.RegisterUsersServer(grpcServer, &impl.UserServerImpl{db})
	// pb.RegisterJmsApiServer(grpcServer, &impl.JmapServerImpl{db})
	// pb.RegisterGmailServer(grpcServer, &impl.GmailServerImpl{db})
	// pb.RegisterGmailServer(grpcServer, &impl.LabelsServerImpl{db})
	// pb.RegisterFilesServer(grpcServer, &impl.FileServerImpl{db})

	wrappedServer = grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool {
			for _, allowedOrigin := range c.Web.AllowedOrigins {
				if allowedOrigin == "*" || origin == allowedOrigin {
					return true
				}
			}
			return false
		}),
	)

	restMux := http.NewServeMux() // http.DefaultServeMux

	httpServer := http.Server{
		Addr: c.Web.Addr(),
		Handler: http.HandlerFunc(
			func(resp http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodOptions {
					for _, allowedOrigin := range c.Web.AllowedOrigins {
						if allowedOrigin == "*" || allowedOrigin == req.Header.Get("Origin") {
							SetCors(resp, allowedOrigin)
							return
						}
					}
					return
				}

				if IsGrpcRequest(req) {
					if wrappedServer.IsGrpcWebRequest(req) {
						wrappedServer.ServeHTTP(resp, req)
					} else {
						grpcServer.ServeHTTP(resp, req)
					}
				} else {
					restMux.ServeHTTP(resp, req)
				}
			},
		),
	}

	systemstorage := &services.SystemStorageImpl{DB}
	draftstorage := &services.DraftStorageImpl{DB}

	restMux.HandleFunc("/alive", systemstorage.Alive)
	restMux.HandleFunc("/drafts", draftstorage.ListDrafts)
	// restMux.HandleFunc(common.AUTH_SIGNIN_PATH, oidc.SignIn())
	// restMux.HandleFunc(common.AUTH_SIGNOUT_PATH, oidc.SignOut())
	// restMux.HandleFunc(common.AUTH_REDIRECT_PATH, oidc.HandleCallback())
	restMux.HandleFunc(c.File.UploadApi, session.AuthorizeRest(repository.FileUploadHandler))

	// make a channel for each server to fail properly
	errc := make(chan error, 1)

	go func() {
		errc <- httpServer.ListenAndServeTLS(c.Web.TLSCert, c.Web.TLSKey)
		close(errc)
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return err
	}
}

/*func (s *logItServer) IAmAlive(ctx context.Context, in *pb.LogData) (*pb.LogSuccess, error) {
	return &pb.LogSuccess{Status: "success", Msg: ""}, nil
}

func Alive(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, I am alive!")
}*/

func SetCors(w http.ResponseWriter, allowedOrigin string) {
	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-grpc-web, sessionid, x-user-agent, authorization-token")
	w.Header().Set("Access-Control-Max-Age", "600")
}

func IsGrpcRequest(req *http.Request) bool {
	return strings.Contains(req.Header.Get("Content-Type"), "application/grpc")
}

/*func GrpcRecoveryHandlerFunc(p interface{}) error {
	fmt.Printf("p: %+v\n", p)
	return status.Errorf(codes.Internal, "Unexpected error")
}*/

func StreamAuthInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	ctx, err := session.AuthorizeGrpc(stream.Context())
	if err != nil {
		log.Errorf("%s %v", info.FullMethod, err)
		return err
	}

	wrapped := grpcmiddleware.WrapServerStream(stream)
	wrapped.WrappedContext = ctx
	return handler(srv, wrapped)
}

func UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, err := session.AuthorizeGrpc(ctx)
	if err != nil {
		log.Errorf("%s %v", info.FullMethod, err)
		return nil, err
	}

	return handler(ctx, req)
}
