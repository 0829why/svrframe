package util_rpc

import (
	"net"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/constants"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"

	"google.golang.org/grpc"
)

func update() {
	go func() {
		exit_ch := constants.GetServiceStopListener().AddListener()
		constants.GetServiceStopWaitGroup().Add(1)
		<-exit_ch.Done()
		grpcServer.Stop()
		grpcServer = nil
		constants.GetServiceStopWaitGroup().Done()
	}()
}

func StartRpcService(regfunc func(server *grpc.Server)) (listenPort uint16) {
	listenPort = 0
	lis, err := net.Listen("tcp4", "0.0.0.0:0")
	if err != nil {
		logx.ErrorF("failed to listen: %v", err)
		return
	}

	addr, err := net.ResolveTCPAddr(lis.Addr().Network(), lis.Addr().String())
	if err != nil {
		logx.ErrorF("%v", err)
		return
	}
	listenPort = uint16(addr.Port)

	logx.InfoF("RPC SERVER [ %s ] RUNNING", constants.Service_Type)

	regfunc(grpcServer)

	go grpcServer.Serve(lis)

	update()

	return
}
