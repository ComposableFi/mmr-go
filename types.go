package mmr

type leaf struct {
	pos  uint64
	item interface{}
}

type peak struct {
	height uint32
	pos uint64
}