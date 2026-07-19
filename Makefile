default:

build-local:
	go build \
		-buildvcs=false \
		-ldflags "-s -w -buildid= -X main.version=$(PRODUCT_VERSION)" \
		-mod=readonly \
		-trimpath

trivy:
	trivy fs . \
		--dependency-tree \
		--exit-code 1 \
		--format table \
		--ignore-unfixed \
		--quiet \
		--scanners license,secret,vuln \
		--severity HIGH,CRITICAL
