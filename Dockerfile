FROM golang:buster

RUN apt-get update && apt-get upgrade -y && apt-get install -y ca-certificates git zlib1g-dev

RUN mkdir -p /tmp/compile
WORKDIR /tmp/compile

RUN git clone --recurse-submodules https://github.com/TicketsBot/worker.git .
RUN cd locale && git pull origin master
RUN go build -tags=jsoniter cmd/worker/main.go

# Prod container
FROM ubuntu:latest

RUN apt-get update && apt-get upgrade -y && apt-get install -y ca-certificates curl

COPY --from=0 /tmp/compile/locale /srv/worker/locale
COPY --from=0 /tmp/compile/main /srv/worker/worker
RUN chmod +x /srv/worker/worker

RUN useradd -m container
USER container
WORKDIR /srv/worker

CMD ["/srv/worker/worker"]