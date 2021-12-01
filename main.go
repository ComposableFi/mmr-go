package mmr

import (
	"fmt"
	"sort"
)

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
	pop := func(ph []interface{}) (interface{}, []interface{}) {
		if len(ph) == 0 {
			return nil, ph[:]
		}
		// return the last item in the slice and the rest of the slice excluding the last item
		return ph[len(ph)-1], ph[:len(ph)-1]
	}

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

func calculateRoot(leaves []leaf, mmrSize uint64, proofs *Iterator) (interface{}, error) {
	var peaksHashes, err = calculatePeaksHashes(leaves, mmrSize, proofs)
	if err != nil {
		return nil, err
	}

	return baggingPeaksHashes(peaksHashes)
}

func takeWhileVec(v []leaf, p func(leaf) bool) []leaf {
	for i := 0; i < len(v); i++ {
		if !p(v[i]) {
			return v[:i]
		}
	}
	return v[:]
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
