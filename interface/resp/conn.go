package resp

// Connection represents a connection with redis client
type Connection interface {
	Write([]byte) error
	// used for multi database
	GetDBIndex() int //客户端连接的DB
	SelectDB(int)    //选择DB
}
