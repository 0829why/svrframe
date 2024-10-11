package redisx

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/constants"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"

	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
)

func GetRedis() redis.Cmdable {
	var cmdable redis.Cmdable = nil
	if rdb_cluster != nil {
		cmdable = redis.Cmdable(rdb_cluster)
	} else if rdb_client != nil {
		cmdable = redis.Cmdable(rdb_client)
	}
	return cmdable
}

func Pipelined(fn func(redis.Pipeliner) error) error {
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return nil
	}
	_, err := cmder.Pipelined(fn)
	return err
}

func ExecLua(script string, keys []string, args ...interface{}) (result interface{}) {
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return nil
	}
	new_keys := []string{}
	for _, v := range keys {
		key := constants.ProjectName + ":" + v
		new_keys = append(new_keys, key)
	}
	var luaScript = redis.NewScript(script)
	cmd := luaScript.Run(cmder, new_keys, args...)
	r, err := cmd.Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("ExecLua, err = %v", err)
		return nil
	}
	return r
}

func ExecLuaString(script string, keys []string, args ...interface{}) string {
	r := ExecLua(script, keys, args...)
	if r == nil {
		return ""
	}
	res, ok := r.(string)
	if !ok {
		return ""
	}
	return res
}

func ExecLuaStringArray(script string, keys []string, args ...interface{}) []string {
	r := ExecLua(script, keys, args...)
	if r == nil {
		return nil
	}
	i_res, ok := r.([]interface{})
	if !ok {
		return nil
	}
	res := []string{}
	for _, v := range i_res {
		s, ok := v.(string)
		if ok {
			res = append(res, s)
		} else {
			res = append(res, "")
		}
	}

	return res
}

func ExecLuaInt[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](script string, keys []string, args ...interface{}) T {
	r := ExecLua(script, keys, args...)
	if r == nil {
		return 0
	}
	res, ok := r.(int64)
	if !ok {
		return 0
	}
	return T(res)
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

func AutoLock(key string, fn func()) {
	key = constants.ProjectName + ":locker:" + key
	c := make(chan int)
	expt := time.Second * 30
	locker_value := uuid.NewV4()
	locker_value = uuid.NewV5(locker_value, key)
	go func() {
		for {
			res := SetNX(key, locker_value.String(), expt)
			if res {
				c <- 1
				break
			}
			time.Sleep(time.Millisecond * 100) //100毫秒尝试一次
		}
	}()
	//等待获取锁
	<-c

	defer func() {
		val := Get(key)
		if val == locker_value.String() {
			//查看,如果锁还是自己的,那就释放锁
			Del(key)
		}
	}()
	//执行逻辑
	fn()
}

func Exists(key string) bool {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return false
	}
	i, err := cmder.Exists(key).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("Exists %s, err = %v", key, err)
		return false
	}

	if i > 0 {
		return true
	}

	return false
}

func Expire(key string, expiration time.Duration) bool {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return false
	}
	b, err := cmder.Expire(key, expiration).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("Expire %s, err = %v", key, err)
		return false
	}
	return b
}

func Del(keys ...string) {
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return
	}
	real_keys := []string{}
	for _, s := range keys {
		real_keys = append(real_keys, constants.ProjectName+":"+s)
	}
	cmder.Del(real_keys...)
}

func SetNX(key string, val interface{}, expiration ...time.Duration) bool {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return false
	}
	exp := time.Second * -1
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	status, err := cmder.SetNX(key, val, exp).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("Set %s, err = %v", key, err)
		return false
	}
	return status
}

func SetAndExp(key string, val interface{}, expiration time.Duration) (status string, err error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err = fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	status, err = cmder.Set(key, val, expiration).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("SetAndExp %s, err = %v", key, err)
	}
	return
}
func Set(key string, val interface{}) (status string, err error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err = fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	status, err = cmder.Set(key, val, time.Second*-1).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("Set %s, err = %v", key, err)
	}
	return
}
func Get(key string) (val string) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	val, err := cmder.Get(key).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("Get %s, err = %v", key, err)
	}
	if err == redis.Nil {
		val = ""
	}
	return
}
func GetI[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](key string) (ival T) {
	val := Get(key)

	ival = 0
	if len(val) > 0 {
		tval, _ := strconv.Atoi(val)
		ival = T(tval)
	}
	return
}
func IncryBy[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](key string, value T) (after T, err error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err = fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	val, err := cmder.IncrBy(key, int64(value)).Result()
	if err != nil {
		logx.ErrorF("%v", err)
		return
	}
	after = T(val)
	return
}
func SetStruct(key string, stc interface{}) (status string, err error) {
	j, err := json.Marshal(stc)
	if err != nil && err != redis.Nil {
		logx.ErrorF("%v", err)
		return
	}
	return Set(key, j)
}
func HSet(key, field string, value interface{}) (status bool, err error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err = fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	status, err = cmder.HSet(key, field, value).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("HSet %s, field %s, err = %v", key, field, err)
	}
	return
}
func HGet(key, field string) (val string) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	val, err := cmder.HGet(key, field).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("HGet %s, field %s, err = %v", key, field, err)
	}
	if err == redis.Nil {
		val = ""
	}
	return
}
func HGetI[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](key, field string) (ival T) {
	val := HGet(key, field)

	ival = 0
	if len(val) > 0 {
		tval, _ := strconv.Atoi(val)
		ival = T(tval)
	}
	return
}
func HIncryBy[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](key, field string, value T) (after T, err error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err = fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	val, err := cmder.HIncrBy(key, field, int64(value)).Result()
	if err != nil {
		logx.ErrorF("%v", err)
		return
	}
	after = T(val)
	return
}
func HDel(key string, field ...string) (err error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err = fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	_, err = cmder.HDel(key, field...).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("HDel %s, field %v, err = %v", key, field, err)
	}
	return
}

