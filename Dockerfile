FROM docker.io/golang:alpine as builder

WORKDIR /src

ENV CGO_ENABLED=0

RUN apk add git 
RUN git clone https://github.com/yggdrasil-network/yggdrasil-go.git .
RUN ./build && go build -o /src/genkeys cmd/genkeys/main.go


FROM alpine:3.14

RUN apk --update add redis 

COPY --from=builder /src/yggdrasil /usr/bin/yggdrasil
COPY --from=builder /src/yggdrasilctl /usr/bin/yggdrasilctl
COPY --from=builder /src/genkeys /usr/bin/genkeys


RUN wget https://github.com/threefoldtech/go-rmb/releases/download/v0.1.8/msgbusd.zip && \
    unzip msgbusd.zip &&\
    mv msgbusd /usr/bin/msgbusd

RUN wget https://github.com/threefoldtech/zinit/releases/download/v0.2.5/zinit -O /sbin/zinit \
    && chmod +x /sbin/zinit

RUN wget https://github.com/threefoldtech/tfgridclient_proxy/releases/download/1.1.0/server -O server \
    && chmod +x server \
    && mv server /usr/bin/server

COPY ygg_entrypoint.sh /etc/ygg_entrypoint.sh
RUN chmod +x /etc/ygg_entrypoint.sh
COPY rootfs /

ENV MNEMONICS=""
ENV SERVER_PORT=":443"
ENV EXPLORER="https://graphql.dev.grid.tf/graphql"
ENV SUBSTRATE="wss://tfchain.dev.grid.tf/ws"
ENV REDIS="localhost:6379"

EXPOSE 443
ENTRYPOINT [ "zinit", "init" ]
