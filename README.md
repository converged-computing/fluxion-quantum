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
 ./bin/fluxion-quantum --spec ./queries/cores-nested.yaml 
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
./bin/fluxion-quantum --spec ./queries/quantum-cores-nested.yaml
```
```console
This is the fluxion quantum resource matcher
Created fluxion resource graph {96bdaf0}
  Match policy: first
  Load format: JGF (jgf)
  Config file: conf/quantum.json

✨️ Init context complete!
   🌀 Request: ./queries/quantum-cores-nested.yaml
  JobID    : 1
  Reserved : false
  Overhead : 0.000595 seconds
  Time at  : 0
  Allocated :
      ---------------core35[1:x]
      ------------socket1[1:x]
      ---------node0[1:x]
      ---------------core35[1:x]
      ------------socket1[1:x]
      ---------node1[1:x]
      ------rack0[1:s]
      ---tiny0[1:s]


😋 Your resources are satisfied.
```

We would next want to DO something with that allocation! :)

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
