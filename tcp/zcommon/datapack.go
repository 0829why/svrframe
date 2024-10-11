package zcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/tcp/utils"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/tcp/ziface"
)

type DataPack struct{}

// Pack 封包方法(压缩数据)
func Pack(msg ziface.IMessage) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	nameSize := uint8(len(msg.GetMsgID()))
	//数据整合总长度 = 头长度+序列号字节长度+id字节长度+id长度+消息长度
	totalSize := uint32(uint32(binary.Size(msg.GetDataLen())) + uint32(binary.Size(msg.GetRequestNo())) + uint32(binary.Size(nameSize)) + uint32(nameSize) + msg.GetDataLen())

	if utils.GlobalObject.MaxPacketSize > 0 && totalSize > utils.GlobalObject.MaxPacketSize {
		return nil, fmt.Errorf("err TotalSize = %d", totalSize)
	}

	//先写入总长度
	if err := binary.Write(dataBuff, ByteOrder, totalSize); err != nil {
		return nil, err
	}
	//再写入序列号长度
	if err := binary.Write(dataBuff, ByteOrder, msg.GetRequestNo()); err != nil {
		return nil, err
	}
	//再写入id字节长度
	if err := binary.Write(dataBuff, ByteOrder, nameSize); err != nil {
		return nil, err
	}
	//再写入id
	if err := binary.Write(dataBuff, ByteOrder, []byte(msg.GetMsgID())); err != nil {
		return nil, err
	}
	//再写入data数据
	data := msg.GetData()
	if len(data) > 0 {
		if err := binary.Write(dataBuff, ByteOrder, data); err != nil {
			return nil, err
		}
	}

	return dataBuff.Bytes(), nil
}

// Unpack 拆包方法(解压数据)
func Unpack(rdpkg *ReadPackage) (ziface.IMessage, error) {

	//跳过数据总长度的字节,取后续内容
	binaryData := rdpkg.Data[binary.Size(rdpkg.TotalSize):]
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//先读取4字节序列号
	var request_no int32 = 0
	if err := binary.Read(dataBuff, ByteOrder, &request_no); err != nil {
		return nil, err
	}

	//读id字节长度
	var nameSize uint8 = 0
	if err := binary.Read(dataBuff, ByteOrder, &nameSize); err != nil {
		return nil, err
	}
	//再读id
	name := make([]byte, nameSize)
	if err := binary.Read(dataBuff, ByteOrder, name); err != nil {
		return nil, err
	}

	msg := &Message{
		DataLen: rdpkg.TotalSize - uint32(binary.Size(request_no)) - uint32(binary.Size(nameSize)) - uint32(nameSize),
	}

	msg.DataLen -= uint32(binary.Size(rdpkg.TotalSize))

	var data []byte
	if msg.DataLen > 0 {
		data = make([]byte, msg.DataLen)
		//再读dataLen
		if err := binary.Read(dataBuff, ByteOrder, data); err != nil {
			return nil, err
		}
	}

	msg.RequestNo = request_no
	msg.ID = string(name)
	msg.Data = data

	return msg, nil
}

// Unpack 拆包方法(解压数据)
func UnpackFromBytes(totaldata []byte, totalsize uint32) (ziface.IMessage, error) {

	//跳过数据总长度的字节,取后续内容
	binaryData := totaldata[binary.Size(totalsize):]
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//先读取4字节序列号
	var request_no int32 = 0
	if err := binary.Read(dataBuff, ByteOrder, &request_no); err != nil {
		return nil, err
	}

	//先读id字节长度
	var nameSize uint8 = 0
	if err := binary.Read(dataBuff, ByteOrder, &nameSize); err != nil {
		return nil, err
	}
	//再读id
	name := make([]byte, nameSize)
	if err := binary.Read(dataBuff, ByteOrder, name); err != nil {
		return nil, err
	}

	msg := &Message{
		DataLen: totalsize - uint32(binary.Size(request_no)) - uint32(binary.Size(nameSize)) - uint32(nameSize),
	}

	msg.DataLen -= uint32(binary.Size(totalsize))
	var data []byte
	if msg.DataLen > 0 {
		data = make([]byte, msg.DataLen)
		//再读dataLen
		if err := binary.Read(dataBuff, ByteOrder, data); err != nil {
			return nil, err
		}
	}
	msg.RequestNo = request_no
	msg.ID = string(name)
	msg.Data = data

	return msg, nil
}
