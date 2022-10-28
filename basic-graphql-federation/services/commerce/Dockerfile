FROM golang:1.19 as build
WORKDIR /build
COPY . .
RUN GOOS=linux make

FROM alpine:3.15
COPY --from=build /build/build/server /

ENTRYPOINT [ "/server" ]
