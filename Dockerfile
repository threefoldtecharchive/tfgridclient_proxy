FROM docker.io/golang:alpine as builder

WORKDIR /src

ENV CGO_ENABLED=0

RUN apk add git 
RUN git clone https://github.com/yggdrasil-network/yggdrasil-go.git .
RUN ./build && go build -o /src/genkeys cmd/genkeys/main.go


FROM golang:1.16-alpine

WORKDIR /grid_proxy_server

RUN apk --update add redis 

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server cmds/proxy_server/main.go

RUN wget https://github.com/threefoldtech/zinit/releases/download/v0.1/zinit -O /sbin/zinit \
    && chmod +x /sbin/zinit

# TODO: copy msgbus from release, currently locally
COPY msgbus /usr/bin/msgbus

COPY rootfs /

EXPOSE 8080

COPY --from=builder /src/yggdrasil /usr/bin/yggdrasil
COPY --from=builder /src/yggdrasilctl /usr/bin/yggdrasilctl
COPY --from=builder /src/genkeys /usr/bin/genkeys

RUN chmod +x ygg_entrypoint.sh

ENTRYPOINT [ "zinit", "init" ]