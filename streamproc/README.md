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

Some strategies for local development with a (local) k8s cluster:

- Use `helm upgrade` after pushing new docker images on code change. Slow and tedious, but real cluster deployment.
- Use `kubectl port-forward` to expose services from the cluster to the local machine, using those in local processes / services.
- Use [telepresence](https://www.telepresence.io) to intercept traffic from/to a service in the cluster to the local machine.
- Use combination of file mounts and file watchers to initiate rebuilds on code change and automatically restarting services.

### Helm upgrade after pushing new docker images

Helm chart contains deployment metadata annotations (`{{ now | quote }}` on `spec.template.metadata.annotations.timestamp`).
In combination with `pullPolicy: Always`, this will cause k8s to redeploy when we upgrade the helm chart to a new revision.

- Run `./scripts/build_images.sh && ./scripts/push_images.sh`.
- Run `helm upgrade streamproc ./helm_chart`.

Use `kubectl get pod -w` to see the redeploy happening.

### Kubectl port-forward

Example running "event-producer" locally, which requires "event-api" from the cluster.

- Run `kubectl port-forward svc/streamproc-event-api 9998:9998` in a seperate shell.
- Run `cd services/event-producer && go run .` in another shell.

### Telepresence

Example running "event-api" locally.

- Install [telepresence](https://www.telepresence.io).
- Run `telepresence connect` to connect to cluster in context.
- Run `telepresence list` to see list of services that can be intercepted.
- Run `telepresence intercept streamproc-event-api --env-file ~/tmp.env` to start intercepting.
- Run `export $(cat ~/tmp.env | xargs)` to export all env vars.
- Run `cd services/event-api && go run .` to start the service locally.

### File mounts and watchers

TODO

## TODO

- Add event producer pod
- Implement db storage for consumer-high|low
- Implement aggregator
- Implement simple web ui
- Update readme with k8s specific commands (logs, restarts)
- Update readme with alternative on local dev w/ file mounts and watchers
