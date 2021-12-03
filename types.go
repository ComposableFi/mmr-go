package mmr

type Merge interface {
	Merge(left, right interface{}) interface{}
}

// Tree is representation type for the merkle tree
type Tree struct {
	Nodes []interface{}
	Merge Merge
}

func merge(item interface{}, item2 interface{}) interface{} {
	return nil
}

type leaf struct {
	pos  uint64
	hash interface{}
}

type leafWithHash struct {
	pos    uint64
	hash   interface{}
	height uint32
}

func (l leafWithHash) popFront(leaves []leafWithHash) []leafWithHash {
	return leaves[1:]
}

type peak struct {
	height uint32
	pos    uint64
}
