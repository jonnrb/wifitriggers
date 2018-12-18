from quay.io/jonnrb/go as build
add . /src
run cd /src && CGO_ENABLED=0 go get ./cmd/autowyze

from gcr.io/distroless/static
copy --from=build /go/bin/autowyze /
entrypoint ["/autowyze"]
