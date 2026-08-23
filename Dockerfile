FROM docker.io/library/golang:1.27.0-alpine@sha256:4c9fe60190a2a3350ddc51de80d0224b8a6698d12bdfc999fee45ea9d6c46dbc AS builder

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
