# This assumes a build in the .devcontainer Dockerfile environment
FLUX_SCHED_ROOT ?= /opt/flux-sched
# We keep this as linux to match your working command
COMMONENVVAR = GOOS=linux

# 1. CGO_CFLAGS: We point to /opt/flux-sched so the compiler finds "resource/reapi/..."
# 2. CGO_LDFLAGS: We use -lresource (NOT -lfluxion-resource) to match your working command.
# 3. We remove -L/usr/lib to let the system find czmq in its default path.
BUILDENVVAR = CGO_CFLAGS="-I${FLUX_SCHED_ROOT}" \
              CGO_LDFLAGS="-L${FLUX_SCHED_ROOT}/resource \
                           -L${FLUX_SCHED_ROOT}/resource/libjobspec \
                           -L${FLUX_SCHED_ROOT}/resource/reapi/bindings \
                           -lresource \
                           -ljobspec_conv \
                           -lreapi_cli \
                           -lflux-idset \
                           -lstdc++ \
                           -lczmq \
                           -ljansson \
                           -lhwloc \
                           -lboost_system \
                           -lflux-hostlist \
                           -lboost_graph \
                           -lyaml-cpp"

.PHONY: all
all: flex

.PHONY: flex
flex: 
	go mod tidy
	mkdir -p ./bin
	export $(COMMONENVVAR); export $(BUILDENVVAR); go build -ldflags '-w' -o bin/fluxion-quantum src/cmd/main.go

.PHONY: clean
clean:
	rm -rf ./bin/fluxion-quantum