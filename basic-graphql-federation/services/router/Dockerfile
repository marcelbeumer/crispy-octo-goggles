# Based on https://github.com/apollographql/router/issues/1192
# TODO: use leaner linux images

# -----
# Build
# -----
FROM --platform=linux/amd64 ubuntu:latest as build 
WORKDIR /build

RUN apt update && apt -y install curl
COPY . .
RUN ./scripts/install.sh

# ----
# Prod
# ----
FROM --platform=linux/amd64 ubuntu:latest as prod 

WORKDIR /app
RUN apt-get update && apt-get -y install openssl ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build /root/.rover /root/.rover
COPY --from=build /build .

CMD ["./scripts/run-prod.sh"]
