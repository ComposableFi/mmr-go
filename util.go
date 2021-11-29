package mmr

func getPeakPosByHeight(height uint32) uint64 {
	return (1 << (height + 1)) - 2
}

func leftPeakHeightPos(mmrSize uint64) (uint32, uint64) {
	var height uint32 = 1
	var prevPos uint64 = 0
	var pos = getPeakPosByHeight(height)
	for pos < mmrSize {
		height += 1
		prevPos = pos
		pos = getPeakPosByHeight(height)
	}
	return height-1, prevPos
}

func siblingOffset(height uint32) uint64 {
	return (2 << height) - 1
}

func parentOffset(height uint32) uint64 {
	return 2 << height
}

func getRightPeak(height uint32, pos, mmrSize uint64) *peak {
	// move to right sibling pos
	pos += siblingOffset(height)
	// loop until we find a pos in mmr
	for pos > mmrSize -1 {
		if height == 0 {
			return nil
		}
		// move to left child
		pos -= parentOffset(height - 1)
		height -= 1
	}
	return &peak{height, pos}
}

func getPeaks(mmrSize uint64) (pos_s []uint64) {
	var height, pos = leftPeakHeightPos(mmrSize)
	pos_s = append(pos_s, pos)

	for height > 0 {
		p := getRightPeak(height, pos, mmrSize)
		if p == nil {
			break
		}
		height = p.height
		pos = p.pos
		pos_s = append(pos_s, pos)
	}

	return pos_s
}

