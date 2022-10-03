package resp

type Connection interface {
	Write([]byte) error
	GetDBIndex() int
	SelectDB(int)
}
