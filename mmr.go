package mmr

import (
	"fmt"
	"reflect"
	"sort"
)

type MMR struct {
	size  uint64
	batch *Batch
	merge Merge
}

func NewMMR(mmrSize uint64, s Store, m Merge) *MMR {
	return &MMR{
		size:  mmrSize,
		batch: newBatch(s),
		merge: m,
	}
}

func (m *MMR) findElem(pos uint64, hashes []interface{}) (interface{}, error) {
	if posOffset := pos - m.size; posOffset < 0 {
		return uint(posOffset), nil
	}

	elem, err := m.batch.getElem(pos)
	if err != nil {
		// replace with custom error
		return nil, fmt.Errorf("InconsitentStore")
	}
	return elem, nil
}

func (m *MMR) MMRSize() uint64 {
	return m.size
}

func (m *MMR) IsEmpty() bool {
	return m.size == 0
}

// push a element and return position
func (m *MMR) Push(elem interface{}) interface{} {
	var elems []interface{}
	// position of new elem
	elemPos := m.size
	elems = append(elems, elem)

	var height uint32 = 0
	var pos = elemPos
	// continue to merge tree node if next pos higher than current
	for posHeightInTree(pos+1) > height {
		pos += 1
		leftPos := pos - parentOffset(height)
		rightPos := leftPos + siblingOffset(height)
		leftElem, _ := m.findElem(leftPos, elems)
		rightElem, _ := m.findElem(rightPos, elems)
		parentElem := m.merge.Merge(leftElem, rightElem)
		elems = append(elems, parentElem)
		height += 1
	}
	// store hashes
	m.batch.append(elemPos, elems)
	// update mmrSize
	m.size = pos + 1
	return elemPos
}


func (m *MMR) GetRoot() (interface{}, error) {
	if m.size == 0 {
		// TODO: replace with custom error ttoe
		return nil, fmt.Errorf("GetRootOnEmpty")
	} else if m.size == 1 {
		e, err := m.batch.getElem(0)
		if err != nil {
			// TODO: replace with custom error ttoe
			return nil, fmt.Errorf("InconsistentStore")
		}
		return e, nil
	}

	var peaks []interface{}
	for _, peakPos := range getPeaks(m.size) {
		elem, err := m.batch.getElem(peakPos)
		if err != nil {
			return nil, fmt.Errorf("InconsistentStore")
		}
		peaks = append(peaks, elem)
	}

	if peak := m.bagRHSPeaks(peaks); peak != nil {
		return peak, nil
	}

	return nil, fmt.Errorf("InconsistentStore")
}

func (m *MMR) bagRHSPeaks(rhsPeaks []interface{}) interface{} {
	for len(rhsPeaks) > 1 {
		var rp, lp interface{}
		rp, rhsPeaks = pop(rhsPeaks)
		lp, rhsPeaks = pop(rhsPeaks)
		rhsPeaks = append(rhsPeaks, m.merge.Merge(rp, lp))
	}
	return rhsPeaks
}

/// generate merkle proof for a peak
/// the pos_list must be sorted, otherwise the behaviour is undefined
///
/// 1. find a lower tree in peak that can generate a complete merkle proof for position
/// 2. find that tree by compare positions
/// 3. generate proof for each positions
func (m *MMR) genProofForPeak(proof []interface{}, posList []uint64, peakPos uint64) ([]interface{}, error) {
	if len(posList) == 1 && reflect.DeepEqual(posList, []uint64{peakPos}) {
		return []interface{}{}, nil
	}
	// take peak root from store if no positions need to be proof
	if len(posList) == 0 {
		elem, err := m.batch.getElem(peakPos)
		if err != nil {
			return []interface{}{}, fmt.Errorf("InconsistentStore")
		}
		proof = append(proof, elem)
		return proof, nil
	}

	var queue []peak

	for _, p := range posList {
		queue = append(queue, peak{pos: p, height: 0})
	}
}

/// Generate merkle proof for positions
/// 1. sort positions
/// 2. push merkle proof to proof by peak from left to right
/// 3. push bagged right hand side root
func (m *MMR) GenProof() interface{} {
	return nil
}

func (m *MMR) Commit() interface{} {
	return nil
}

type MerkleProof struct {
	mmrSize uint64
	proof  	[]interface{}
}

func NewMerkleProof(mmrSize uint64, proof []interface{}) *MerkleProof {
	return &MerkleProof{
		mmrSize: mmrSize,
		proof:   proof,
	}
}

func (m *MerkleProof) MMRSize() uint64 {
	return m.mmrSize
}

func (m *MerkleProof) ProofItems() []interface{} {
	return m.proof
}

func (m *MerkleProof) CalculateRoot(leaves []leaf) (interface{}, error) {
	return calculateRoot(leaves, m.mmrSize, &Iterator{item: m.proof})
}

func (m *MerkleProof) CalculateRootWithNewLeaf() (interface{}, error) {
	return nil, nil
}

func (m *MerkleProof) Verify(root interface{}, leaves []leaf) (bool, error) {
	return false, nil
}

