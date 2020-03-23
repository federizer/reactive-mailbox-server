package services

import (
	"database/sql"
	"net/http"
)

type SystemStorageImpl struct {
	DB *sql.DB
}

func (s *SystemStorageImpl) Alive(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, I am alive!"))
}
