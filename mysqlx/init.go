package mysqlx

import (
	"database/sql"
	"fmt"
	"time"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/config"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"

	_ "github.com/go-sql-driver/mysql"
)

const (
	//最大连接数
	mysql_max_open_conns = 100
	//闲置连接数
	mysql_max_idle_conns = 20
	//最大连接周期
	mysql_max_lifttime = 60 * time.Second
)

var (
	mysqls map[string]*mysqlClient
)

func init() {
	mysqls = make(map[string]*mysqlClient)
}

func InitMysqlHelper() error {
	configs := config.GetMysqlConfigs()
	if configs == nil || len(configs) <= 0 {
		return nil
	}

	for _, cfg := range configs {
		cli, err := makeMysql(cfg)
		if err != nil {
			logx.ErrorF("makeMysql err -> %v", err)
			return err
		}
		mysqls[cfg.Name] = cli
		logx.InfoF("mysql helper [ %s ] create success", cfg.Name)
	}

	logx.InfoF("InitMysqlHelper success")
	return nil
}

func GetMysqlClient(dbname string) MysqlClient {
	cli, ok := mysqls[dbname]
	if !ok {
		return nil
	}
	return cli
}

func GetMysqlDatebase(dbname string) string {
	cli, ok := mysqls[dbname]
	if !ok {
		return ""
	}
	return cli.Database
}

func TableExists(dbname string, table_name string) bool {
	m := GetMysqlClient(dbname)
	if m == nil {
		err := fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("TableExists err -> %v", err)
		return false
	}
	var count int32 = 0
	sql := fmt.Sprintf(`select count(*) from information_schema.tables where TABLE_SCHEMA = '%s' and TABLE_NAME = '%s' limit 1`, m.GetDatabaseName(), table_name)
	err, has := m.Get(&count, sql)
	if err != nil {
		logx.ErrorF("TableExists err -> %v", err)
		return false
	}
	if !has {
		return false
	}

	return count > 0
}

func Exec(dbname string, query string, args ...interface{}) (res sql.Result, err error) {
	m := GetMysqlClient(dbname)
	if m == nil {
		err = fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("Get err -> %v", err)
		return
	}
	return m.Exec(query, args...)
}

func Get(dbname string, single interface{}, query string, args ...interface{}) (err error, has bool) {
	m := GetMysqlClient(dbname)
	if m == nil {
		err = fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("Get err -> %v", err)
		return
	}
	return m.Get(single, query, args...)
}

func Select(dbname string, arr interface{}, query string, args ...interface{}) (err error) {
	m := GetMysqlClient(dbname)
	if m == nil {
		err = fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("Select err -> %v", err)
		return
	}
	return m.Select(arr, query, args...)
}

func Delete(dbname string, query string, args ...interface{}) (rowsAffected int64, err error) {
	m := GetMysqlClient(dbname)
	if m == nil {
		err = fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("Delete err -> %v", err)
		return
	}
	return m.Delete(query, args...)
}
func Update(dbname string, query string, args ...interface{}) (rowsAffected int64, err error) {
	m := GetMysqlClient(dbname)
	if m == nil {
		err = fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("Update err -> %v", err)
		return
	}
	return m.Update(query, args...)
}

func Insert(dbname string, query string, args ...interface{}) (lastInsertID int64, err error) {
	m := GetMysqlClient(dbname)
	if m == nil {
		err = fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("Insert err -> %v", err)
		return
	}
	return m.Insert(query, args...)
}

func Query(dbname string, query string, args ...interface{}) (results []*result, err error) {
	m := GetMysqlClient(dbname)
	if m == nil {
		err = fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("Query err -> %v", err)
		return
	}
	return m.Query(query, args...)
}

// 以orm方式更新
func UpdateORM(dbname string, table_name string, s interface{}) (rowsAffected int64, err error) {
	m := GetMysqlClient(dbname)
	if m == nil {
		err = fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("Update err -> %v", err)
		return
	}
	return m.UpdateORM(table_name, s)
}

// 以orm方式插入
func InsertORM(dbname string, table_name string, s interface{}) (lastInsertID int64, err error) {
	m := GetMysqlClient(dbname)
	if m == nil {
		err = fmt.Errorf("not found mysql client -> name %s", dbname)
		logx.ErrorF("Update err -> %v", err)
		return
	}
	return m.InsertORM(table_name, s)
}
