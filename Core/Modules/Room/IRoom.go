package Room

type IRoom interface {
	GetMaxSilenceTime() int32                                           //设置房间最大静默时间 超过静默时间将自动回收房间
	GetTickRate() int32                                                 //设置房间心跳频率 只初始化用 中途修改无效
	OnRoomMessage(playerId string, msgName string, msgBody interface{}) //房间收到消息
	OnInit()                                                            //房间初始化
	OnTick()                                                            //房间心跳
	OnDelete()                                                          //房间结束
}
