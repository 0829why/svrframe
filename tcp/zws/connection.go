package zws

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/constants"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/helper"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/tcp/utils"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/tcp/zcommon"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/tcp/ziface"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

// Connection 链接
type Connection struct {
	//当前Conn属于哪个Server
	TCPServer ziface.IServer
	//当前连接的socket TCP套接字
	Conn *websocket.Conn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//消息管理MsgID和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle
	//告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan *zcommon.Message

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

	real_ip string
}

// NewConnection 创建连接的方法
func NewConnection(server ziface.IServer, conn *websocket.Conn, connID uint32, msgHandler ziface.IMsgHandle, real_ip string) *Connection {
	//初始化Conn属性
	limit := rate.Every(time.Millisecond * 200)
	c := &Connection{
		TCPServer:        server,
		Conn:             conn,
		ConnID:           connID,
		isClosed:         false,
		MsgHandler:       msgHandler,
		msgBuffChan:      make(chan *zcommon.Message, utils.GlobalObject.MaxMsgChanLen),
		property:         sync.Map{},
		rdpkg:            zcommon.NewReadPackage(),
		limiter:          rate.NewLimiter(limit, zcommon.Limiter_bucket),
		limitFailedCount: 0,
		valid:            false,
	}
	c.real_ip = real_ip

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
				msg, err := zcommon.Pack(data)
				if err != nil {
					logx.ErrorF("Pack error  = %v", err)
					return
				}
				//有数据要写给客户端
				err = c.Conn.WriteMessage(websocket.BinaryMessage, msg)
				if err != nil {
					logx.Errorln("Send Data error:, ", err, " Conn Writer exit")
					return
				}
				if data.ID != "pb_battle.MsgNtfBattleFrames" {
					logx.DebugF("SendBuffMsg success, ConnID = %d, msgName = %s, RequestNo = %v", c.ConnID, data.ID, data.RequestNo)
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
			t, b, err := c.Conn.ReadMessage()
			if err != nil {
				logx.ErrorF("StartReader ReadMessage => %v", err)
				return
			}
			if t == websocket.TextMessage {
				logx.DebugF("TextMessage -> %s", string(b))
			} else if t == websocket.PingMessage {
				c.keepalive = 0
				logx.DebugF("PingMessage -> %s", string(b))
			} else if t == websocket.PongMessage {
				c.keepalive = 0
				logx.DebugF("PongMessage -> %s", string(b))
			} else if t == websocket.CloseMessage {
				return
			} else if t == websocket.BinaryMessage {
				func() {
					msg, err := zcommon.UnpackFromBytes(b, uint32(len(b)))
					if err != nil {
						logx.ErrorF("StartReader UnpackFromBytes data => %s, err => %v", string(b), err)
						return
					}
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
			} else {
				logx.ErrorF("非法消息类型 -> %d", t)
				return
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (c *Connection) startKeepAlive() {
	logx.Debugln("[KeepAlive Goroutine is running]")
	defer logx.Debugln(c.ClientIP(), "[conn KeepAlive exit!]")
	defer c.Stop()

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
	defer logx.Debugln(c.ClientIP(), "[sendRest!]")
	for {
		select {
		case data, ok := <-c.msgBuffChan:
			if !ok || data == nil {
				break
			}
			logx.Debugln("sendRest")
			msg, err := zcommon.Pack(data)
			if err == nil {
				c.Conn.WriteMessage(websocket.BinaryMessage, msg)
			}
		default:
			return
		}
	}
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

	logx.Debugln(c.ClientIP(), "[running!]")
	<-c.ctx.Done()
	logx.Debugln(c.ClientIP(), "[stoping!]")

	//关闭连接,将还未发送完的数据发完
	c.sendRest()

	c.finalizer()
}

// Stop 停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	c.cancel()
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
	return c.real_ip
}

// SendBuffMsg  发生BuffMsg
func (c *Connection) SendBuffMsg(ctx context.Context, msgID string, data []byte) error {
	c.RLock()
	defer c.RUnlock()
	if c.isClosed {
		return errors.New("Connection closed when send buff msg")
	}

	//将data封包，并且发送
	//msg, err := zcommon.Pack(zcommon.NewMsgPackage(msgID, data))
	//if err != nil {
	//	logx.Debugln("Pack error msg ID = ", msgID)
	//	return errors.New("Pack error msg ")
	//}

	//写回客户端
	c.msgBuffChan <- zcommon.NewMsgPackage(msgID, data, constants.ParseCtxRequestLocalNo(ctx))

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
