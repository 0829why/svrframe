package mysqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/config"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"

	"github.com/jmoiron/sqlx"
)

type MysqlClient interface {
	GetDatabaseName() string
	Exec(query string, args ...interface{}) (res sql.Result, err error)
	Get(single interface{}, query string, args ...interface{}) (err error, has bool) //查询单行
	Select(arr interface{}, query string, args ...interface{}) (err error)           //查询多行
	Delete(query string, args ...interface{}) (rowsAffected int64, err error)
	Update(query string, args ...interface{}) (rowsAffected int64, err error)
	Insert(query string, args ...interface{}) (lastInsertID int64, err error)
	Query(query string, args ...interface{}) (results []*result, err error)

	//以orm方式更新
	UpdateORM(table_name string, s interface{}) (rowsAffected int64, err error)
	//以orm方式插入
	InsertORM(table_name string, s interface{}) (lastInsertID int64, err error)
}

type mysqlClient struct {
	*config.MysqlConfig
	mysqlDB *sqlx.DB
}

func makeMysql(cfg *config.MysqlConfig) (cli *mysqlClient, err error) {
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", cfg.UserName, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)

	cli = &mysqlClient{
		MysqlConfig: cfg,
	}
	cli.mysqlDB, err = sqlx.Open("mysql", dbDSN)
	if err != nil {
		return
	}
	cli.mysqlDB.SetMaxOpenConns(mysql_max_open_conns)
	cli.mysqlDB.SetMaxIdleConns(mysql_max_idle_conns)
	cli.mysqlDB.SetConnMaxLifetime(mysql_max_lifttime)

	if err = cli.mysqlDB.Ping(); err != nil {
		return
	}

	logx.InfoF("mysql conn success -> %s", dbDSN)
	return
}

func (m *mysqlClient) _update_or_delete(query string, args ...interface{}) (rowsAffected int64, err error) {
	var ret sql.Result
	ret, err = m.mysqlDB.Exec(query, args...)
	if err != nil {
		logx.ErrorF("%v", err)
		return
	}
	rowsAffected, err = ret.RowsAffected()
	if err != nil {
		logx.ErrorF("%v", err)
		return
	}
	return
}

func (m *mysqlClient) GetDatabaseName() string {
	return m.Database
}

func (m *mysqlClient) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	res, err = m.mysqlDB.Exec(query, args...)
	return
}

func (m *mysqlClient) Get(single interface{}, query string, args ...interface{}) (err error, has bool) {
	err = m.mysqlDB.Get(single, query, args...)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			err = nil
			has = false
			return
		}
		logx.ErrorF("%v", err)
		return
	}
	has = true
	return
}

func (m *mysqlClient) Select(arr interface{}, query string, args ...interface{}) (err error) {
	err = m.mysqlDB.Select(arr, query, args...)
	if err != nil {
		logx.ErrorF("%v", err)
		return
	}
	return
}

func (m *mysqlClient) Delete(query string, args ...interface{}) (rowsAffected int64, err error) {
	s := strings.ToLower(query[0:6])
	if strings.Compare(s, "delete") != 0 {
		err = fmt.Errorf("this function just run delete")
		logx.ErrorF("%v", err)
		return
	}
	return m._update_or_delete(query, args...)
}
func (m *mysqlClient) Update(query string, args ...interface{}) (rowsAffected int64, err error) {
	s := strings.ToLower(query[0:6])
	if strings.Compare(s, "update") != 0 {
		err = fmt.Errorf("this function just run update")
		logx.ErrorF("%v", err)
		return
	}
	return m._update_or_delete(query, args...)
}

func (m *mysqlClient) Insert(query string, args ...interface{}) (lastInsertID int64, err error) {
	s := strings.ToLower(query[0:6])
	if strings.Compare(s, "insert") != 0 {
		err = fmt.Errorf("this function just run insert")
		logx.ErrorF("%v", err)
		return
	}
	var ret sql.Result
	ret, err = m.mysqlDB.Exec(query, args...)
	if err != nil {
		logx.ErrorF("%v", err)
		return
	}
	lastInsertID, err = ret.LastInsertId()
	if err != nil {
		logx.ErrorF("%v", err)
		return
	}
	return
}

