FROM docker.io/library/golang:1.27.4-alpine@sha256:0a5f7fe796599ff487605987432e7619ae45beeb2fba4c2ce0b907879932aebf AS builder

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
