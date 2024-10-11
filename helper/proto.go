package helper

import (
	"errors"
	"fmt"
	"strings"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func GetProtoMessageName(protoFullName string) (name string) {
	items := strings.Split(protoFullName, ".")
	name = items[len(items)-1]
	return
}
func GetProtoMessagePrefixName(protoFullName string) (prefix string) {
	items := strings.Split(protoFullName, ".")
	prefix = strings.Join(items[0:len(items)-1], ".")
	return
}
func GetProtoMessageTypeByName(protoFullName string) protoreflect.MessageType {
	msgName := protoreflect.FullName(protoFullName)
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(msgName)
	if err != nil {
		logx.ErrorF("GetProtoMessageTypeByName err: %v", err)
		return nil
	}
	return msgType
}
func NewProtoMessageByName(protoFullName string) (msg proto.Message, err error) {
	msgType := GetProtoMessageTypeByName(protoFullName)
	if msgType == nil {
		err = fmt.Errorf("can't find message type")
		logx.ErrorF("NewProtoMessageByName err: %v", err)
		return
	}
	msg = msgType.New().Interface()
	err = nil
	return
}
func MakeProtoMessage(protoFullName string, data []byte) proto.Message {
	msg, err := NewProtoMessageByName(protoFullName)
	if err != nil {
		return nil
	}
	if msg == nil {
		logx.ErrorF("msg = nil")
		return nil
	}
	err = proto.Unmarshal(data, msg)
	if err != nil {
		logx.ErrorF("err = %v", err)
		return nil
	}

	return msg
}
func MakeProtoMessage1(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}
func MakeProtoMessage2[T protoreflect.ProtoMessage](protoFullName string, data []byte) (ret T, err error) {
	msg := MakeProtoMessage(protoFullName, data)
	if msg == nil {
		err = errors.New("MakeProtoMessage2 Error, protoFullName -> " + protoFullName)
		return
	}
	ret, ok := msg.(T)
	if !ok {
		return ret, errors.New("MakeProtoMessage2 Error, protoFullName -> " + protoFullName)
	}
	return ret, nil
}
func GetProtoMsgInfo(msg proto.Message) (name string, data []byte) {
	name = GetProtoFullName(msg)
	data, _ = proto.Marshal(msg)
	return
}
func GetProtoFullName(msg proto.Message) string {
	return string(proto.MessageName(msg))
}
