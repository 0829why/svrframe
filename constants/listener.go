package constants

type ChanPacking struct {
	ch  chan interface{}
	del bool
}

func (c *ChanPacking) Close() {
	c.del = true
}

func (c *ChanPacking) Done() <-chan interface{} {
	return c.ch
}

type IListenerManager struct {
	listenList []*ChanPacking
}

func (m *IListenerManager) del(idx int) {
	m.listenList = append(m.listenList[:idx], m.listenList[idx+1:]...)
}

func (l *IListenerManager) AddListener() *ChanPacking {
	ch := &ChanPacking{
		ch:  make(chan interface{}, 1),
		del: false,
	}
	l.listenList = append(l.listenList, ch)
	return ch
}
func (l *IListenerManager) NotifyAllListeners(param ...interface{}) {
	for i, c := range l.listenList {
		if !c.del {
			if len(param) > 0 {
				c.ch <- param
			} else {
				c.ch <- true
			}
		} else {
			l.del(i)
		}
	}
}
func (l *IListenerManager) Clear() {
	l.listenList = []*ChanPacking{}
}

func NewListenerManager() *IListenerManager {
	return &IListenerManager{
		listenList: []*ChanPacking{},
	}
}
