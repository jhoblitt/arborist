FROM golang:1.19-alpine as builder

ARG BIN=arborist
RUN apk --update --no-cache add \
    binutils \
    && rm -rf /root/.cache
WORKDIR /go/src/github.com/jhoblitt/arborist
COPY . .
RUN go build && strip "$BIN"

FROM alpine:3
WORKDIR /root/
COPY --from=builder /go/src/github.com/jhoblitt/arborist/$BIN /bin/$BIN
ENTRYPOINT ["/bin/arborist"]
