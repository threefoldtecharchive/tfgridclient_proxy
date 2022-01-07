FROM docker.io/golang:alpine as builder

WORKDIR /src

ENV CGO_ENABLED=0

RUN apk add git
RUN git clone https://github.com/yggdrasil-network/yggdrasil-go.git .
RUN ./build && go build -o /src/genkeys cmd/genkeys/main.go


FROM golang:1.16-alpine as gobuilder

WORKDIR /grid_proxy_server

RUN apk add git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GIT_COMMIT=$(git rev-list -1 HEAD) && \
  go build -ldflags "-X main.GitCommit=$GIT_COMMIT" cmds/proxy_server/main.go

RUN git clone https://github.com/threefoldtech/rmb-go.git
WORKDIR rmb-go/cmds/msgbusd/
RUN go build .


FROM alpine:3.14

RUN apk --update add redis

COPY --from=builder /src/yggdrasil /usr/bin/yggdrasil
COPY --from=builder /src/yggdrasilctl /usr/bin/yggdrasilctl
COPY --from=builder /src/genkeys /usr/bin/genkeys

COPY ygg_entrypoint.sh /etc/ygg_entrypoint.sh
RUN chmod +x /etc/ygg_entrypoint.sh

COPY --from=gobuilder /grid_proxy_server/rmb-go/cmds/msgbusd/msgbusd /usr/bin/msgbus
COPY --from=gobuilder /grid_proxy_server/main /usr/bin/server

COPY rootfs /

RUN wget https://github.com/threefoldtech/zinit/releases/download/v0.1/zinit -O /sbin/zinit \
    && chmod +x /sbin/zinit

ENV TWIN=65
ENV SERVER_IP="0.0.0.0:8080"
ENV EXPLORER_URL="https://graphql.grid.tf/graphql"
ENV REDIS_URL="localhost:6379"
ENV SUBSTRATE_URL="ws://main-tfchainnodepublic-substrate-node"

EXPOSE 8080
ENTRYPOINT [ "zinit", "init" ]
