package resp

// Connection 代表连接redis的客户端
type Connection interface {
	Write([]byte) error
	GetDBIndex() int //客户端连接的DB
	SelectDB(int)    //选择DB
}
