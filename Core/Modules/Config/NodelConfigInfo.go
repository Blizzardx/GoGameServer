package Config

import (
	"encoding/json"
	"io/ioutil"
	"litgame.cn/Server/Core/Network"
	"os"
	"runtime"
)

type NodeConfigGameServerInfo struct {
	Id                    int32  //实例id
	LogicId               int32  //逻辑id
	InternalAddress       string //内网ip
	ExternalAddress       string //外网ip
	ClientListenPort      string //客户端监听端口
	ClientListenPortProxy string //客户端监听端口代理
	GameServerListenPort  string //game server监听端口
	ClientProtocol        string // "tcp" "ws" ...
}
type NodeConfigComponentServerInfo struct {
	Id                   int32  //实例id
	LogicId              int32  //逻辑id 代表类型 （社交服务，聊天服务，战斗服务）
	InternalAddress      string //内网ip
	GameServerListenPort string //监听game server端口
}
type NodeConfigInfo struct {
	GameServerList               []*NodeConfigGameServerInfo
	SingletonComponentServerList []*NodeConfigComponentServerInfo //社交，聊天，公会 排行榜 等，只有一个实例的属于单例组件服务器
	MultiComponentServerList     []*NodeConfigComponentServerInfo //battle server等 可以动态负载均衡的，不属于单例组件服务器
}

//这句话是阻塞的
func FetchRemoteConfig(url string) *NodeConfigInfo {
	config := &NodeConfigInfo{}
	if runtime.GOOS != "linux" {
		fi, err := os.Open("Config/node-config.json")
		if err != nil {
			panic(err)
		}
		defer fi.Close()
		fd, err := ioutil.ReadAll(fi)
		json.Unmarshal(fd, config)
		return config
	}
	err := Network.HttpGet(url, config)
	if nil != err {
		return nil
	}

	return config
}
