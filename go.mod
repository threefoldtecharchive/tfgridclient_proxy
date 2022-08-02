module github.com/threefoldtech/grid_proxy_server

go 1.17

require (
	github.com/alecthomas/assert v1.0.0
	github.com/go-acme/lego/v4 v4.4.0
	github.com/gorilla/mux v1.8.0
	github.com/lib/pq v1.10.4
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.26.0
	github.com/stretchr/testify v1.8.0
	github.com/swaggo/http-swagger v1.1.1
	github.com/swaggo/swag v1.8.1
	github.com/threefoldtech/substrate-client v0.0.0-20220629072807-544d41e5b81a
	github.com/threefoldtech/zos v0.5.6-0.20220629073825-73f495f28c80
	gorm.io/driver/postgres v1.3.5
	gorm.io/gorm v1.23.5
)

require (
	github.com/ChainSafe/go-schnorrkel v1.0.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v3 v3.2.2
	github.com/cenkalti/backoff/v4 v4.1.0 // indirect
	github.com/cosmos/go-bip39 v1.0.0 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/base58 v1.0.3 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/ethereum/go-ethereum v1.10.17 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gtank/merlin v0.1.1 // indirect
	github.com/gtank/ristretto255 v0.1.2 // indirect
	github.com/jbenet/go-base58 v0.0.0-20150317085156-6237cf65f3a6 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/miekg/dns v1.1.40 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20210601165009-122bf33a46e0 // indirect
	github.com/pierrec/xxHash v0.1.5 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14 // indirect
	github.com/vedhavyas/go-subkey v1.0.3 // indirect
	golang.org/x/crypto v0.0.0-20220518034528-6f7dac969898 // indirect
	golang.org/x/net v0.0.0-20220403103023-749bd193bc2b // indirect
	golang.org/x/sys v0.0.0-20220403020550-483a9cbc67c0 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.10 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require (
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20210801044451-80ca428c5142 // indirect
	github.com/centrifuge/go-substrate-rpc-client/v4 v4.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.12.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.11.0 // indirect
	github.com/jackc/pgx/v4 v4.16.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/centrifuge/go-substrate-rpc-client/v4 v4.0.2 => github.com/threefoldtech/go-substrate-rpc-client/v4 v4.0.3-0.20220629062423-919d421680f4