func (m *mysqlClient) Query(query string, args ...interface{}) (results []*result, err error) {
	// s := strings.ToLower(query[0:6])
	// if strings.Compare(s, "select") != 0 {
	// 	s = strings.ToLower(query[0:4])
	// 	if strings.Compare(s, "call") != 0 {
	// 		err = fmt.Errorf("this function just run select or call")
	// 		logx.ErrorF("%v", err)
	// 		return
	// 	}
	// }

	results = []*result{}
	var rows *sql.Rows
	//查询数据，取所有字段
	rows, err = m.mysqlDB.Query(query, args...)
	if err != nil {
		return
	}

	defer rows.Close()

	for {
		res := &result{}
		//返回所有列
		res.Fileds, err = rows.Columns()
		if err != nil {
			return
		}
		//这里表示一行所有列的值，用[]byte表示
		vals := make([][]byte, len(res.Fileds))
		//这里表示一行填充数据
		scans := make([]interface{}, len(res.Fileds))
		//这里scans引用vals，把数据填充到[]byte里
		for k := range vals {
			scans[k] = &vals[k]
		}
		i := 0
		for rows.Next() {
			//填充数据
			rows.Scan(scans...)
			//每行数据
			r := row{
				Values: map[string]string{},
			}
			//把vals中的数据复制到row中
			for k, v := range vals {
				key := res.Fileds[k]
				//这里把[]byte数据转成string
				r.Values[key] = string(v)
			}
			//放入结果集
			res.Rows = append(res.Rows, r)
			i++
		}

		results = append(results, res)
		if !rows.NextResultSet() {
			break
		}
	}

	if err != nil {
		logx.ErrorF("%v", err)
	}
	return
}

// ////////////////////////////////////////////////////////////////////
func toSqlUpdate(val reflect.Value, table_name string) string {
	typ := val.Type()

	kd := val.Kind()
	if kd != reflect.Struct {
		logx.ErrorF("toSqlUpdate must struct")
		return ""
	}

	cond := ""
	update := ""

	f := func(key string, field reflect.Value) string {
		k := field.Kind()
		switch k {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Sprintf("%s = %v", key, field)
		case reflect.String:
			return fmt.Sprintf("%s = '%v'", key, field)
		default:
			logx.ErrorF("不受支持的类型 => %v", k)
		}

		return ""
	}

	num := val.NumField()
	for i := 0; i < num; i++ {
		t_field := typ.Field(i)
		tags := t_field.Tag
		tag_db := tags.Get("db")
		if len(tag_db) <= 0 {
			continue
		}
		field := val.Field(i)
		primary := tags.Get("primary")
		if len(primary) > 0 {
			cond = f(tag_db, field)
		} else {
			s := f(tag_db, field)
			if len(s) <= 0 {
				continue
			}
			if len(update) > 0 {
				update += ", "
			}
			update += s
		}
	}
	if len(cond) <= 0 || len(update) <= 0 {
		return ""
	}
	sql := fmt.Sprintf("update %s set %s where %s limit 1", table_name, update, cond)
	logx.DebugF("sql = %s", sql)
	return sql
}
func (m *mysqlClient) UpdateORM(table_name string, s interface{}) (rowsAffected int64, err error) {
	val := reflect.ValueOf(s)
	kd := val.Kind()
	var sql string = ""
	if kd == reflect.Pointer {
		sql = toSqlUpdate(val.Elem(), table_name)
	} else {
		sql = toSqlUpdate(val, table_name)
	}
	if len(sql) <= 0 {
		return 0, errors.New("UpdateORM failed")
	}

	return m.Update(sql)
}

func toSqlInsert(val reflect.Value, table_name string) string {
	typ := val.Type()

	kd := val.Kind()
	if kd != reflect.Struct {
		logx.ErrorF("toSqlInsert must struct")
		return ""
	}

	table := ""
	value := ""

	f := func(field reflect.Value) string {
		k := field.Kind()
		switch k {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Sprintf("%v", field)
		case reflect.String:
			return fmt.Sprintf("'%v'", field)
		default:
			logx.ErrorF("不受支持的类型 => %v", k)
		}

		return ""
	}

	num := val.NumField()
	for i := 0; i < num; i++ {
		t_field := typ.Field(i)
		tags := t_field.Tag
		tag_db := tags.Get("db")
		if len(tag_db) <= 0 {
			continue
		}
		field := val.Field(i)
		fvalue := f(field)
		if len(fvalue) <= 0 {
			continue
		}
		if len(table) > 0 {
			table += ", "
			value += ", "
		}
		table += tag_db
		value += fvalue
	}
	if len(table) <= 0 {
		return ""
	}
	sql := fmt.Sprintf("insert into %s (%s) values (%s)", table_name, table, value)
	logx.DebugF("sql = %s", sql)
	return sql
}
func (m *mysqlClient) InsertORM(table_name string, s interface{}) (lastInsertID int64, err error) {
	val := reflect.ValueOf(s)
	kd := val.Kind()
	var sql string = ""
	if kd == reflect.Pointer {
		sql = toSqlInsert(val.Elem(), table_name)
	} else {
		sql = toSqlInsert(val, table_name)
	}
	if len(sql) <= 0 {
		return 0, errors.New("InsertORM failed")
	}

	return m.Insert(sql)
}
