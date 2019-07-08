# compute source code references
GIT_COMMIT = $(shell git rev-parse --short HEAD || echo 'none')
_TMP_STR = $(shell git show-ref --abbrev --head | grep "^$(GIT_COMMIT) " | grep -v HEAD | head -1 | awk '{print $$NF}')
GIT_REF ?= $(shell [ "$(_TMP_STR)" = "" ] && echo "HEAD" || echo $(_TMP_STR))
GIT_DIRTY = $(shell git diff --quiet || echo '-dirty')
CODE_VERSION = "$(GIT_REF)@$(GIT_COMMIT)$(GIT_DIRTY)"

# Image URL to use all building/pushing image targets
IMG ?= federatorai-emulator:latest

SRC_DIR = $(shell pwd)
INSTALL_ROOT = $(SRC_DIR)/install_root
PRODUCT_ROOT = /opt/alameda/federatorai-emulator
DEST_PREFIX = $(INSTALL_ROOT)$(PRODUCT_ROOT)
######################################################################

.PHONY: all test transmitter
all: test federatorai-emulator

# Run tests
test: generate fmt vet
	go test ./pkg/... -coverprofile cover.out

# Build transmitter binary
federatorai-emulator: generate fmt vet binaries 	# go build -ldflags "-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'" -o transmitter/transmitter github.com/containers-ai/federatorai-emulator/emulator.go
	#go build -o emulator/emulator github.com/containers-ai/federatorai-emulator/emulator.go

binaries:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
	  -ldflags "-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'" \
	  -a -o ./emulator/emulator github.com/containers-ai/federatorai-emulator

install_dir:
	mkdir -pv $(INSTALL_ROOT)/etc/logrotate.d $(DEST_PREFIX)/etc/emulator $(DEST_PREFIX)/bin

install: install_dir
	cp -fv etc/emulator.toml $(DEST_PREFIX)/etc/emulator/emulator.toml
	cp -fv etc/metric_cpu.csv $(DEST_PREFIX)/etc/emulator/metric_cpu.csv
	cp -fv etc/metric_memory.csv $(DEST_PREFIX)/etc/emulator/metric_memory.csv
	cp -fv etc/node.json $(DEST_PREFIX)/etc/emulator/node.json
	cp -fv etc/pod.json $(DEST_PREFIX)/etc/emulator/pod.json
	cp -fv emulator/emulator $(DEST_PREFIX)/bin/
	cp -fv $(SRC_DIR)/xray.sh $(DEST_PREFIX)/bin/ && chmod 755 $(DEST_PREFIX)/bin/xray.sh
	# generate version.txt
	echo "CODE_VERSION=$(CODE_VERSION)" >> $(DEST_PREFIX)/etc/version.txt
	# logrotate.conf
	cp -fv $(SRC_DIR)/logrotate.conf $(DEST_PREFIX)/etc/
	ln -sfv $(PRODUCT_ROOT)/etc/logrotate.conf $(INSTALL_ROOT)/etc/logrotate.d/alameda-emulator
	ln -sfv $(PRODUCT_ROOT)/etc/emulator $(INSTALL_ROOT)/etc/emulator
	# init.sh
	cp -fv $(SRC_DIR)/init.sh $(INSTALL_ROOT)/init.sh && chmod 755 $(INSTALL_ROOT)/init.sh
	cd $(INSTALL_ROOT); tar -czvf $(SRC_DIR)/install_root.tgz .; cd -

clean:
	rm -fv build/build-image/bin/emulator install_root.tgz *~
	rm -rf emulator

clobber: clean
	rm -rf install_root

build: binaries install

.PHONY: run binaries

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./emulator.go run

.PHONY: fmt vet generate docker-build docker-push

# Run go fmt against code
fmt:
	go fmt ./pkg/...

# Run go vet against code
vet:
	go vet ./pkg/...

# Generate code
generate:
	go generate ./cmd/...

## docker-build: Build the docker image.
docker-build-alpine:
	docker build -t ${IMG} -f Dockerfile .

docker-build-ubi:
	docker build -t ${IMG} -f Dockerfile.ubi .

docker-build: docker-build-ubi
