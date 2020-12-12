module github.com/TicketsBot/worker

go 1.14

require (
	github.com/TicketsBot/archiverclient v0.0.0-20200703191016-b27de6fd6919
	github.com/TicketsBot/common v0.0.0-20201123160756-4e90f3902175
	github.com/TicketsBot/database v0.0.0-20201105181405-1cf81496bbca
	github.com/TicketsBot/logarchiver v0.0.0-20200425163447-199b93429026 // indirect
	github.com/elliotchance/orderedmap v1.2.1
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/jackc/pgtype v1.4.0
	github.com/jackc/pgx/v4 v4.7.1
	github.com/klauspost/compress v1.10.10 // indirect
	github.com/rxdn/gdl v0.0.0-20201211152733-ddf7d77e7684
	github.com/sirupsen/logrus v1.5.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	gopkg.in/alexcesaro/statsd.v2 v2.0.0
)
