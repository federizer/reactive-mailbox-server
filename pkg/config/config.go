package config

import (
	// "database/sql"
	"fmt"
	"github.com/federizer/reactive-mailbox/storage"
	"strings"
)

// Config is the config format for the main application.
type Config struct {
	Web      Web      `yaml:"web"`
	Database Database `yaml:"database"`
	File     File     `yaml:"file"`
}

//Validate the configuration
func (c Config) Validate() error {
	// Fast checks. Perform these first for a more responsive CLI.
	checks := []struct {
		bad    bool
		errMsg string
	}{
		{c.Database.Config == nil, "no database supplied in config file"},
		{c.Web.Protocol != "http" && c.Web.Protocol != "https", "must supply 'http' or 'https' Protocol"},
		{c.Web.Host == "", "must supply a Host to listen on"},
		{c.Web.Protocol == "https" && (c.Web.TLSCert == "" || c.Web.TLSKey == ""), "must specific both a TLS cert and key"},
		{len(c.Web.AllowedOrigins) == 0, "must specify at least one Allowed Origin"},

		{c.File.UploadApi == "", "must supply a File Upload API to upload files to the storage"},
	}

	var checkErrors []string

	for _, check := range checks {
		if check.bad {
			checkErrors = append(checkErrors, check.errMsg)
		}
	}
	if len(checkErrors) != 0 {
		return fmt.Errorf("invalid Config:\n\t-\t%s", strings.Join(checkErrors, "\n\t-\t"))
	}
	return nil
}

// Web is the config format for the HTTP server.
type Web struct {
	Protocol       string   `yaml:"protocol"`
	Host           string   `yaml:"host"`
	Port           string   `yaml:"port"`
	TLSCert        string   `yaml:"tlsCert"`
	TLSKey         string   `yaml:"tlsKey"`
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

func (w Web) Addr() string {
	addr := w.Host
	if w.Port != "" {
		addr += ":" + w.Port
	}
	return addr
}

type Database struct {
	Type   string         `yaml:"type"`
	Config DatabaseConfig `yaml:"config"`
}

type DatabaseConfig interface {
	Open() (storage.Database, error)
}

var databases = map[string]func() DatabaseConfig{
	"sqlite3":  func() DatabaseConfig { return &storage.SQLite3{} },
	"postgres": func() DatabaseConfig { return &storage.Postgres{} },
	"mysql":    func() DatabaseConfig { return &storage.MySQL{} },
}

func (s *Database) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var dbType struct {
		Type string `yaml:"type"`
	}

	if err := unmarshal(&dbType); err != nil {
		return fmt.Errorf("parse database: %v", err)
	}

	s.Type = dbType.Type

	switch s.Type {
	case "sqlite3":
		var dbConfig struct {
			Config storage.SQLite3 `yaml:"config"`
		}

		if err := unmarshal(&dbConfig); err != nil {
			return fmt.Errorf("parse database: %v", err)
		}
		s.Config = &dbConfig.Config
	case "postgres":
		var dbConfig struct {
			Config storage.Postgres `yaml:"config"`
		}

		if err := unmarshal(&dbConfig); err != nil {
			return fmt.Errorf("parse database: %v", err)
		}
		s.Config = &dbConfig.Config
	case "mysql":
		var dbConfig struct {
			Config storage.MySQL `yaml:"config"`
		}

		if err := unmarshal(&dbConfig); err != nil {
			return fmt.Errorf("parse database: %v", err)
		}
		s.Config = &dbConfig.Config
	default:
		return fmt.Errorf("unknown connector type %q", s.Type)
	}

	return nil
}

type File struct {
	UploadApi string `yaml:"uploadApi"`
	Folder    string `yaml:"folder"`
}
