FROM golang:1.16-alpine

WORKDIR /grid_proxy_server

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server cmds/proxy_server/main.go

EXPOSE 8080

CMD [ "./server" ]