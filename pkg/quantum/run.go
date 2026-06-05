package quantum

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/converged-computing/qrmi-go/pkg/qrmi"
)

// Public IBM Quantum Platform defaults for the Qiskit Runtime Service.
const (
	DefaultQRSEndpoint = "https://quantum.cloud.ibm.com/api/v1"
	DefaultIAMEndpoint = "https://iam.cloud.ibm.com"
)

// ExecutionMode selects how a primitive runs.
type ExecutionMode string

const (
	// JobMode submits a single job with no session. Works on every plan,
	// including the IBM open plan. This is the default.
	JobMode ExecutionMode = "job"
	// SessionMode opens a Qiskit Runtime session (acquire/release). Holds the
	// backend for the call, but requires a paid plan — the open plan rejects
	// session creation (HTTP 400, error 1352).
	SessionMode ExecutionMode = "session"
)

// ExecutionModeFromEnv reads QRMI_EXECUTION_MODE ("job" or "session") and
// defaults to JobMode.
func ExecutionModeFromEnv() ExecutionMode {
	if strings.ToLower(os.Getenv("QRMI_EXECUTION_MODE")) == "session" {
		return SessionMode
	}
	return JobMode
}

// allocation is the subset of a Fluxion allocation graph we need: the metadata
// type and name of each allocated vertex.
type allocation struct {
	Graph struct {
		Nodes []struct {
			Metadata struct {
				Type string `json:"type"`
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"nodes"`
	} `json:"graph"`
}

// BackendFromAllocation returns the name of the first vertex of vertexType
// (e.g. "qpu") in a Fluxion allocation graph. In the virtual-quantum resource
// graph each qpu vertex is named after a QRMI backend (e.g. "ibm_fez"), so this
// is how we turn "Fluxion scheduled a quantum resource" into "which backend do
// I submit to."
func BackendFromAllocation(alloc string, vertexType string) (string, error) {
	var a allocation
	if err := json.Unmarshal([]byte(alloc), &a); err != nil {
		return "", fmt.Errorf("parse allocation: %w", err)
	}
	for _, n := range a.Graph.Nodes {
		if n.Metadata.Type == vertexType {
			return n.Metadata.Name, nil
		}
	}
	return "", fmt.Errorf("no %q vertex found in allocation", vertexType)
}

// ConfigureQRSEnv sets the per-backend environment variables QRMI reads for the
// Qiskit Runtime Service, derived from a few simple inputs. QRMI expects each
// var prefixed by the backend name (e.g. ibm_fez_QRMI_IBM_QRS_IAM_APIKEY) — the
// same variables the Slurm SPANK plugin injects from qrmi_config.json.
//
//	IBM_CLOUD_TOKEN   -> <backend>_QRMI_IBM_QRS_IAM_APIKEY   (required)
//	IBM_CLOUD_CRN     -> <backend>_QRMI_IBM_QRS_SERVICE_CRN  (required)
//	QRMI_QRS_ENDPOINT -> <backend>_QRMI_IBM_QRS_ENDPOINT     (default public URL)
//	QRMI_QRS_IAM_ENDPOINT -> <backend>_QRMI_IBM_QRS_IAM_ENDPOINT (default public URL)
//
// Optional session knobs (only set if the corresponding env var is present, so
// QRMI's defaults otherwise apply). These matter because acquire() creates a
// Qiskit Runtime session, and an instance's plan limits the allowed mode and
// max TTL — a dedicated 8h session is rejected (HTTP 400) on plans that don't
// permit it:
//
//	QRMI_QRS_SESSION_MODE    -> <backend>_QRMI_IBM_QRS_SESSION_MODE    ("batch" or "dedicated")
//	QRMI_QRS_SESSION_MAX_TTL -> <backend>_QRMI_IBM_QRS_SESSION_MAX_TTL (seconds)
//
// With cgo enabled, Go propagates these to the C environment QRMI reads.
func ConfigureQRSEnv(backend string) error {
	apiKey := os.Getenv("IBM_CLOUD_TOKEN")
	crn := os.Getenv("IBM_CLOUD_CRN")
	if backend == "" || apiKey == "" || crn == "" {
		return fmt.Errorf("need a backend plus IBM_CLOUD_TOKEN and IBM_CLOUD_CRN in the environment")
	}
	vars := map[string]string{
		backend + "_QRMI_IBM_QRS_ENDPOINT":     getenvDefault("QRMI_QRS_ENDPOINT", DefaultQRSEndpoint),
		backend + "_QRMI_IBM_QRS_IAM_ENDPOINT": getenvDefault("QRMI_QRS_IAM_ENDPOINT", DefaultIAMEndpoint),
		backend + "_QRMI_IBM_QRS_IAM_APIKEY":   apiKey,
		backend + "_QRMI_IBM_QRS_SERVICE_CRN":  crn,
	}
	// Pass through optional session settings only when provided.
	for src, suffix := range map[string]string{
		"QRMI_QRS_SESSION_MODE":    "_QRMI_IBM_QRS_SESSION_MODE",
		"QRMI_QRS_SESSION_MAX_TTL": "_QRMI_IBM_QRS_SESSION_MAX_TTL",
	} {
		if v := os.Getenv(src); v != "" {
			vars[backend+suffix] = v
		}
	}
	for k, v := range vars {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("setenv %s: %w", k, err)
		}
	}
	return nil
}

// RunSampler submits a SamplerV2 primitive to backend via qrmi-go and returns
// the serialized result. samplerParamsJSON is the params-only SamplerV2 input
// (the {"pubs": ..., "version": 2, ...} document, e.g. the
// sampler_input_<backend>_params_only.json produced by QRMI's
// gen_sampler_inputs.py).
//
// The execution mode comes from QRMI_EXECUTION_MODE (see ExecutionModeFromEnv);
// it defaults to job mode, which is the only mode the IBM open plan allows.
func RunSampler(ctx context.Context, backend, samplerParamsJSON string) (string, error) {
	return RunSamplerMode(ctx, backend, samplerParamsJSON, ExecutionModeFromEnv())
}

// RunSamplerMode runs the sampler in an explicit execution mode.
//
//   - JobMode submits a single job with no session. Allowed on every plan,
//     including open. This is the default.
//   - SessionMode opens a Qiskit Runtime session (acquire/release), which keeps
//     the backend for the call but requires a paid plan; on the open plan IBM
//     rejects session creation with HTTP 400 (error 1352).
func RunSamplerMode(ctx context.Context, backend, samplerParamsJSON string, mode ExecutionMode) (string, error) {
	if err := ConfigureQRSEnv(backend); err != nil {
		return "", err
	}
	r, err := qrmi.New(backend, qrmi.QiskitRuntimeService)
	if err != nil {
		return "", fmt.Errorf("qrmi.New(%s): %w", backend, err)
	}
	defer r.Close()

	pub := qrmi.QiskitPrimitive{Input: samplerParamsJSON, ProgramID: "sampler"}
	if mode == SessionMode {
		return qrmi.RunPrimitive(ctx, r, pub, qrmi.RunOptions{PollInterval: 3 * time.Second})
	}
	return runJob(ctx, r, pub, 3*time.Second)
}

// runJob submits a single job (no acquire/release) and polls to completion.
// This is the "job" execution mode the open plan permits.
func runJob(ctx context.Context, r qrmi.Resource, pub qrmi.QiskitPrimitive, poll time.Duration) (string, error) {
	taskID, err := r.TaskStart(pub)
	if err != nil {
		return "", fmt.Errorf("qrmi: task_start: %w", err)
	}

	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	for {
		status, err := r.TaskStatus(taskID)
		if err != nil {
			return "", fmt.Errorf("qrmi: task_status: %w", err)
		}
		switch status {
		case qrmi.Completed:
			return r.TaskResult(taskID)
		case qrmi.Failed, qrmi.Cancelled:
			logs, _ := r.TaskLogs(taskID) // best effort
			return "", fmt.Errorf("qrmi: task %s ended %s: %s", taskID, status, logs)
		}
		select {
		case <-ctx.Done():
			_ = r.TaskStop(taskID) // best effort
			return "", ctx.Err()
		case <-ticker.C:
		}
	}
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
