package etcd

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	gresolver "google.golang.org/grpc/resolver"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/config"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/constants"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"
)

func getResolver() gresolver.Builder {
	once.Do(func() {
		etcdCfg := config.GetEtcdInfo()
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   etcdCfg.EtcdCenters,
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			logx.ErrorF("%+v", err)
			return
		}
		etcdResolver, err = resolver.NewBuilder(cli)
		if err != nil {
			log.Fatalf("CreateResolver error %+v", err)
			return
		}
	})
	return etcdResolver
}

func CreateEtcdConnect(name string) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(getResolver()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(constants.RpcSendRecvMaxSize), grpc.MaxCallSendMsgSize(constants.RpcSendRecvMaxSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             100 * time.Millisecond,
			PermitWithoutStream: true}),
	}
	full_name := fmt.Sprintf("etcd:///%s/%s", constants.ProjectName, name)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	conn, err = grpc.DialContext(ctx, full_name, opts...)
	if err != nil {
		logx.ErrorF("CreateEtcdConnect err => %+v", err)
		return
	}
	return
}
