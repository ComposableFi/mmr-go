package mmr

import (
	"fmt"
	"sort"
)

func calculatePeakRoot() {}

func calculateRoot() {}

func takeWhileVec(v []leaf, p func(leaf)bool) []leaf {
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
	if mmrSize == 1 && len(leaves) == 1 && leaves[0].item == 0 {
		var items []interface{}
		for _, l := range leaves {
			items = append(items, l.item)
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
			peakRoot = leaves[0].item
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
			// TODO: implement method
			calculatePeakRoot()
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
	if proofs.next() != nil {
		// replace with custom error
		return nil, fmt.Errorf("corruptedProof")
	}

	// ensure nothing left in proof_iter

	return peaksHashes, nil
}
