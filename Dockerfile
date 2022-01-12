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
RUN GIT_COMMIT=$(git describe --tags --abbrev=0) && \
    cd cmds/proxy_server && CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.GitCommit=$GIT_COMMIT -extldflags '-static'"  -o proxy_server
  
RUN git clone https://github.com/threefoldtech/go-rmb.git
RUN cd go-rmb/cmds/msgbusd && CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -extldflags '-static'"  -o msgbusd


FROM alpine:3.14

RUN apk --update add redis 

COPY --from=builder /src/yggdrasil /usr/bin/yggdrasil
COPY --from=builder /src/yggdrasilctl /usr/bin/yggdrasilctl
COPY --from=builder /src/genkeys /usr/bin/genkeys

COPY ygg_entrypoint.sh /etc/ygg_entrypoint.sh
RUN chmod +x /etc/ygg_entrypoint.sh

COPY --from=gobuilder /grid_proxy_server/go-rmb/cmds/msgbusd/msgbusd /usr/bin/msgbusd
COPY --from=gobuilder /grid_proxy_server/cmds/proxy_server/proxy_server /usr/bin/server


COPY rootfs /

RUN wget https://github.com/threefoldtech/zinit/releases/download/v0.2.5/zinit -O /sbin/zinit \
    && chmod +x /sbin/zinit

ENV TWIN=60
ENV SERVER_PORT=":8080"
ENV EXPLORER="https://graphql.dev.grid.tf/graphql"
ENV SUBSTRATE="wss://tfchain.dev.grid.tf/ws"
ENV REDIS="localhost:6379"
# ENV DOMAIN="gridproxy.3botmain.grid.tf"
# ENV EMAIL="gridproxy@gmail.com"
# ENV CA="https://acme-v02.api.letsencrypt.org/directory"

EXPOSE 8080
ENTRYPOINT [ "zinit", "init" ]
