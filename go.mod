module github.com/TicketsBot/worker

go 1.14

require (
	github.com/TicketsBot/archiverclient v0.0.0-20210220155137-a562b2f1bbbb
	github.com/TicketsBot/common v0.0.0-20210727134627-35eb7ed03a44
	github.com/TicketsBot/database v0.0.0-20210808170243-bcb4f117ed18
	github.com/elliotchance/orderedmap v1.2.1
	github.com/gin-gonic/gin v1.7.1
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/jackc/pgx/v4 v4.7.1
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/compress v1.10.10 // indirect
	github.com/rxdn/gdl v0.0.0-20210816102755-d158a03388df
	github.com/sirupsen/logrus v1.5.0
	go.uber.org/atomic v1.6.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	gopkg.in/alexcesaro/statsd.v2 v2.0.0
)
