FROM docker.io/library/golang:1.27.1-alpine@sha256:4cb7ac979db5fcc41cae44b2227ba5ab8a51e8807f40d9ba4dee20a0ad960b5b AS builder

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
