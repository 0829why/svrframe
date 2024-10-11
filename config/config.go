package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/constants"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/helper"
)

type ServiceConfig struct {
	ProjectName  string
	Etcd         *EtcdConfig
	Mysql        map[string]*MysqlConfig
	RedisCluster *RedisClusterConfig
	Custom       interface{}
}
type EtcdConfig struct {
	EtcdCenters []string
}
type MysqlConfig struct {
	Name     string
	UserName string
	Password string
	Host     string
	Port     uint16
	Database string
	Charset  string
}
type RedisClusterConfig struct {
	Password string
	Redis    []*RedisConfig
}
type RedisConfig struct {
	Host string
	Port uint16
}

func ParseConfig() (err error) {
	//解析配置文件
	if constants.FlagValue_ConfigFile == nil || len(*constants.FlagValue_ConfigFile) <= 0 {
		err = fmt.Errorf("need cmdline param -config=config file path")
		return
	}

	err = readConfig(*constants.FlagValue_ConfigFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	return
}

func readConfig(config_file string) (err error) {
	dir, _ := os.Getwd()
	dir = strings.ReplaceAll(dir, "\\", "/")
	fullPath := fmt.Sprintf("%s/%s", dir, config_file)

	filePtr, err := os.Open(fullPath)
	if err != nil {
		paths := strings.Split(dir, "/")
		paths = paths[0 : len(paths)-1]
		dir = strings.Join(paths, "/")
		fmt.Printf("not found config %s, try found in gopath\n", fullPath)
		// dir = os.Getenv(constants.Env_Gopath)
		fullPath = fmt.Sprintf("%s/%s", dir, config_file)
		filePtr, err = os.Open(fullPath)
		if err != nil {
			fmt.Printf("Open file failed [Err:%s]\n", err.Error())
			return
		}
	}
	defer filePtr.Close()

	allbytes, err := io.ReadAll(filePtr)
	if err != nil {
		fmt.Printf("Read file failed [Err:%s]\n", err.Error())
		return
	}

	err = json.Unmarshal(allbytes, serviceConfig)
	if err != nil {
		fmt.Printf("Parse file failed [Err:%s]\n", err.Error())
		return
	}
	constants.ProjectName = serviceConfig.ProjectName

	if serviceConfig.Mysql != nil && len(serviceConfig.Mysql) > 0 {
		for name, m := range serviceConfig.Mysql {
			if m.Port == 0 || len(m.Database) <= 0 || len(m.UserName) <= 0 || len(m.Password) <= 0 {
				return fmt.Errorf("mysql config has wrong")
			}
			m.Charset = "utf8mb4"
			m.Name = name
		}
	}

	fmt.Printf("Parse file [%s] success \n", config_file)
	return
}

func GetServiceConfig() *ServiceConfig {
	return serviceConfig
}
func GetMysqlConfigs() map[string]*MysqlConfig {
	return serviceConfig.Mysql
}
func GetMysqlConfig(dbName string) *MysqlConfig {
	dbcfg, ok := serviceConfig.Mysql[dbName]
	if !ok {
		return nil
	}
	return dbcfg
}
func GetRedisClusterConfig() *RedisClusterConfig {
	return serviceConfig.RedisCluster
}
func GetEtcdInfo() *EtcdConfig {
	return serviceConfig.Etcd
}
func GetCustomConfig(out interface{}) {
	if serviceConfig.Custom == nil {
		return
	}
	b := helper.ToJson(serviceConfig.Custom)
	helper.FromJson(b, out)
}
