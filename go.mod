module github.com/TicketsBot/worker

go 1.18

require (
	github.com/TicketsBot/archiverclient v0.0.0-20220326163414-558fd52746dc
	github.com/TicketsBot/common v0.0.0-20220703211704-f792aa9f0c42
	github.com/TicketsBot/database v0.0.0-20220830131231-b5540b57f6cb
	github.com/caarlos0/env/v6 v6.9.3
	github.com/elliotchance/orderedmap v1.2.1
	github.com/gin-gonic/gin v1.7.1
	github.com/go-redis/redis/v8 v8.11.3
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/jackc/pgx/v4 v4.7.1
	github.com/json-iterator/go v1.1.12
	github.com/prometheus/client_golang v1.12.2
	github.com/rxdn/gdl v0.0.0-20220830131333-09a2e5819976
	github.com/schollz/progressbar/v3 v3.8.2
	github.com/sirupsen/logrus v1.6.0
	go.uber.org/atomic v1.6.0
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/alexcesaro/statsd.v2 v2.0.0
)

require (
	github.com/TicketsBot/logarchiver v0.0.0-20220326162808-cdf0310f5e1c // indirect
	github.com/TicketsBot/ttlcache v1.6.1-0.20200405150101-acc18e37b261 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/certifi/gocertifi v0.0.0-20200211180108-c7c1fbc02894 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-errors/errors v1.1.0 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.6.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.0.2 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200307190119-3430c5407db8 // indirect
	github.com/jackc/pgtype v1.4.0 // indirect
	github.com/jackc/pgx v3.6.2+incompatible // indirect
	github.com/jackc/puddle v1.1.1 // indirect
	github.com/juju/ratelimit v1.0.1 // indirect
	github.com/klauspost/compress v1.10.10 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pasztorpisti/qs v0.0.0-20171216220353-8d6c33ee906c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/tatsuworks/czlib v0.0.0-20190916144400-8a51758ea0d9 // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	nhooyr.io/websocket v1.8.4 // indirect
)

replace (
	github.com/TicketsBot/database => "../database"
)
