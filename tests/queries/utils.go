package main

import (
	"math/rand"
	"strings"
	"time"
)

var (
	nodeStateFactor int64 = 3
	reportInterval        = time.Hour
)

func calcFreeResources(total node_resources_total, used node_resources_total) node_resources_total {
	if total.mru < used.mru {
		panic("total mru is less than mru")
	}
	if total.hru < used.hru {
		panic("total hru is less than hru")
	}
	if total.sru < used.sru {
		panic("total sru is less than sru")
	}
	return node_resources_total{
		hru: total.hru - used.hru,
		sru: total.sru - used.sru,
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
	return int64(timestamp) > time.Now().Unix()-nodeStateFactor*int64(reportInterval.Seconds())
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

func changeCase(s string) string {
	idx := rand.Intn(len(s))
	return strings.Replace(s, string(s[idx]), strings.ToUpper(string(s[idx])), 1)
}

func stringMatch(str string, sub_str string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(sub_str))
}
