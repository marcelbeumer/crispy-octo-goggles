FROM golang:1.18 as build
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine:3.15
COPY --from=build /build/event-producer /

ENTRYPOINT [ "/event-producer" ]
