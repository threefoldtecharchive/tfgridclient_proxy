FROM docker.io/golang:alpine as builder

WORKDIR /src

ENV CGO_ENABLED=0

RUN apk add git 
RUN git clone https://github.com/yggdrasil-network/yggdrasil-go.git .
RUN ./build && go build -o /src/genkeys cmd/genkeys/main.go


FROM golang:1.17-alpine as gobuilder

WORKDIR /grid_proxy_server

RUN apk add git 
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 && \
    GIT_COMMIT=$(git describe --tags --abbrev=0) && \
    go build -ldflags "-X main.GitCommit=$GIT_COMMIT -extldflags '-static'"  cmds/proxy_server/main.go
  
RUN git clone https://github.com/threefoldtech/rmb-go.git
RUN cd rmb-go && go build cmds/msgbus/main.go


FROM alpine:3.14

RUN apk --update add redis 

COPY --from=builder /src/yggdrasil /usr/bin/yggdrasil
COPY --from=builder /src/yggdrasilctl /usr/bin/yggdrasilctl
COPY --from=builder /src/genkeys /usr/bin/genkeys

COPY ygg_entrypoint.sh /etc/ygg_entrypoint.sh
RUN chmod +x /etc/ygg_entrypoint.sh

COPY --from=gobuilder /grid_proxy_server/rmb-go/main /usr/bin/msgbus
COPY --from=gobuilder /grid_proxy_server/main /usr/bin/server

COPY rootfs /

RUN wget https://github.com/threefoldtech/zinit/releases/download/v0.2.5/zinit -O /sbin/zinit \
    && chmod +x /sbin/zinit

ENV TWIN=60
ENV SERVER_PORT=":443"
ENV EXPLORER="https://graphql.dev.grid.tf/graphql"
ENV REDIS_URL="localhost:6379"
ENV DOMAIN="gridproxy.3botmain.grid.tf"
ENV EMAIL="gridproxy@gmail.com"
ENV CA="https://acme-v02.api.letsencrypt.org/directory"
ENV SUBSTRATE="wss://tfchain.dev.grid.tf/ws"

EXPOSE 443
ENTRYPOINT [ "zinit", "init" ]
