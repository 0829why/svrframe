package etcd

import (
	"sync"

	gresolver "google.golang.org/grpc/resolver"
)

var once sync.Once
var etcdResolver gresolver.Builder

func init() {
	once = sync.Once{}
	etcdResolver = nil
}
