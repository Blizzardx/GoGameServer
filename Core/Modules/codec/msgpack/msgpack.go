package msgpack

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/vmihailenco/msgpack"
)

type msgpackCodec struct {
}

func (self *msgpackCodec) Name() string {
	return "msgpack"
}
func (self *msgpackCodec) MimeType() string {
	return "application/MaJiangProto-pack"
}

func (self *msgpackCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {
	return msgpack.Marshal(msgObj)
}

func (self *msgpackCodec) Decode(data interface{}, msgObj interface{}) error {
	return msgpack.Unmarshal(data.([]byte), msgObj)
}

func init() {
	codec.RegisterCodec(new(msgpackCodec))
}
