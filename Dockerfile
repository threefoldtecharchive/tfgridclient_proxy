FROM docker.io/golang:alpine as builder

ENV CGO_ENABLED=0

WORKDIR /src

ADD . /src

ARG version

RUN cd /src/cmds/proxy_server &&\
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.GitCommit=${version} -extldflags '-static'"  -o gridrest &&\
    chmod +x gridrest

FROM alpine:3.14

COPY --from=builder /src/cmds/proxy_server/gridrest /usr/bin/gridrest

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
