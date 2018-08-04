from golang:1.10.3 as build
workdir /go/src/go.jonnrb.io/wifitriggers
add . .
run go get ./cmd/autowyze

from gcr.io/distroless/base
copy --from=build /go/bin/autowyze /
entrypoint ["/autowyze"]
