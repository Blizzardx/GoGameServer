package Network

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/rpc"
	"time"
)

//监听目标服务 使用tcp协议
func ListenAtTCP(address string, sessionChangeAction OnSessionChange, hooker cellnet.EventHooker, arg interface{}) *ServerObject {

	serverObject := createPipeline("tcp.ltv", "tcp.Acceptor", address, sessionChangeAction, hooker, arg)
	return serverObject
}

//链接目标服务 使用tcp协议
func ConnectToTCP(address string, sessionChangeAction OnSessionChange, hooker cellnet.EventHooker, arg interface{}) *ServerObject {
	serverObject := createPipeline("tcp.ltv", "tcp.Connector", address, sessionChangeAction, hooker, arg)
	serverObject.peer.(cellnet.TCPConnector).SetReconnectDuration(1)
	return serverObject
}

//监听目标服务 使用tcp协议
func ListenAtWS(address string, sessionChangeAction OnSessionChange, hooker cellnet.EventHooker, arg interface{}) *ServerObject {

	serverObject := createPipeline("gorillaws.ltv", "gorillaws.Acceptor", address, sessionChangeAction, hooker, arg)
	return serverObject
}

func RPC(session cellnet.Session, msg interface{}, timeOut time.Duration, userCallback func(raw interface{})) {
	// 异步RPC
	rpc.Call(session, msg, timeOut, userCallback)

}
