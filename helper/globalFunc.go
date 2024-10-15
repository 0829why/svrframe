package helper

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sync"
	"unicode"

	"github.com/0829why/svrframe/constants"
	"github.com/0829why/svrframe/logx"

	"github.com/mitchellh/mapstructure"
)

func GetHash32s(s string) uint32 {
	return GetHash32([]byte(s))
}
func GetHash32(b []byte) uint32 {
	h32 := fnv.New32()
	h32.Write(b)
	return h32.Sum32()
}
func GetHash64s(s string) uint64 {
	return GetHash64([]byte(s))
}
func GetHash64(b []byte) uint64 {
	h64 := fnv.New64()
	h64.Write(b)
	return h64.Sum64()
}

func ToJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func FromJson(j string, out interface{}) error {
	if len(j) <= 0 {
		return fmt.Errorf("FromJson, data empty")
	}
	err := json.Unmarshal([]byte(j), out)
	return err
}

func ConvertInterface2Struct(src interface{}, out interface{}) error {
	s := ToJson(src)
	return FromJson(s, out)
}

func GetCharactorCount(val string) (count int) {
	b := []rune(val)
	count = len(b)
	return
}
func GetChineseCharactorCount(val string) (count int) {
	for _, char := range val {
		if unicode.Is(unicode.Han, char) {
			count++
		} else {
			return 0
		}
	}
	return
}

func Struct2Map(input interface{}) (out map[string]interface{}, err error) {
	out = map[string]interface{}{}
	err = mapstructure.Decode(input, &out)
	if err != nil {
		logx.ErrorF("Struct2Map err -> %v", err)
	}
	return
}

func AutoLock(lc *sync.Mutex) func() {
	if lc == nil {
		return func() {}
	}
	lc.Lock()
	return func() {
		lc.Unlock()
	}
}

// 数值转万分比小数
func GetTenThousandthRatio(num float64) float64 {
	return num * constants.TenThousandthRatio
}

func MakeUInt64(hi, lo uint32) uint64 {
	n := uint64(hi) << 32
	n += uint64(lo)
	return n
}
func ParseUInt64(n uint64) (hi uint32, lo uint32) {
	lo = uint32(n & 0xFFFFFFFF)
	hi = uint32(n >> 32)
	return
}
func MakeUInt32(hi, lo uint16) uint32 {
	n := uint32(hi) << 16
	n += uint32(lo)
	return n
}
func ParseUInt32(n uint32) (hi uint16, lo uint16) {
	lo = uint16(n & 0xFFFF)
	hi = uint16(n >> 16)
	return
}

// //////////////////////////////////////////////////////////////
func JsonArrayInterface2ArrayInt[T int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | float32 | float64](i []interface{}) [][]T {
	if len(i) <= 0 {
		return [][]T{}
	}
	t := [][]T{}
	err := json.Unmarshal([]byte(ToJson(i)), &t)
	if err != nil {
		return t
	}
	return t
}
