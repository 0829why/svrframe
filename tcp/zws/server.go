package zws

import (
	"fmt"
	"net/http"

	"github.com/0829why/svrframe/constants"
	"github.com/0829why/svrframe/logx"
	"github.com/0829why/svrframe/tcp/utils"
	"github.com/0829why/svrframe/tcp/zcommon"
	"github.com/0829why/svrframe/tcp/ziface"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Server 接口实现，定义一个Server服务类
type Server struct {
	//服务器的名称
	Name string
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgID和对应的处理方法
	msgHandler ziface.IMsgHandle
	//当前Server的链接管理器
	ConnMgr ziface.IConnManager
	//该Server的连接创建时Hook函数
	OnConnStart func(conn ziface.IConnection)
	//该Server的连接断开时的Hook函数
	OnConnStop func(conn ziface.IConnection)
	//是否关闭中
	Closing bool
}

// NewServer 创建一个服务器句柄
func NewServer() ziface.IServer {
	zcommon.PrintLogo()

	s := &Server{
		Name:       utils.GlobalObject.Name,
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TCPPort,
		msgHandler: zcommon.NewMsgHandle(),
		ConnMgr:    zcommon.NewConnManager(),
	}

	return s
}

// ============== 实现 ziface.IServer 里的全部接口方法 ========

func (s *Server) IsClosing() bool { //是否关闭中
	return s.Closing
}

func (s *Server) Close() { //是否关闭中
	s.Closing = true
}

func (s *Server) handshake(c *gin.Context) {
	if websocket.IsWebSocketUpgrade(c.Request) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, c.Writer.Header())
		if err != nil {
			logx.ErrorF("upgrade err -> : %v", err)
			return
		}
		if s.Closing || constants.GetSystemStatus() == constants.SystemStatus_Maintain {
			//正在关闭中,不接收新连接了
			return
		}

		//3.2 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
		if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
			logx.ErrorF("当前连接数量超过最大值,放弃新连接")
			conn.Close()
		}

		//TODO server.go 应该有一个自动生成ID的方法

		//3.3 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
		id := cID.Add(1)
		dealConn := NewConnection(s, conn, id, s.msgHandler, c.ClientIP())
		if cID.Load() >= 0x7fffffff {
			cID.Store(0)
		}

		//3.4 启动当前链接的处理业务
		go dealConn.Start()
	} else {
		logx.ErrorF("不是websocket请求")
	}
}

// Start 开启网络服务
func (s *Server) Start() {
	fmt.Printf("[START] Server name: %s,listenner at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)

	//开启一个go去做服务端Linster业务
	go func() {
		//0 启动worker工作池机制
		s.msgHandler.StartWorkerPool()

		g := gin.Default()
		g.GET("/", s.handshake)

		addr := fmt.Sprintf(":%d", s.Port)
		httpServer := &http.Server{
			Addr:    addr,
			Handler: g,
		}

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.ErrorF("ListenAndServe => %v", err)
		}
	}()
}

// Stop 停止服务
func (s *Server) Stop() {
	logx.Debugln("[STOP] Zinx server , name ", s.Name)

	//将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.ConnMgr.ClearConn()
}

// Serve 运行服务
func (s *Server) Serve() {
	s.Start()

	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则主Go退出， listenner的go将会退出
	select {}
}

// RegisterRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) RegisterRouter(msgID string, router ziface.IRouter) {
	s.msgHandler.RegisterRouter(msgID, router)
}

// 路由功能: 没有指定的消息,都通过这个处理
func (s *Server) RegisterGlobalRouter(router ziface.IRouter) {
	s.msgHandler.RegisterGlobalRouter(router)
}

// GetConnMgr 得到链接管理
func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

// SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		logx.Debugln("---> CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

// CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		logx.Debugln("---> CallOnConnStop....")
		s.OnConnStop(conn)
	}
}
