package helper

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/0829why/svrframe/logx"
)

type ITimerItem interface {
	GetID() int64
	GetTimerMilliTime() int64
	IsLoop() bool
	Stop()
	GetLeftTime() int64 //剩余多长时间才执行,毫秒
	GetUData() []interface{}
}

type timerItem struct {
	id        int64
	timer     *time.Timer
	milli     int64 //
	starttime int64 //启动时间
	loop      bool  //是否循环
	delflag   bool  //删除标识
	cb        func(item ITimerItem)
	udata     []interface{}
}

func (m *timerItem) GetID() int64 {
	return m.id
}
func (m *timerItem) GetTimerMilliTime() int64 {
	return m.milli
}
func (m *timerItem) IsLoop() bool {
	return m.loop
}
func (m *timerItem) Stop() {
	if m.delflag {
		return
	}
	m.delflag = true
	m.timer.Stop()
}
func (m *timerItem) GetLeftTime() int64 {
	_now := GetNowTimestampMilli()
	last_lefttime := _now - m.starttime

	return m.milli - last_lefttime
}
func (m *timerItem) GetUData() []interface{} {
	return m.udata
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////
type ITimerManager interface {
	StartTimer(milli int64, loop bool, cb func(item ITimerItem), udata ...interface{}) ITimerItem
	ClearAll()
	Update() bool
}

type timerManager struct {
	id_genc atomic.Int64
	channel chan *timerItem
	mList   sync.Map //id->*timerItem
}

func (m *timerManager) StartTimer(milli int64, loop bool, cb func(item ITimerItem), udata ...interface{}) ITimerItem {
	if milli <= 0 {
		logx.ErrorF("定时器时差必须大于0毫秒")
		return nil
	}
	t := &timerItem{
		id:      m.id_genc.Add(1),
		milli:   milli,
		loop:    loop,
		delflag: false,
		cb:      cb,
		udata:   append([]interface{}{}, udata...),
	}

	logx.DebugF("开启定时器,id = %d", t.id)
	m.add(t)

	return t
}
func (m *timerManager) ClearAll() {
	m.mList.Range(func(key, value any) bool {
		item := value.(*timerItem)
		item.Stop()
		return true
	})
}

func (m *timerManager) add(t *timerItem) {
	d := time.Millisecond * time.Duration(t.milli)
	t.timer = time.NewTimer(d)
	t.starttime = GetNowTimestampMilli()

	m.channel <- t
}

func (m *timerManager) Update() bool {
	select {
	case t, ok := <-m.channel:
		if ok && t != nil {
			m.mList.Store(t.id, t)
		}
	default:
		break
	}
	m.mList.Range(func(key, value any) bool {
		id := key.(int64)
		item := value.(*timerItem)
		if item.delflag {
			m.mList.Delete(id)
		} else {
			select {
			case <-item.timer.C:
				if !item.delflag {
					item.cb(item)
					if item.loop {
						d := time.Millisecond * time.Duration(item.milli)
						item.timer.Reset(d)
					} else {
						item.Stop()
					}
				}
			default:
			}
		}
		return true
	})

	return true
}

func NewTimerManager() ITimerManager {
	t := &timerManager{
		id_genc: atomic.Int64{},
		channel: make(chan *timerItem, 10),
		mList:   sync.Map{},
	}
	return t
}
