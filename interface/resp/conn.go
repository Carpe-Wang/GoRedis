package resp

// Connection represents a connection with redis client
type Connection interface {
	Write([]byte) error
	// used for multi database
	GetDBIndex() int
	SelectDB(int)
}
