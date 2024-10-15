package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"github.com/0829why/svrframe/config"
	"github.com/0829why/svrframe/constants"
	"github.com/0829why/svrframe/logx"
)

var watcher *Watcher

type Watcher struct {
	Client     *clientv3.Client
	manager    endpoints.Manager
	watcher_ch <-chan []*endpoints.Update
	callback   func(watch *endpoints.Update)
}

func StartWatcherService(callback func(watch *endpoints.Update)) error {
	etcdCfg := config.GetEtcdInfo()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdCfg.EtcdCenters,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logx.ErrorF("%+v", err)
		return err
	}
	manager, err := endpoints.NewManager(cli, constants.ProjectName+"/")
	if err != nil {
		logx.ErrorF("%+v", err)
		return err
	}

	watcher_ch, err := manager.NewWatchChannel(context.Background())
	if err != nil {
		logx.ErrorF("%+v", err)
		return err
	}

	watcher = &Watcher{
		manager:    manager,
		Client:     cli,
		watcher_ch: watcher_ch,
		callback:   callback,
	}

	go watcher.update()

	return nil
}

func (w *Watcher) update() {
	exit_ch := constants.GetServiceStopListener().AddListener()
	constants.GetServiceStopWaitGroup().Add(1)

	logx.InfoF("etcd watcher start success -> watcher => %+v", w)

	defer func() {
		w.Client.Close()
		constants.GetServiceStopWaitGroup().Done()
	}()

	for {
		select {
		case <-exit_ch.Done():
			return
		case watchs, ok := <-w.watcher_ch:
			if ok && len(watchs) > 0 {
				for _, watch := range watchs {
					logx.DebugF("watch => %+v", watch)
					if w.callback != nil {
						w.callback(watch)
					}
				}
			}
		default:
			time.Sleep(time.Second * 3)
		}
	}
}
