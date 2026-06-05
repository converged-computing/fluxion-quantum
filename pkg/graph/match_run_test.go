//go:build cgo_qrmi

package graph

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/converged-computing/fluxion-quantum/pkg/jobspec"
	"github.com/converged-computing/fluxion-quantum/pkg/quantum"
)

// TestMatchAndRunSampler is the end-to-end path: load the virtual-quantum
// resource graph, ask Fluxion to match-allocate one quantum backend, then submit
// a real SamplerV2 job to the allocated backend through qrmi-go.
//
// Needs the flux-sched and QRMI native libraries (build with -tags cgo_qrmi and
// the merged CGO flags; see `make test-quantum-live`). Skipped unless these are
// set:
//
//	IBM_CLOUD_TOKEN     IBM Cloud IAM API key
//	IBM_CLOUD_CRN       Qiskit Runtime service instance CRN
//
// The backend comes from Fluxion's allocation; QRMI_TEST_BACKEND is only a
// fallback if the allocation carries no name. Job input: $QRMI_SAMPLER_INPUT,
// else pkg/quantum/testdata/sampler_input.json.
func TestMatchAndRunSampler(t *testing.T) {
	if os.Getenv("IBM_CLOUD_TOKEN") == "" || os.Getenv("IBM_CLOUD_CRN") == "" {
		t.Skip("set IBM_CLOUD_TOKEN and IBM_CLOUD_CRN to run the end-to-end test")
	}

	root := repoRoot(t)

	// 1. Build the graph of virtual quantum resources and match-allocate one.
	// Use the JGF allocation format so BackendFromAllocation can parse it.
	g := FluxionGraph{MatchFormat: "jgf"}
	g.Init(filepath.Join(root, "conf", "quantum-virtual.json"), "first", "")

	// Load the jobspec (YAML or JSON) and hand Fluxion a YAML string.
	js, err := jobspec.LoadFile(filepath.Join(root, "queries", "quantum-backend.yaml"))
	if err != nil {
		t.Fatalf("load jobspec: %v", err)
	}
	specYAML, err := js.YAML()
	if err != nil {
		t.Fatalf("render jobspec: %v", err)
	}
	req, err := g.MatchAllocateSpec(specYAML)
	if err != nil {
		t.Fatalf("MatchAllocate: %v", err)
	}

	// 2. Submit to whatever backend Fluxion allocated.
	backend, err := quantum.BackendFromAllocation(req.Allocation, "qpu")
	if err != nil {
		t.Fatalf("BackendFromAllocation: %v", err)
	}
	if backend == "" {
		// Allocation carried no usable name; fall back to the configured one.
		backend = os.Getenv("QRMI_TEST_BACKEND")
	}
	if backend == "" {
		t.Fatal("no backend name in the allocation and QRMI_TEST_BACKEND is unset")
	}
	t.Logf("submitting to Fluxion-allocated backend: %q", backend)

	// 3. Run a real job on it via qrmi-go.
	inputPath := os.Getenv("QRMI_SAMPLER_INPUT")
	if inputPath == "" {
		inputPath = filepath.Join(root, "pkg", "quantum", "testdata", "sampler_input.json")
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("read sampler input %s: %v", inputPath, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	result, err := quantum.RunSampler(ctx, backend, string(input))
	if err != nil {
		t.Fatalf("RunSampler(%s): %v", backend, err)
	}
	if result == "" {
		t.Fatal("empty sampler result")
	}
	t.Logf("sampler result from %s: %d bytes", backend, len(result))
}

// repoRoot returns the repository root, derived from this test file's location
// (pkg/graph/...), so paths work regardless of the test's working directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
