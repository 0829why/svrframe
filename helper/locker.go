package helper

import (
	"sync"
	"time"
)

const (
	// 默认超时的毫秒数(1小时)
	con_Default_Timeout_Milliseconds = 60 * 60 * 1000

	// 写锁每次休眠的时间比读锁的更短，这样是因为写锁有更高的优先级，所以尝试的频率更大
	// 写锁每次休眠的毫秒数
	con_Lock_Sleep_Millisecond = 1

	// 读锁每次休眠的毫秒数
	con_RLock_Sleep_Millisecond = 2
)

// 获取超时时间
func getTimeout(timeout int) int {
	if timeout > 0 {
		return timeout
	} else {
		return con_Default_Timeout_Milliseconds
	}
}

// 写锁对象
type Locker struct {
	write int
	mutex sync.Mutex
}

// 内部锁
// 返回值：
// 加锁是否成功
func (l *Locker) lock() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 如果已经被锁定，则返回失败
	if l.write == 1 {
		return false
	}

	// 否则，将写锁数量设置为１，并返回成功
	l.write = 1

	return true
}

// 尝试加锁，如果在指定的时间内失败，则会返回失败；否则返回成功
// timeout:指定的毫秒数,timeout<=0则将会死等
// 返回值：
// 成功或失败
// 如果失败，返回上一次成功加锁时的堆栈信息
// 如果失败，返回当前的堆栈信息
func (l *Locker) Lock(timeout int) (successful bool) {
	timeout = getTimeout(timeout)

	// 遍历指定的次数（即指定的超时时间）
	for i := 0; i < timeout; i = i + con_Lock_Sleep_Millisecond {
		// 如果锁定成功，则返回成功
		if l.lock() {
			successful = true
			break
		}

		// 如果锁定失败，则休眠con_Lock_Sleep_Millisecond ms，然后再重试
		time.Sleep(con_Lock_Sleep_Millisecond * time.Millisecond)
	}

	return
}

// 解锁
func (l *Locker) Unlock() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.write = 0
}

// 创建新的锁对象
func NewLocker() *Locker {
	return &Locker{}
}
