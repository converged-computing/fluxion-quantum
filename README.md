# Fluxion Quantum

> Testing scheduling with fluxion bindings and quantum resources.

The Flux Framework "flux-sched" or fluxion project provides modular bindings in different languages for intelligent,
graph-based scheduling. The project here demonstrates using fluxion to:

1. Generate a [graph schema](conf/quantum.graphml) (graphml) that describes nodes and edges of quantum resources
2. Create an example [query](query.yaml) that a customer wants
3. Use the tool to determine if resource request can be satisfied.

This is a simple example intended for learning.

## Concepts

From the above, the following definitions might be useful.

 - **[Flux Framework](https://flux-framework.org)**: a modular framework for putting together a workload manager. It is traditionally for HPC, but components have been used in other places (e.g., here, Kubernetes, etc). It is analogous to Kubernetes in that it is modular and used for running batch workloads.
 - **[fluxion](fluxion)**: refers to [flux-framework/flux-sched](https://github.com/flux-framework/flux-sched) and is the scheduler component or module of Flux Framework. There are bindings in several languages, and specifically the Go bindings (server at [flux-framework/flux-k8s](https://github.com/flux-framework/flux-k8s)) assemble into the project "fluence."

## Usage

### Build

This demonstrates how to build the bindings. You will need to be in the VSCode developer container environment, or produce the same
on your host. 

```bash
make
```
```console
go mod tidy
mkdir -p ./bin
export GOOS=linux; export CGO_CFLAGS="-I/opt/flux-sched" CGO_LDFLAGS="-L/opt/flux-sched/resource -L/opt/flux-sched/resource/libjobspec -L/opt/flux-sched/resource/reapi/bindings -lresource -ljobspec_conv -lreapi_cli -lflux-idset -lstdc++ -lczmq -ljansson -lhwloc -lboost_system -lflux-hostlist -lboost_graph -lyaml-cpp"; go build -ldflags '-w' -o bin/fluxion-quantum src/cmd/main.go
```

The output is generated in bin:

```bash
$ ls bin/
fluxion-quantum
```

### Run

```bash
export LD_LIBRARY_PATH=/usr/lib:/opt/flux-sched/resource:/opt/flux-sched/resource/reapi/bindings:/opt/flux-sched/resource/libjobspec
```

Note that the resource graph is a modified [tiny.json](https://raw.githubusercontent.com/flux-framework/flux-sched/f0e815f1f354a58b2737d45144087f5ed4564140/t/data/resource/jgfs/tiny.json) with quantum devices added on the same level as racks. The idea is that we would schedule the resources at the same time but as separate entities. This might be like saying "quantum resources are *considered* part of the cluster, but are distinct from resources we find under a rack."

#### Example 1: Satisfied Cores

```bash
./bin/fluxion-quantum --spec ./queries/cores-nested.yaml
```
```console
This is the fluxion quantum resource matcher
Created fluxion resource graph {1d538af0}
  Match policy: first
  Load format: JGF (jgf)
  Config file: conf/quantum.json

✨️ Init context complete!
   🌀 Request: ./queries/cores-nested.yaml
  JobID    : 1
  Reserved : false
  Overhead : 0.000356 seconds
  Time at  : 0
  Allocated :
      ---------------core35[1:x]
      ------------socket1[1:x]
      ---------node1[1:x]
      ------rack0[1:x]
      ---tiny0[1:s]


😋 Your resources are satisfied.
```

You can do an a-la-carte, satisfy only request (no allocate):

```bash
./bin/fluxion-quantum --spec ./queries/cores-nested.yaml  --satisfy
```

#### Example 2: Unsatisfied Cores

```bash
./bin/fluxion-quantum --spec ./queries/unsatisfied-cores.yaml 
```

#### Example 3: Quantum Resources and Cores

This models quantum resources `qdevice` alongside a rack.

```bash
 ./bin/fluxion-quantum --spec ./queries/quantum-rack-nested.yaml 
```
```console
This is the fluxion quantum resource matcher
Created fluxion resource graph
  Match policy: first
  Load format: JGF (jgf)
  Match format: simple
  Config file: conf/quantum.json

✨️ Init context complete!
   🌀 Request (file): ./queries/quantum-rack-nested.yaml
  JobID    : 1
  Reserved : false
  Overhead : 0.000613 seconds
  Time at  : 0
  Allocated :
      ---------------core35[1:x]
      ------------socket1[1:x]
      ---------node0[1:x]
      ---------------core35[1:x]
      ------------socket1[1:x]
      ---------node1[1:x]
      ------rack0[1:x]
      ------------q0[1:x]
      ------------q1[1:x]
      ------------c0_1[1:x]
      ---------qpu0[1:x]
      ------qdevice0[1:x]
      ---tiny0[1:s]

😋 Your resources are satisfied.
```

We would next want to DO something with that allocation! :)

### Quantum

Test quantum

```bash
make test-quantum
```

And then build and run with `fluxion-quantum`

```bash
./bin/fluxion-quantum -conf conf/quantum-virtual.json -spec queries/quantum-virtual.yaml
```
```console
This is the fluxion quantum resource matcher
Created fluxion resource graph {15eceaf0}
  Match policy: first
  Load format: JGF (jgf)
  Config file: conf/quantum-virtual.json

✨️ Init context complete!
   🌀 Request: queries/quantum-virtual.yaml
  JobID    : 1
  Reserved : false
  Overhead : 0.000586 seconds
  Time at  : 0
  Allocated :
      ---------ibm_marrakesh[1:x]
      ------qgateway0[1:s]
      ---tiny0[1:s]


😋 Your resources are satisfied.
```

And add `--satisfy` to check that. This isn't running a real job, it is just querying the resource graph. Let's do that next. To login I usually do:

```bash
curl -fsSL https://clis.cloud.ibm.com/install/linux | sudo sh
ibmcloud login --apikey <key>
# 12 for us-east
```

```bash
export IBM_CLOUD_TOKEN=<key>
export QRMI_TEST_BACKEND=ibm_fez 
export QRMI_TEST_TYPE=qiskit-runtime-service
export IBM_CLOUD_CRN=$(ibmcloud resource service-instances --service-name quantum-computing --output json | jq -r '.[] | {name: .name, crn: .crn}' | jq -r .crn)
```

And now test with the live example.

```bash
export QRMI_QRS_SESSION_MAX_TTL=600     # 10m
make test-quantum-live
```
```console
CGO_ENABLED=1 CGO_CFLAGS="-I/opt/flux-sched -I/usr/local/include" CGO_LDFLAGS="-L/opt/flux-sched/resource -L/opt/flux-sched/resource/libjobspec -L/opt/flux-sched/resource/reapi/bindings -L/usr/local/lib -lresource -ljobspec_conv -lreapi_cli -lflux-idset -lstdc++ -lczmq -ljansson -lhwloc -lboost_system -lflux-hostlist -lboost_graph -lyaml-cpp" \
  go test -tags cgo_qrmi -v -run 'TestMatchAndRunSampler|TestRunSamplerLive' ./pkg/...
=== RUN   TestMatchAndRunSampler
Created fluxion resource graph
  Match policy: first
  Load format: JGF (jgf)
  Match format: jgf
  Config file: /workspaces/fluxion-quantum/conf/quantum-virtual.json

✨️ Init context complete!
  JobID    : 1
  Reserved : false
  Overhead : 0.000344 seconds
  Time at  : 0
  Allocated :
{"graph": {"nodes": [{"id": "3", "metadata": {"type": "qpu", "name": "ibm_marrakesh", "id": 1, "rank": -1, "exclusive": true, "paths": {"containment": "/tiny0/qgateway0/ibm_marrakesh"}}}, {"id": "1", "metadata": {"type": "qgateway", "id": 0, "rank": -1, "paths": {"containment": "/tiny0/qgateway0"}}}, {"id": "0", "metadata": {"type": "cluster", "basename": "tiny", "id": 0, "rank": -1, "paths": {"containment": "/tiny0"}}}], "edges": [{"source": "1", "target": "3"}, {"source": "0", "target": "1"}]}}

    match_run_test.go:69: submitting to Fluxion-allocated backend: "ibm_marrakesh"
    match_run_test.go:91: sampler result from ibm_marrakesh: 2070 bytes
--- PASS: TestMatchAndRunSampler (11.18s)
PASS
ok      github.com/converged-computing/fluxion-quantum/pkg/graph        (cached)
testing: warning: no tests to run
PASS
ok      github.com/converged-computing/fluxion-quantum/pkg/jobspec      (cached) [no tests to run]
testing: warning: no tests to run
PASS
ok      github.com/converged-computing/fluxion-quantum/pkg/quantum      0.005s [no tests to run]
```

In the above, what happens? This is what I think. We got a 2070 byte sampler result back from `ibm_marrakesh` in ~11 seconds. This includes:

- Fluxion loading the virtual resource graph, matching the jobspec, and allocating a qpu vertex. 
- qmri-go submitted the SamplerV2 job to IBM and got back the result.

E.g., `graph → match → allocate → QRMI job → result`

And next we can actually integrate this into Kubernetes, paried with traditional resources, and likely a custom scheduler plugin. Stay tuned!

### Multiple Vendors

```bash
./bin/fluxion-quantum --conf conf/quantum-vendors.json --spec queries/device-region-east.yaml --retval
./bin/fluxion-quantum --conf conf/quantum-vendors.json --spec queries/device-region-missing.yaml --retval
./bin/fluxion-quantum --conf conf/quantum-vendors.json --spec queries/hybrid-constraint.yaml --retval
./bin/fluxion-quantum --conf conf/quantum-vendors.json --spec queries/no-constraint-qpu.yaml --retval
./bin/fluxion-quantum --conf conf/quantum-vendors.json --spec queries/unsatisfied-cores.yaml --retval
./bin/fluxion-quantum --conf conf/quantum-vendors.json --spec queries/virtual-false.yaml --retval
./bin/fluxion-quantum --conf conf/quantum-vendors.json --spec queries/virtual-true.yaml --retval
```

### Debugging

When I first tested a session I was getting 400 errors, and I needed more detail. I ran:

```bash
bash ./test/debug_request.sh
```

And it told me exactly what I needed to know - my account did not have scope to request a session!

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
