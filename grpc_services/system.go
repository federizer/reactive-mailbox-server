package grpc_services

import (
	"context"
	"database/sql"
	"github.com/golang/protobuf/ptypes/empty"

	// storage "github.com/federizer/reactive-mailbox/storage/sql/postgres"
	// "github.com/golang/protobuf/ptypes/empty"

	// biz "pes/common"
	pbsystem "github.com/federizer/reactive-mailbox/api/generated/system"
)

const (
	systemTable = "system"
)

type SystemStorageImpl struct {
	DB *sql.DB
}

func (s *SystemStorageImpl) Alive(ctx context.Context, in *empty.Empty) (*pbsystem.AliveResponse, error) {
	return &pbsystem.AliveResponse{Msg: "Hello, I am alive!"}, nil
}