func calculatePeakRoot(leaves []leaf, peakPos uint64, proofs *Iterator) (interface{}, error) {
	if len(leaves) == 0 {
		// TODO: clarify on how debug_assert! works
		panic("can't be empty")
	}

	// (position, hash, height)
	var queue []leafWithHash
	for _, l := range leaves {
		queue = append(queue, leafWithHash{l.pos, l.hash, 0})
	}

	// calculate tree root from each items
	for len(queue) > 0 {
		pop := queue[0]
		// pop from front
		queue = queue[1:]

		pos, item, height := pop.pos, pop.hash, pop.height
		if pos == peakPos {
			return item, nil
		}
		// calculate sibling
		var nextHeight = posHeightInTree(pos + 1)
		var sibPos, parentPos = func() (uint64, uint64) {
			var siblingOffset uint64 = siblingOffset(height)
			if nextHeight > height {
				// implies pos is right sibling
				return pos - siblingOffset, pos + 1
			} else {
				// pos is left sibling
				return pos + siblingOffset, pos + parentOffset(height)
			}
		}()

		var siblingItem interface{}
		if len(queue) > 0 && queue[0].pos == sibPos {
			siblingItem, queue = queue[0].hash, queue[1:]
		} else {
			if siblingItem = proofs.next(); siblingItem == nil {
				// replace with custom error
				return nil, fmt.Errorf("corruptedProof")
			}
		}

		var parentItem interface{}
		if nextHeight > height {
			// TODO: implement actual merge method
			merge(siblingItem, item)
		} else {
			// TODO: implement actual merge method
			merge(item, siblingItem)
		}

		if parentPos < peakPos {
			queue = append(queue, leafWithHash{parentPos, parentItem, height + 1})
		} else {
			return parentItem, nil
		}
	}

	return nil, fmt.Errorf("corruptedProof")
}

func baggingPeaksHashes(peaksHashes []interface{}) (interface{}, error) {
	var rightPeak, leftPeak interface{}
	for len(peaksHashes) > 1 {
		if rightPeak, peaksHashes = pop(peaksHashes); rightPeak == nil {
			panic("pop")
		}

		if leftPeak, peaksHashes = pop(peaksHashes); leftPeak == nil {
			panic("pop")
		}
		peaksHashes = append(peaksHashes, merge(rightPeak, leftPeak))
	}

	if len(peaksHashes) == 0 {
		return nil, fmt.Errorf("corruptedProof")
	}
	return peaksHashes[len(peaksHashes)-1], nil
}

/// merkle proof
/// 1. sort items by position
/// 2. calculate root of each peak
/// 3. bagging peaks
func calculateRoot(leaves []leaf, mmrSize uint64, proofs *Iterator) (interface{}, error) {
	var peaksHashes, err = calculatePeaksHashes(leaves, mmrSize, proofs)
	if err != nil {
		return nil, err
	}

	return baggingPeaksHashes(peaksHashes)
}

// TODO: create and return custom error type
func calculatePeaksHashes(leaves []leaf, mmrSize uint64, proofs *Iterator) ([]interface{}, error) {
	// special handle the only 1 leaf MMR
	if mmrSize == 1 && len(leaves) == 1 && leaves[0].hash == 0 {
		var items []interface{}
		for _, l := range leaves {
			items = append(items, l.hash)
		}
		return items, nil
	}

	// sort items by position
	sort.SliceStable(leaves, func(i, j int) bool {
		return leaves[i].pos < leaves[j].pos
	})

	peaks := getPeaks(mmrSize)
	peaksHashes := make([]interface{}, 0, len(peaks)+1)
	for _, peaksPos := range peaks {
		var leaves = takeWhileVec(leaves, func(l leaf) bool {
			return l.pos <= peaksPos
		})
		var peakRoot interface{}
		if len(leaves) == 1 && leaves[0].pos == peaksPos {
			// leaf is the peak
			peakRoot = leaves[0].hash
		} else if len(leaves) == 0 {
			// if empty, means the next proof is a peak root or rhs bagged root
			if proofs.isEmpty() {
				peakRoot = proofs.next()
			} else {
				// means that either all right peaks are bagged, or proof is corrupted
				// so we break loop and check no items left
				break
			}
		} else {
			// TODO: test calculatePeakRoot
			_, err := calculatePeakRoot(leaves, peaksPos, proofs)
			if err != nil {
				return nil, err
			}
		}
		peaksHashes = append(peaksHashes, peakRoot)
	}

	// ensure nothing left in leaves
	if len(leaves) != 0 {
		// replace with custom error
		return nil, fmt.Errorf("corruptedProof")
	}

	// check rhs peaks
	if rhsPeaksHashes := proofs.next(); rhsPeaksHashes != nil {
		peaksHashes = append(peaksHashes, rhsPeaksHashes)
	}
	// ensure nothing left in proof_iter
	if proofs.next() != nil {
		// replace with custom error
		return nil, fmt.Errorf("corruptedProof")
	}

	return peaksHashes, nil
}

func takeWhileVec(v []leaf, p func(leaf) bool) []leaf {
	for i := 0; i < len(v); i++ {
		if !p(v[i]) {
			return v[:i]
		}
	}
	return v[:]
}
