module github.com/TicketsBot/worker

go 1.14

require (
	github.com/TicketsBot/archiverclient v0.0.0-20210220155137-a562b2f1bbbb
	github.com/TicketsBot/common v0.0.0-20210903095620-eb02b87cb4ca
	github.com/TicketsBot/database v0.0.0-20210902205640-76b8973364e8
	github.com/elliotchance/orderedmap v1.2.1
	github.com/gin-gonic/gin v1.7.1
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/go-redis/redis/v8 v8.11.3
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/jackc/pgx/v4 v4.7.1
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/compress v1.10.10 // indirect
	github.com/rxdn/gdl v0.0.0-20210903095530-5a1c35525d2a
	github.com/schollz/progressbar/v3 v3.8.2
	github.com/sirupsen/logrus v1.5.0
	go.uber.org/atomic v1.6.0
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gopkg.in/alexcesaro/statsd.v2 v2.0.0
)
