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
all: fluxion

.PHONY: fluxion
fluxion: 
	go mod tidy
	mkdir -p ./bin
	export $(COMMONENVVAR); export $(BUILDENVVAR); go build -ldflags '-w' -o bin/fluxion-quantum cmd/main.go

.PHONY: clean
clean:
	rm -rf ./bin/fluxion-quantum

# --- Quantum job execution (qrmi-go) -------------------------------------
# Builds with BOTH cgo dependencies: flux-sched (the matcher) and libqrmi (the
# job runner). QRMI is installed by the .devcontainer into $(QRMI_PREFIX).
QRMI_PREFIX ?= /usr/local
QUANTUM_CGO_CFLAGS = -I$(FLUX_SCHED_ROOT) -I$(QRMI_PREFIX)/include
QUANTUM_CGO_LDFLAGS = -L$(FLUX_SCHED_ROOT)/resource \
                      -L$(FLUX_SCHED_ROOT)/resource/libjobspec \
                      -L$(FLUX_SCHED_ROOT)/resource/reapi/bindings \
                      -L$(QRMI_PREFIX)/lib \
                      -lresource -ljobspec_conv -lreapi_cli -lflux-idset \
                      -lstdc++ -lczmq -ljansson -lhwloc -lboost_system \
                      -lflux-hostlist -lboost_graph -lyaml-cpp
 
# Pure-Go unit tests for the quantum runner (no native libs, no API calls). 
.PHONY: test-quantum
test-quantum:
	go test ./pkg/quantum/... ./pkg/jobspec/...

# End-to-end: Fluxion matches a virtual quantum resource, qrmi-go runs a job.
# Needs flux-sched + libqrmi and (to actually run) QRMI_TEST_BACKEND + creds.
.PHONY: test-quantum-live
test-quantum-live:
	export $(COMMONENVVAR); \
	CGO_ENABLED=1 CGO_CFLAGS="$(QUANTUM_CGO_CFLAGS)" CGO_LDFLAGS="$(QUANTUM_CGO_LDFLAGS)" \
	  go test --count=1 -tags cgo_qrmi -v -run 'TestMatchAndRunSampler|TestRunSamplerLive' ./pkg/...