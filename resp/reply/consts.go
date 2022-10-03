package reply

// PongReply
type PongReply struct { //ping指令的恢复

}

var pongReply = []byte("+PONG\r\n") //\r\n是redis回复的标准

func (p PongReply) ToBytes() []byte { //实现reply接口
	return pongReply
}
func makePongReply() *PongReply { //方便外部调用
	return &PongReply{}
}

// OkReply
type OkReply struct{}

var okBytes = []byte("+OK\r\n")

func (r *OkReply) ToBytes() []byte {
	return okBytes
}

var theOkReply = new(OkReply)

func MakeOkReply() *OkReply {
	return theOkReply
}

var nullBulkBytes = []byte("$-1\r\n")

type NullBulkReply struct {
}

func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}
func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

var emptyMultiBulkBytes = []byte("*0\r\n")

type EmptyMultiBulkReply struct{}

func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

type NoReply struct{}

var noBytes = []byte("")

func (r *NoReply) ToBytes() []byte {
	return noBytes
}
