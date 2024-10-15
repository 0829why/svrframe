package helper

import (
	"math/rand"
	"time"

	"github.com/0829why/svrframe/logx"

	filter "github.com/antlinker/go-dirtyfilter"
	"github.com/antlinker/go-dirtyfilter/store"
)

const ()

var (
	globalTimer ITimerManager //全局定时器

	filterManager *filter.DirtyManager
)

func init() {
	globalTimer = NewTimerManager()
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func InitFilterManager(filterFile string) {
	memStore, err := store.NewMemoryStore(store.MemoryConfig{
		DataSource: []string{},
	})
	if err != nil {
		panic(err)
	}
	filterManager = filter.NewDirtyManager(memStore)
}

// 铭感词过滤
func FilterVerify(text string) bool {
	if filterManager == nil {
		return false
	}
	result, err := filterManager.Filter().Filter(text)
	if err != nil {
		logx.ErrorF("FilterVerify err -> %v", err)
		return false
	}
	return result == nil
}

func GetGlobalTimer() ITimerManager {
	return globalTimer
}
