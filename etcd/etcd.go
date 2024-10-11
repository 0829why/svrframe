package etcd

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/config"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/constants"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/helper"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"
)

const ttl = 10

var service *Service

// 服务元数据信息
type MetaData struct {
	ServiceType string
	Name        string
	Addr        string
	Port        uint16
	Ext         interface{} //扩展数据
}

// 命名服务结构体
type Service struct {
	md            *MetaData
	Client        *clientv3.Client
	full_name     string
	manager       endpoints.Manager
	lease         *clientv3.LeaseGrantResponse
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

func GetEtcdName() string {
	if service == nil {
		return ""
	}
	return service.md.Name
}

// 创建一个新的命名服务
func StartNamingService(rpcListenPort uint16, custom_name_func func(lease_id int64) string, ext interface{}) error {
	etcdCfg := config.GetEtcdInfo()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdCfg.EtcdCenters,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logx.ErrorF("%+v", err)
		return err
	}

	target := fmt.Sprintf("%s/%s", constants.ProjectName, constants.Service_Type)
	// etcd的endpoints管理
	manager, err := endpoints.NewManager(cli, target)
	if err != nil {
		logx.ErrorF("%+v", err)
		return err
	}

	// 在etcd创建一个续期的lease对象
	lease, err := cli.Grant(context.Background(), ttl)
	if err != nil {
		logx.ErrorF("%+v", err)
		return err
	}

	name := fmt.Sprintf("%s/%d", constants.Service_Type, lease.ID)
	if custom_name_func != nil {
		name = custom_name_func(int64(lease.ID)) //fmt.Sprintf("%s/%s", constants.Service_Type, name)
	}
	md := &MetaData{
		ServiceType: constants.Service_Type,
		Name:        name,
		Addr:        constants.ServiceHost,
		Port:        rpcListenPort,
		Ext:         ext,
	}
	ep := endpoints.Endpoint{
		Addr:     fmt.Sprintf("%s:%d", constants.GetServiceHost(), rpcListenPort),
		Metadata: helper.ToJson(md),
	}

	full_name := fmt.Sprintf("%s/%s", constants.ProjectName, name)
	// 向etcd注册一个Endpoint并绑定续期
	err = manager.AddEndpoint(context.Background(), full_name, ep, clientv3.WithLease(lease.ID))
	if err != nil {
		logx.ErrorF("%+v", err)
		return err
	}

	ch, err := cli.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		logx.ErrorF("%+v", err)
		return err
	}

	service = &Service{
		md:            md,
		manager:       manager,
		full_name:     full_name,
		Client:        cli,
		lease:         lease,
		keepAliveChan: ch,
	}

	go service.update()

	return nil
}

func (s *Service) update() {
	exit_ch := constants.GetServiceStopListener().AddListener()
	constants.GetServiceStopWaitGroup().Add(1)

	logx.InfoF("etcd service start success -> leaseID => %d, service => %+v, metadata => %+v", s.lease.ID, s, s.md)

	defer func() {
		s.manager.DeleteEndpoint(context.Background(), s.full_name)
		s.Client.Revoke(context.Background(), s.lease.ID)
		s.Client.Close()

		constants.GetServiceStopWaitGroup().Done()
	}()

	for {
		select {
		case <-exit_ch.Done():
			return
		case success, ok := <-s.keepAliveChan:
			if ok && success != nil {
				//logx.DebugF("续约结果 -> %v", success)
			} else {
				// etcd.startRunner()
			}
		default:
			time.Sleep(time.Second * 3)
		}
	}
}
