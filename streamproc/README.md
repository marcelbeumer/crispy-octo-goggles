# StreamProc

Basic stream processing exercise and [Kubernetes](https://kubernetes.io) local development setup.

- [Event producer](./services/event-producer) that generates events in JSON format of `{ amount: number }`.
- [Event API service](./services/event-api) that takes events and writes them to [Kafka](https://kafka.apache.org).
- [Consumer service "high"](./services/consumer-high) that consumes from Kafka and writes amounts >5 to [TimescaleDB](https://www.timescale.com).
- [Consumer service "low"](./services/consumer-low) that consumes from Kafka and writes amounts <=5 to [InfluxDB](https://www.influxdata.com).
- [Aggregator service](./services/aggregator) providing both high and low time series data.
- [Web UI](./services/web-ui) that renders line chart for high and low data.

## System requirements

- [go](https://go.dev) >=1.18
- [k3d](https://k3d.io) >=5.4.3 or [kind](https://kind.sigs.k8s.io) >=0.14
- [helm](https://helm.sh) >=3.9.0

## Installation

- Build local docker images using `./scripts/build_images.sh`.
- Create a cluster with `./scripts/create_k3d_cluster.sh` or `./scripts/create_kind_cluster.sh`.
- Push local images to the cluster registry with `./scripts/push_images.sh`.
- Install helm chart with `helm install streamproc ./helm_chart`.
- Check when pods are ready with `kubectl get po`
- Add `127.0.0.1 streamproc.local` to your `/etc/hosts` file
- Open [http://streamproc.local](http://streamproc.local). The graph updates every x-seconds.

## Local development

Some strategies for local development with a (local) k8s cluster:

- Use `helm upgrade` after pushing new docker images on code change. Slow and tedious, but real cluster deployment.
- Use `kubectl port-forward` to expose services from the cluster to the local machine, using those in local processes / services.
- Use [telepresence](https://www.telepresence.io) to intercept traffic from/to a service in the cluster to the local machine.
- Use combination of file mounts and file watchers to initiate rebuilds on code change and automatically restarting services.

Each service implements a `DISABLE` env var (`-x` CLI opt) that disables IO/processing in the cluster so you can do
that in your local service instead. That's not always neccesary, but can be helpful to prevent the cluster processes to
impact downstream services while you are making local implementation changes.

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

This is not demonstrated in this repository because it requires a bunch of configuration and scripting, worth a separate demo project/repo.

First, you need to decide if you want to rebuild in the docker container on the cluster or on your local machine.

When you want to rebuild in the docker container you will to mount the source tree from your local machine to the cluster and containers. You then create a separate "dev target" in each Dockerfile that watches the source tree (using [watchexec](https://watchexec.github.io) is recommended), rebuilds from source and restarts the service, or just `go run` from source and restart on change. The advantage is that you don't need to do cross compilation _for_ the container, but the downside is that you need to mount the source tree, which can become a performance issue.

When you want to rebuild on your local machine (_for_ the docker container) you will need to mount the local build artifacts to your cluster and containers. Also for this solution your create a separate "dev target" in each Dockerfile that watches the binary (build artifact(s)) and restarts the service on change (also here running [watchexec](https://watchexec.github.io) in the container is recommended). The local machine then needs to correctly cross compile for the container (`GOOS=<os> GOARCH=<arch> go build`). The advantage is that file watching will not become a performance issue, but the downside is that you need to manage your cross compilation (and OS/arch detection) in scripts.

## Improvements

- Write README's for each service.
- Time buckets used by the aggregator should be fixed in time and not change on every call/update.
- Currently the web-ui ingress is not rewriting path prefixes, which would be nice. Traefik (k3d) and NGINX ingress (kind) work differently with rewriting, was avoiding the k8s config complexity so far. See [traefik doc here](https://doc.traefik.io/traefik/migration/v1-to-v2/#strip-and-rewrite-path-prefixes).
