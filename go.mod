module github.com/TicketsBot/worker

go 1.14

require (
	github.com/TicketsBot/archiverclient v0.0.0-20210220155137-a562b2f1bbbb
	github.com/TicketsBot/common v0.0.0-20210224150303-f1fc5a6a0627
	github.com/TicketsBot/database v0.0.0-20210225212136-5ca68d2ef59f
	github.com/elliotchance/orderedmap v1.2.1
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/jackc/pgx/v4 v4.7.1
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/compress v1.10.10 // indirect
	github.com/rxdn/gdl v0.0.0-20210218195151-aa9dd0d1fa53
	github.com/sirupsen/logrus v1.5.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	gopkg.in/alexcesaro/statsd.v2 v2.0.0
)
