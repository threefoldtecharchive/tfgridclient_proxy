FROM golang:1.16-alpine

WORKDIR /grid_proxy_server

RUN echo vm.overcommit_memory = 1 >> /etc/sysctl.conf
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

ENTRYPOINT [ "zinit", "init" ]