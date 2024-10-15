package znet

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"github.com/0829why/svrframe/constants"
	"github.com/0829why/svrframe/helper"
	"github.com/0829why/svrframe/logx"
	"github.com/0829why/svrframe/tcp/utils"
	"github.com/0829why/svrframe/tcp/zcommon"
	"github.com/0829why/svrframe/tcp/ziface"

	"golang.org/x/time/rate"
)

// Connection 链接
type Connection struct {
	//当前Conn属于哪个Server
	TCPServer ziface.IServer
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//消息管理MsgID和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle
	//告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte

	sync.RWMutex
	//链接属性
	//property map[string]interface{}
	property sync.Map
	////保护当前property的锁
	//propertyLock sync.Mutex
	//当前连接的关闭状态
	isClosed bool

	//读package
	rdpkg *zcommon.ReadPackage

	//心跳
	keepalive int
	//限流器
	limiter          *rate.Limiter
	limitFailedCount int
	valid            bool
}

// NewConnection 创建连接的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	//初始化Conn属性
	limit := rate.Every(time.Millisecond * 200)
	c := &Connection{
		TCPServer:        server,
		Conn:             conn,
		ConnID:           connID,
		isClosed:         false,
		MsgHandler:       msgHandler,
		msgBuffChan:      make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:         sync.Map{},
		rdpkg:            zcommon.NewReadPackage(),
		limiter:          rate.NewLimiter(limit, zcommon.Limiter_bucket),
		limitFailedCount: 0,
		valid:            false,
	}

	//将新创建的Conn添加到链接管理中
	c.TCPServer.GetConnMgr().Add(c)
	return c
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) StartWriter() {
	logx.Debugln("[Writer Goroutine is running]")
	defer logx.Debugln(c.ClientIP(), "[conn Writer exit!]")

	for {
		select {
		case <-c.ctx.Done():
			return
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					logx.Debugln("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				logx.Debugln("msgBuffChan is Closed")
				return
			}
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}

// StartReader 读消息Goroutine，用于从客户端中读取数据
func (c *Connection) StartReader() {
	logx.Debugln("[Reader Goroutine is running]")
	defer logx.Debugln(c.ClientIP(), "[conn Reader exit!]")
	defer c.Stop()

	// 创建拆包解包的对象
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			err := c.rdpkg.ReadFromConn(c.Conn)
			if err != nil {
				logx.ErrorF("%v", err)
				return
			}
			if c.rdpkg.Success() {
				func() {
					defer c.rdpkg.Clear()
					msg, err := zcommon.Unpack(c.rdpkg)
					if err != nil {
						logx.ErrorF("%v", err)
						return
					}
					//得到当前客户端请求的Request数据
					req := zcommon.Request{
						Conn: c,
						Msg:  msg,
					}
					c.keepalive = 0
					if utils.GlobalObject.WorkerPoolSize > 0 {
						//已经启动工作池机制，将消息交给Worker处理
						c.MsgHandler.SendMsgToTaskQueue(&req)
					} else {
						//从绑定好的消息和对应的处理方法中执行对应的Handle方法
						go c.MsgHandler.DoMsgHandler(&req)
					}
				}()
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (c *Connection) startKeepAlive() {
	logx.Debugln("[KeepAlive Goroutine is running]")
	defer logx.Debugln(c.ClientIP(), "[conn KeepAlive exit!]")

	//20秒检测一次,180秒视为连接关闭,20秒无player属性视为非法
	interval_impl := time.Second * 20
	_timer := time.NewTimer(interval_impl)
	c.keepalive = 0
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-_timer.C:
			c.keepalive++
			if c.keepalive >= 9 {
				logx.ErrorF("tcp心跳超时")
				return
			} else if !c.IsValid() {
				logx.ErrorF("tcp连接20秒内都没有绑定player")
				return
			}

			_timer.Reset(interval_impl)
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (c *Connection) sendRest() {
	for {
		data, ok := <-c.msgBuffChan
		if !ok || data == nil {
			break
		}
		c.Conn.Write(data)
	}
	logx.DebugF("sendRest conn id = %d", c.ConnID)
}

// Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()
	//3 开启心跳检测
	go c.startKeepAlive()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TCPServer.CallOnConnStart(c)

	<-c.ctx.Done()

	//关闭连接,将还未发送完的数据发完
	c.sendRest()

	c.finalizer()
}

// Stop 停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	c.cancel()
}

// GetTCPConnection 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// GetConnID 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// RemoteAddr 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
func (c *Connection) ClientIP() string {
	return c.RemoteAddr().String()
}

// SendBuffMsg  发生BuffMsg
func (c *Connection) SendBuffMsg(ctx context.Context, msgID string, data []byte) error {
	c.RLock()
	defer c.RUnlock()
	if c.isClosed {
		return errors.New("Connection closed when send buff msg")
	}

	//将data封包，并且发送
	msg, err := zcommon.Pack(zcommon.NewMsgPackage(msgID, data, constants.ParseCtxRequestLocalNo(ctx)))
	if err != nil {
		logx.Debugln("Pack error msg ID = ", msgID)
		return errors.New("pack error msg ")
	}

	if msgID != "pb_battle.MsgNtfBattleFrames" {
		logx.DebugF("SendBuffMsg success, ConnID = %d, msgName = %s", c.ConnID, msgID)
	}
	//写回客户端
	c.msgBuffChan <- msg

	return nil
}
func (c *Connection) SendProtoBuffer(ctx context.Context, msg protoreflect.ProtoMessage) error { //直接将Message数据发送给远程的TCP客户端(有缓冲)
	name, data := helper.GetProtoMsgInfo(msg)
	return c.SendBuffMsg(ctx, name, data)
}

// SetProperty 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	if value != nil {
		c.property.Store(key, value)
	} else {
		c.property.Delete(key)
	}
}

// GetProperty 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	if value, ok := c.property.Load(key); ok && value != nil {
		return value, nil
	}

	return nil, errors.New("no property found")
}

// RemoveProperty 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.property.Delete(key)
}

// 返回ctx，用于用户自定义的go程获取连接退出状态
func (c *Connection) Context() context.Context {
	return c.ctx
}

func (c *Connection) finalizer() {
	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TCPServer.CallOnConnStop(c)

	c.Lock()
	defer c.Unlock()

	//如果当前链接已经关闭
	if c.isClosed {
		return
	}

	logx.Debugln("Conn Stop()...ConnID = ", c.ConnID)

	// 关闭socket链接
	_ = c.Conn.Close()

	//将链接从连接管理器中删除
	c.TCPServer.GetConnMgr().Remove(c)

	//关闭该链接全部管道
	close(c.msgBuffChan)
	//设置标志位
	c.isClosed = true
}
func (c *Connection) GetLimiterToken() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), zcommon.Limiter_Timeout)
	defer cancel()
	err := c.limiter.Wait(ctx)
	if err != nil {
		c.limitFailedCount++
	} else {
		c.limitFailedCount = 0
	}
	return c.limitFailedCount >= zcommon.Limiter_FailedMaxCount, err
}
func (c *Connection) IsValid() bool { //是否有效连接
	if c.valid {
		return true
	}
	if p, err := c.GetProperty("player"); err == nil && p != nil {
		c.valid = true
	}
	return c.valid
}
