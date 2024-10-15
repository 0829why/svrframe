package util_rpc

import (
	"time"

	"github.com/0829why/svrframe/constants"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	grpcServer *grpc.Server
)

func init() {
	var options = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(constants.RpcSendRecvMaxSize),
		grpc.MaxSendMsgSize(constants.RpcSendRecvMaxSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             time.Second * 5,
			PermitWithoutStream: true,
		}),
	}

	grpcServer = grpc.NewServer(options...)
}
