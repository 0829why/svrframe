package zcommon

import (
	"oversea-git.hotdogeth.com/poker/slots/svrframe/tcp/ziface"
)

// Request 请求
type Request struct {
	Conn ziface.IConnection //已经和客户端建立好的 链接
	Msg  ziface.IMessage    //客户端请求的数据
}

// GetConnection 获取请求连接信息
func (r *Request) GetConnection() ziface.IConnection {
	return r.Conn
}

// GetData 获取请求消息的数据
func (r *Request) GetData() []byte {
	return r.Msg.GetData()
}

// GetMsgID 获取请求的消息的ID
func (r *Request) GetMsgRequestNo() int32 {
	return r.Msg.GetRequestNo()
}

// GetMsgID 获取请求的消息的ID
func (r *Request) GetMsgID() string {
	return r.Msg.GetMsgID()
}
