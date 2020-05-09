FROM golang:1.13.0-alpine3.10 AS builder
RUN apk add git
ADD . /src/app
WORKDIR /src/app
RUN go mod download
ARG VERSION
RUN echo "Building version: ${VERSION}"
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags " -X main.Version=${VERSION}" -o rpd .

FROM alpine:edge
RUN apk add tzdata
COPY --from=builder /src/app/rpd /rpd
ADD prometheus.yml /etc/prometheus/auto/prometheus.yml
ENTRYPOINT ["./rpd"]
