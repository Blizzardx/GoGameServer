package GameServer

import (
	"github.com/davyxu/cellnet"
	"litgame.cn/Server/Core/Network"
)

func (server *GameServer) listenGameServer() {
	//监听其他gameServer服务器连接
	server.listenGameServerSession = Network.ListenAtTCP("0.0.0.0:"+server.selfConfig.GameServerListenPort, server.onListenServerSessionChange, server, nil)

}

//监听： 有GameServer 建立连接 或 有GameServer离线
func (server *GameServer) onListenServerSessionChange(eventName string, ev cellnet.Event, arg interface{}) {
}
