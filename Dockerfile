FROM golang:1.24

RUN mkdir -p /go/src/github.com/sllt/kite
WORKDIR /go/src/github.com/sllt/kite
COPY . .

RUN go build -ldflags "-linkmode external -extldflags -static" -a examples/http-server/main.go

FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates
COPY --from=0 /go/src/github.com/sllt/kite/main /main
EXPOSE 8000
CMD ["/main"]
