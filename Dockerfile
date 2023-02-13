FROM docker.io/golang:alpine as builder

ARG YGG_VERSION=v0.4.7
ENV CGO_ENABLED=0

WORKDIR /src

RUN apk add git

RUN git clone --depth 1 --branch $YGG_VERSION https://github.com/yggdrasil-network/yggdrasil-go.git .
RUN ./build && go build -o /src/genkeys cmd/genkeys/main.go

RUN git clone https://github.com/threefoldtech/tfgridclient_proxy && cd tfgridclient_proxy/cmds/proxy_server &&\
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.GitCommit=$(shell git describe --tags --abbrev=0) -extldflags '-static'"  -o server &&\
    chmod +x server

RUN git clone https://github.com/threefoldtech/rmb_go && cd rmb_go/cmds/msgbusd &&\
    CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s -extldflags "-static"' -o msgbusd &&\
    chmod +x msgbusd

FROM alpine:3.14

RUN apk --update add redis 

COPY --from=builder /src/yggdrasil /usr/bin/yggdrasil
COPY --from=builder /src/yggdrasilctl /usr/bin/yggdrasilctl
COPY --from=builder /src/genkeys /usr/bin/genkeys
COPY --from=builder /src/tfgridclient_proxy/cmds/proxy_server/server /usr/bin/server
COPY --from=builder /src/rmb_go/cmds/msgbusd/msgbusd /usr/bin/msgbusd

RUN wget https://github.com/threefoldtech/zinit/releases/download/v0.2.10/zinit -O /sbin/zinit \
    && chmod +x /sbin/zinit

COPY ygg_entrypoint.sh /etc/ygg_entrypoint.sh
RUN chmod +x /etc/ygg_entrypoint.sh
COPY rootfs /

ENV MNEMONICS=""
ENV SERVER_PORT=":443"
ENV POSTGRES_HOST="postgres"
ENV POSTGRES_PORT="5432"
ENV POSTGRES_DB="name"
ENV POSTGRES_USER="postgres"
ENV POSTGRES_PASSWORD="123"
ENV SUBSTRATE="wss://tfchain.dev.grid.tf/ws"
ENV REDIS="tcp://127.0.0.1:6379"
ENV RMB_TIMEOUT="30"

EXPOSE 443 8051
ENTRYPOINT [ "zinit", "init" ]
