package mmr

func pop(ph []interface{}) (interface{}, []interface{}) {
	if len(ph) == 0 {
		return nil, ph[:]
	}
	// return the last item in the slice and the rest of the slice excluding the last item
	return ph[len(ph)-1], ph[:len(ph)-1]
}
