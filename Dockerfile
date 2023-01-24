FROM docker.io/golang:alpine as builder

ARG YGG_VERSION=v0.4.3

WORKDIR /src

ENV CGO_ENABLED=0

RUN apk add git
RUN git clone --depth 1 --branch $YGG_VERSION https://github.com/yggdrasil-network/yggdrasil-go.git .
RUN ./build && go build -o /src/genkeys cmd/genkeys/main.go


FROM alpine:3.14

RUN apk --update add redis 

COPY --from=builder /src/yggdrasil /usr/bin/yggdrasil
COPY --from=builder /src/yggdrasilctl /usr/bin/yggdrasilctl
COPY --from=builder /src/genkeys /usr/bin/genkeys


RUN wget https://github.com/threefoldtech/go-rmb/releases/download/v0.2.1/msgbusd && \
    mv msgbusd /usr/bin/msgbusd &&\
    chmod +x /usr/bin/msgbusd

RUN wget https://github.com/threefoldtech/zinit/releases/download/v0.2.6/zinit -O /sbin/zinit \
    && chmod +x /sbin/zinit

RUN wget https://github.com/threefoldtech/tfgridclient_proxy/releases/download/v1.6.9/tfgridclient_proxy_1.6.9_linux_amd64.tar.gz \
    && tar -xzf tfgridclient_proxy_1.6.9_linux_amd64.tar.gz \
    && chmod +x server \
    && mv server /usr/bin/server

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