// -1代表出错
func HLen(key string) int64 {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return -1
	}
	len, err := cmder.HLen(key).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("HLen %s, err = %v", key, err)
		return -1
	}
	return len
}

func HMSet(key string, fields map[string]interface{}) (status string, err error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err = fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	status, err = cmder.HMSet(key, fields).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("HMSet %s, err = %v", key, err)
	}
	return
}
func HMGet(key string, fields []string) (result map[string]interface{}, err error) {
	key = constants.ProjectName + ":" + key

	result = map[string]interface{}{}

	cmder := GetRedis()
	if cmder == nil {
		err = fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	if len(fields) <= 0 {
		return
	}
	var res []interface{}
	res, err = cmder.HMGet(key, fields...).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("HMGet %s, err = %v", key, err)
		return
	}
	if err == nil && len(res) > 0 {
		for idx, field := range fields {
			if res[idx] != nil {
				result[field] = res[idx]
			}
		}
	}
	return
}

func HGetAll(key string) (result map[string]string, err error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err = fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return
	}
	result, err = cmder.HGetAll(key).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("HGetAll %s, err = %v", key, err)
	}
	return
}

func LLen(key string, val interface{}) int64 {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return 0
	}
	result, err := cmder.LLen(key).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("LPush %s, err = %v", key, err)
		return 0
	}
	return result
}

func LPush(key string, val ...interface{}) int64 {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return 0
	}
	result, err := cmder.LPush(key, val...).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("LPush %s, err = %v", key, err)
		return 0
	}
	return result
}

func LPop(key string) string {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return ""
	}
	result, err := cmder.LPop(key).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("LPop %s, err = %v", key, err)
		return ""
	}
	return result
}

func RPush(key string, val ...interface{}) int64 {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return 0
	}
	result, err := cmder.RPush(key, val...).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("RPush %s, err = %v", key, err)
		return 0
	}
	return result
}

func RPop(key string) string {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return ""
	}
	result, err := cmder.RPop(key).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("RPop %s, err = %v", key, err)
		return ""
	}
	return result
}

func LRange(key string, start int64, stop int64) []string {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		err := fmt.Errorf("no redis instance")
		logx.ErrorF("%v", err)
		return nil
	}
	result, err := cmder.LRange(key, start, stop).Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("LRange %s, err = %v", key, err)
		return nil
	}
	return result
}

func ZAdd(key string, increment float64, member string) (int64, error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return 0, fmt.Errorf("no redis instance")
	}

	z := redis.Z{
		Score:  increment,
		Member: member,
	}

	val, err := cmder.ZAdd(key, z).Result()
	if err != nil {
		logx.ErrorF("ZAdd err -> %v", err)
		return 0, err
	}

	return val, err
}

func ZIncrBy(key string, increment float64, member string) (float64, error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return 0, fmt.Errorf("no redis instance")
	}

	val, err := cmder.ZIncrBy(key, increment, member).Result()
	if err != nil {
		logx.ErrorF("ZIncrBy err -> %v", err)
		return 0, err
	}

	return val, err
}

func ZRevRank(key string, member string) (int64, error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return 0, fmt.Errorf("no redis instance")
	}

	val, err := cmder.ZRevRank(key, member).Result()
	if err != nil {
		logx.ErrorF("ZRevRank err -> %v", err)
		return 0, err
	}

	return val, err
}

func ZScore(key string, member string) (float64, error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return 0, fmt.Errorf("no redis instance")
	}

	val, err := cmder.ZScore(key, member).Result()
	if err != nil {
		logx.ErrorF("ZScore err -> %v", err)
		return 0, err
	}

	return val, err
}

func ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	key = constants.ProjectName + ":" + key
	cmder := GetRedis()
	if cmder == nil {
		logx.ErrorF("no redis instance")
		return nil, fmt.Errorf("no redis instance")
	}

	zList, err := cmder.ZRevRangeWithScores(key, start, stop).Result()
	if err != nil {
		logx.ErrorF("ZRevRange err -> %v", err)
		return nil, err
	}

	return zList, err
}
