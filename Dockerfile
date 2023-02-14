FROM docker.io/golang:alpine as builder

ENV CGO_ENABLED=0

WORKDIR /src

RUN apk add git

RUN git clone https://github.com/threefoldtech/tfgridclient_proxy && cd tfgridclient_proxy/cmds/proxy_server &&\
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.GitCommit=$(shell git describe --tags --abbrev=0) -extldflags '-static'"  -o server &&\
    chmod +x server

FROM alpine:3.14

COPY --from=builder /src/tfgridclient_proxy/cmds/proxy_server/server /usr/bin/server

RUN wget https://github.com/threefoldtech/zinit/releases/download/v0.2.10/zinit -O /sbin/zinit \
    && chmod +x /sbin/zinit

COPY rootfs /

ENV SERVER_PORT=":443"
ENV POSTGRES_HOST="postgres"
ENV POSTGRES_PORT="5432"
ENV POSTGRES_DB="name"
ENV POSTGRES_USER="postgres"
ENV POSTGRES_PASSWORD="123"

EXPOSE 443 8051
ENTRYPOINT [ "zinit", "init" ]
