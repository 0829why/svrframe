package redisx_test

import (
	"fmt"
	"testing"

	"github.com/go-redis/redis"
)

func ExecLua(script string, keys []string, args ...interface{}) (result interface{}) {
	rdb_client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	var luaScript = redis.NewScript(script)
	r, err := luaScript.Run(rdb_client, keys, args...).Result()
	if err != nil && err != redis.Nil {
		return nil
	}
	return r
}

func ExecLuaIntArray[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](script string, keys []string, args ...interface{}) []T {
	r := ExecLua(script, keys, args...)
	if r == nil {
		return nil
	}
	i_res, ok := r.([]interface{})
	if !ok {
		return nil
	}
	res := []T{}
	for _, v := range i_res {
		i, ok := v.(int64)
		if ok {
			res = append(res, T(i))
		} else {
			res = append(res, 0)
		}
	}

	return res
}

func TestScriptGet(t *testing.T) {
	lua := `
		local res1 = redis.call("get", "test_key1")
		local res2 = redis.call("get", "test_key2")
		local res3 = redis.call("get", "test_key3")
		local res = {}
		table.insert(res, (res1))
		table.insert(res, (res2))
		table.insert(res, (res3))
		return res
	`
	res := ExecLuaIntArray[int](lua, nil)
	fmt.Println(res)
}
