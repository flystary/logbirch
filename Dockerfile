FROM docker.io/library/golang:1.20 as build

WORKDIR /app
COPY . /app

ENV CGO_ENABLED=0

RUN go install mvdan.cc/garble@latest
RUN garble -tiny -literals -seed=Betaidc6666 build -o logbrich ./cmd/logbrich

FROM nixery.dev/shell/upx as upx

COPY --from=build /app/logbrich /
RUN upx -9 /logbrich

FROM alpine:edge

COPY --from=upx /logbrich /

ENTRYPOINT ["/logbrich"]
