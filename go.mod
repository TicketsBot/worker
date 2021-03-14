module github.com/TicketsBot/worker

go 1.14

require (
	github.com/TicketsBot/archiverclient v0.0.0-20210220155137-a562b2f1bbbb
	github.com/TicketsBot/common v0.0.0-20210314144843-3ac00a091e42
	github.com/TicketsBot/database v0.0.0-20210314143312-464ac4588cf2
	github.com/elliotchance/orderedmap v1.2.1
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/jackc/pgx/v4 v4.7.1
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/compress v1.10.10 // indirect
	github.com/rxdn/gdl v0.0.0-20210301221508-d84ed0db0f5c
	github.com/sirupsen/logrus v1.5.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	gopkg.in/alexcesaro/statsd.v2 v2.0.0
)
