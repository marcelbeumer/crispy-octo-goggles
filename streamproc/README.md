# StreamProc

**WARNING: WORK IN PROGRESS, SEE [TODO](#TODO)**

Basic stream processing exercise and [Kubernetes](https://kubernetes.io) local development setup.

- [Event producer](./services/event-producer) that generates events in JSON format of `{ amount: number }`.
- [Event API service](./services/event-api) that takes events and writes them to [Kafka](https://kafka.apache.org).
- [Consumer service "high"](./services/consumer-high) that consumes from Kafka and writes amounts >5 to [TimescaleDB](https://www.timescale.com).
- [Consumer service "low"](./services/consumer-low) that consumes from Kafka and writes amounts <=5 to [InfluxDB](https://www.influxdata.com).
- [Aggregator service](./services/aggregator) providing both high and low time series data.
- [Web UI](./services/web-ui) that renders line chart for high and low data.

## System requirements

- [golang](https://go.dev) >=1.18
- [docker](https://www.docker.com)
- [k3d](https://k3d.io) or [kind](https://kind.sigs.k8s.io)
- [helm](https://helm.sh)

## Setup

- Build local docker images using `./scripts/build_images.sh`.
- Create a k3d cluster with `./scripts/create_k3d_cluster.sh`.
- Push local images to k3d registry with `./scripts/push_images.sh`.
- Install helm chart with `helm install streamproc ./helm_chart`.

## Local development

Pushing new docker images on code changes can be slow and tedious. A few strategies for local development with a (local) k8s cluster:

- Use [telepresence](https://www.telepresence.io) to intercept traffic from/to a service in the cluster to the local machine.
- Use `kubectl port-forward` to expose services from the cluster to the local machine, using those in local processes / services.
- Use combination of file mounts and file watchers to initiate rebuilds on code change and automatically restarting services.

### Telepresence

TODO

### Kubectl port-forward

TODO

### File mounts and watchers

TODO

## TODO

- Setup local dev with telepresence
- Implement services properly, gracefully waiting for kafka, gracefully recovering from lost connections
- Add event producer pod
- Implement db storage for consumer-high|low
- Implement aggregator
- Implement simple web ui
- Update readme with k8s specific commands (logs, restarts)
- Update readme with alternative on local dev w/ file mounts and watchers
