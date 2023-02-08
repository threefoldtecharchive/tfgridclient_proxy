// nolint
package main

// TODO: the one in tools/db/types.go is unexported but it's the same file

var (
	POSTGRES_HOST      string
	POSTGRES_PORT      int
	POSTGRES_USER      string
	POSTGRES_PASSSWORD string
	POSTGRES_DB        string
	ENDPOINT           string
	SEED               int
	STATUS_DOWN        = "down"
	STATUS_UP          = "up"
)

type contract_resources struct {
	id          string
	hru         uint64
	sru         uint64
	cru         uint64
	mru         uint64
	contract_id string
}
type farm struct {
	id                string
	grid_version      uint64
	farm_id           uint64
	name              string
	twin_id           uint64
	pricing_policy_id uint64
	certification     string
	stellar_address   string
	dedicated_farm    bool
}

type node struct {
	id                string
	grid_version      uint64
	node_id           uint64
	farm_id           uint64
	twin_id           uint64
	country           string
	city              string
	uptime            uint64
	created           uint64
	farming_policy_id uint64
	certification     string
	secure            bool
	virtualized       bool
	serial_number     string
	created_at        uint64
	updated_at        uint64
	location_id       string
}
type twin struct {
	id           string
	grid_version uint64
	twin_id      uint64
	account_id   string
	relay        string
	public_key   string
}
type public_ip struct {
	id          string
	gateway     string
	ip          string
	contract_id uint64
	farm_id     string
}
type node_contract struct {
	id                    string
	grid_version          uint64
	contract_id           uint64
	twin_id               uint64
	node_id               uint64
	deployment_data       string
	deployment_hash       string
	number_of_public_i_ps uint64
	state                 string
	created_at            uint64
	resources_used_id     string
}
type node_resources_total struct {
	id      string
	hru     uint64
	sru     uint64
	cru     uint64
	mru     uint64
	node_id string
}
type public_config struct {
	id      string
	ipv4    string
	ipv6    string
	gw4     string
	gw6     string
	domain  string
	node_id string
}
type rent_contract struct {
	id           string
	grid_version uint64
	contract_id  uint64
	twin_id      uint64
	node_id      uint64
	state        string
	created_at   uint64
}

type contract_bill_report struct {
	id                string
	contract_id       uint64
	discount_received string
	amount_billed     uint64
	timestamp         uint64
}

type name_contract struct {
	id           string
	grid_version uint64
	contract_id  uint64
	twin_id      uint64
	name         string
	state        string
	created_at   uint64
}
