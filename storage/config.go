package storage

import (
	"database/sql"
	"fmt"
	mysql "github.com/go-sql-driver/mysql"
	postgres "github.com/lib/pq"
	sqlite3 "github.com/mattn/go-sqlite3"
)

type Database interface {
	Close() error
	// ListDrafts() ([]Draft, error)
}

type SQLite3 struct {
	// File to
	File string `yaml:"file"`
}

func (s *SQLite3) Open() (Database, error) {
	db, err := s.open()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *SQLite3) open() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", s.File)
	if err != nil {
		sqlErr, ok := err.(sqlite3.Error)
		if !ok {
			return nil, err
		}
		return nil, sqlErr.ExtendedCode
	}
	if s.File == ":memory:" {
		db.SetMaxOpenConns(1)
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

type Postgres struct {
	Database string
	User     string
	Password string
	Host     string
	Port     uint16
}

func (p *Postgres) Open() (Database, error) {
	db, err := p.open()
	if err != nil {
		sqlErr, ok := err.(postgres.Error)
		if !ok {
			return nil, err
		}
		return nil, sqlErr
	}
	return db, nil
}

func (p *Postgres) open() (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", p.User, p.Password, p.Host, p.Port, p.Database)
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

type MySQL struct {
	Database string
	User     string
	Password string
	Host     string
	Port     uint16
}

func (s *MySQL) Open() (Database, error) {
	db, err := s.open()
	if err != nil {
		sqlErr, ok := err.(*mysql.MySQLError)
		if !ok {
			return nil, err
		}
		return nil, sqlErr
	}
	return db, nil
}

func (s *MySQL) open() (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&tls=false", s.User, s.Password, s.Host, s.Port, s.Database)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
