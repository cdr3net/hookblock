FROM golang:1.14-alpine as builder

RUN mkdir -p /build

WORKDIR /build

COPY go.mod /build/
COPY go.sum /build/

RUN go mod download

COPY . /build/

RUN CGO_ENABLED=0 go build -o hookblock cmd/main.go

FROM alpine:3.12

RUN apk --update --no-cache add ca-certificates

RUN mkdir -p /etc/hookblock

COPY --from=builder /build/hookblock /opt/hookblock

ENTRYPOINT ["/opt/hookblock"]
CMD ["/etc/hookblock/main.hcl"]
