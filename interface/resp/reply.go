package resp

type Reply interface {
	ToBytes() []byte
}
