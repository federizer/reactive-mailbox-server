package grpc_services

import (
	"context"
	"database/sql"
	"errors"
	pbauth "github.com/federizer/reactive-mailbox/api/generated/auth"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"sync"
)

type AuthStorageImpl struct {
	DB *sql.DB
	// List of connected clients
	SignUpClients map[string]pbauth.AuthService_SignupServer
	SignInClients map[string]pbauth.AuthService_SigninServer
	Mu            sync.RWMutex
}

func (s *AuthStorageImpl) addSignUpClient(uid string, srv pbauth.AuthService_SignupServer) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.SignUpClients[uid] = srv
}

func (s *AuthStorageImpl) removeSignUpClient(uid string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.SignUpClients, uid)
}

func (s *AuthStorageImpl) addSignInClient(uid string, srv pbauth.AuthService_SigninServer) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.SignInClients[uid] = srv
}

func (s *AuthStorageImpl) removeSignInClient(uid string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.SignInClients, uid)
}

func (s *AuthStorageImpl) Signup(in *pbauth.SignUpRequest, srv pbauth.AuthService_SignupServer) error {
	uid := uuid.Must(uuid.NewRandom()).String()
	log.Printf("new session: %s", uid)

	s.addSignUpClient(uid, srv)
	defer s.removeSignUpClient(uid)

	authSession := pbauth.AuthSession{}
	authSession.State = pbauth.AuthState_USER_SIGNED_UP
	authSession.AccessToken = "accessToken123"
	authSession.RefreshToken = "refreshToken123"

	if err := srv.Send(&authSession); err != nil {
		log.Printf("send err: %v", err)
		return err
	}

	for {
		select {
		case <-srv.Context().Done():
			return srv.Context().Err()
		}
	}
}

func (s *AuthStorageImpl) Signin(in *pbauth.SignInRequest, srv pbauth.AuthService_SigninServer) error {
	uid := uuid.Must(uuid.NewRandom()).String()
	log.Printf("new session: %s", uid)

	s.addSignInClient(uid, srv)
	defer s.removeSignInClient(uid)

	authSession := pbauth.AuthSession{}
	authSession.State = pbauth.AuthState_USER_SIGNED_IN
	authSession.AccessToken = "accessToken123"
	authSession.RefreshToken = "refreshToken123"

	if err := srv.Send(&authSession); err != nil {
		log.Printf("send err: %v", err)
		return err
	}

	for {
		select {
		case <-srv.Context().Done():
			return srv.Context().Err()
		}
	}
}

func (s *AuthStorageImpl) Signout(ctx context.Context, in *pbauth.SignOutRequest) (*pbauth.SignOutResponse, error) {
	s.Mu.RLock()
	client := s.SignInClients[in.RefreshToken]
	s.Mu.RUnlock()

	if client == nil {
		return nil, errors.New("client not found")
	}

	authSession := pbauth.AuthSession{}
	authSession.State = pbauth.AuthState_USER_SIGNED_OUT

	if err := client.Send(&authSession); err != nil {
		log.Printf("send err: %v", err)
		return nil, err
	}

	signoutResponse := pbauth.SignOutResponse{}
	signoutResponse.State = pbauth.AuthState_USER_SIGNED_OUT

	return &signoutResponse, nil
}
