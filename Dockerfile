FROM docker.io/library/golang:1.27.1-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS builder

ENV CGO_ENABLED=0

COPY ./ /src/kopia-healthcheck/
WORKDIR /src/kopia-healthcheck/

RUN <<EOF
  set -ex

  apk --no-cache add \
    git

  install -dm0755 /rootfs/usr/local/bin
  go build \
    -ldflags "-X main.version=$(git describe --tags --always || echo dev)" \
    -mod=readonly \
    -modcacherw \
    -o /rootfs/usr/local/bin/kopia-healthcheck \
    -trimpath
EOF


FROM docker.io/kopia/kopia:0.23.1@sha256:89fd95ee2942880ca00eae964266958a394421ddbdf69bca62e38afc55f5900e

COPY --from=builder /rootfs/ /

ENTRYPOINT ["/usr/local/bin/kopia-healthcheck"]
