package main

import (
	"math/rand"
	"time"
)

func calcFreeResources(total node_resources_total, used node_resources_total) node_resources_total {
	if total.mru < used.mru {
		panic("total mru is less than mru")
	}
	if total.hru < used.hru {
		panic("total hru is less than hru")
	}
	if 2*total.sru < used.sru {
		panic("2 * total sru is less than sru")
	}
	return node_resources_total{
		hru: total.hru - used.hru,
		sru: 2*total.sru - used.sru,
		mru: total.mru - used.mru,
	}
}

func isIn(l []uint64, v uint64) bool {
	for _, i := range l {
		if i == v {
			return true
		}
	}
	return false
}

func isUp(timestamp uint64) bool {
	return int64(timestamp) > time.Now().Unix()*1000-nodeStateFactor*int64(reportInterval/time.Millisecond)
}

func flip(success float32) bool {
	return rand.Float32() < success
}

func rndref(min, max uint64) *uint64 {
	v := rand.Uint64()%(max-min+1) + min
	return &v
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
