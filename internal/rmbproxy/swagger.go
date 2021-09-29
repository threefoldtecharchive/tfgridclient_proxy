package rmbproxy

// Message swagger example
type Message struct {
	Version    int      `json:"ver" example:"1"`
	UID        string   `json:"uid" example:""`
	Command    string   `json:"cmd" example:"zos.statistics.get"`
	Expiration int      `json:"exp" example:"0"`
	Retry      int      `json:"try" example:"2"`
	Data       string   `json:"dat" example:""`
	TwinSrc    uint32   `json:"src" example:"1"`
	TwinDest   []uint32 `json:"dst" example:[ "2" ]`
	Retqueue   string   `json:"ret" example:""`
	Schema     string   `json:"shm" example:""`
	Epoch      int64    `json:"now" example:"1631078674"`
	Err        string   `json:"err" example:""`
}
