package InternalMessage

type InternalMsg struct {
	PlayerId    string
	MessageId   int
	MessageBody []byte
	MsgUnEncode interface{}
}
type RegisterGameServerMsg struct {
	GameServerLogicId int32
}
