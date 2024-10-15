package redisx

import (
	"fmt"

	"github.com/0829why/svrframe/config"
	"github.com/0829why/svrframe/logx"

	"github.com/go-redis/redis"
)

var (
	rdb_cluster *redis.ClusterClient
	rdb_client  *redis.Client
)

func init() {
	rdb_cluster = nil
	rdb_client = nil
}

func initRedisCluster(addrs []string, pwd string) error {
	rdb_cluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,
		Password: pwd,
	})

	_, err := rdb_cluster.Ping().Result()
	return err
}

func initRedisSingleton(addr string, pwd string) error {
	rdb_client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       0,
	})

	_, err := rdb_client.Ping().Result()
	return err
}

func InitRedisHelper() error {
	configs := config.GetRedisClusterConfig()
	if configs == nil || configs.Redis == nil || len(configs.Redis) <= 0 {
		return nil
	}

	Addrs := []string{}

	for _, v := range configs.Redis {
		// if net.ParseIP(v.Host) == nil {
		// 	logx.WarnF("redis config , host is err -> %s", v.Host)
		// 	continue
		// }
		if len(v.Host) <= 0 {
			logx.WarnF("redis config , host is nil")
			continue
		}
		if v.Port <= 0 {
			logx.WarnF("redis config , port is 0")
			continue
		}
		addr := fmt.Sprintf("%s:%d", v.Host, v.Port)
		Addrs = append(Addrs, addr)
	}

	if len(Addrs) <= 0 {
		return nil
	}

	err := initRedisCluster(Addrs, configs.Password)
	if err != nil {
		rdb_cluster = nil
		err = initRedisSingleton(Addrs[0], configs.Password)
		if err != nil {
			logx.ErrorF("%v", err)
			return err
		}
	}

	logx.InfoF("InitRedisHelper success")

	return nil
}
