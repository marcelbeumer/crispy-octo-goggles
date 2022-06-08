# StreamProc

**WARNING: WORK IN PROGRESS, NOT FUNCTIONAL YET**

Basic stream processing exercise and [Kubernetes](https://kubernetes.io) local development setup.

- [Event producer](./services/event-producer) that generates events in JSON format of `{ amount: number }`
- [Event API service](./services/event-api) that takes events and writes them to [Kafka](https://kafka.apache.org)
- [Consumer service "high"](./services/consumer-high) that consumes from Kafka and writes amounts >5 to [TimescaleDB](https://www.timescale.com)
- [Consumer service "low"](./services/consumer-low) that consumes from Kafka and writes amounts <=5 to [InfluxDB](https://www.influxdata.com)
- [Aggregator service](./services/aggregator) providing both high and low time series data
- [Web UI](./services/web-ui) that renders line chart for high and low data

## System requirements

- Install [node](https://nodejs.org/en/) >=16
- Install [golang](https://go.dev) >=1.18
- Install [watchexec](https://github.com/watchexec/watchexec)
- Install [docker](https://www.docker.com)
- Install [k3d](https://k3d.io) or [kind](https://kind.sigs.k8s.io)
- Install [helm](https://helm.sh)
- Install [helmdiff](https://github.com/databus23/helm-diff)
- Install [helmfile](https://github.com/roboll/helmfile)
