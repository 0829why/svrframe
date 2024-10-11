package zcommon

// Message 消息
type Message struct {
	DataLen   uint32 //消息的长度
	RequestNo int32  //请求序列号
	ID        string //消息的ID
	Data      []byte //消息的内容
}

// NewMsgPackage 创建一个Message消息包
func NewMsgPackage(ID string, data []byte, request_no int32) *Message {
	return &Message{
		DataLen:   uint32(len(data)),
		RequestNo: request_no,
		ID:        ID,
		Data:      data,
	}
}

// GetDataLen 获取消息数据段长度
func (msg *Message) GetDataLen() uint32 {
	return msg.DataLen
}

// GetDataLen 获取消息请求序列号
func (msg *Message) GetRequestNo() int32 {
	return msg.RequestNo
}

// GetMsgID 获取消息ID
func (msg *Message) GetMsgID() string {
	return msg.ID
}

// GetData 获取消息内容
func (msg *Message) GetData() []byte {
	return msg.Data
}

// SetDataLen 设置消息数据段长度
func (msg *Message) SetDataLen(len uint32) {
	msg.DataLen = len
}

// SetRequestNo 设计消息请求序列号
func (msg *Message) SetRequestNo(request_no int32) {
	msg.RequestNo = request_no
}

// SetMsgID 设计消息ID
func (msg *Message) SetMsgID(msgID string) {
	msg.ID = msgID
}

// SetData 设计消息内容
func (msg *Message) SetData(data []byte) {
	msg.Data = data
}
