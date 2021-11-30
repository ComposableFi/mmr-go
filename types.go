package mmr

type leaf struct {
	pos  uint64
	hash interface{}
}

type leafWithHash struct {
	pos    uint64
	hash   interface{}
	height uint32
}

type peak struct {
	height uint32
	pos    uint64
}

func (l leafWithHash) popFront(leaves []leafWithHash) []leafWithHash {
	return leaves[1:]
}
