FROM golang:1.20-alpine AS builder
ARG INSTANCE_TZ=Africa/Kampala
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
    ENV TZ=${INSTANCE_TZ}
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o consumer
WORKDIR /dist
RUN cp /build/consumer .

FROM scratch
COPY --from=builder /dist/consumer /go/bin/consumer
COPY --from=builder  /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /
ARG INSTANCE_TZ=Africa/Kampala
ENV TZ=${INSTANCE_TZ}
ENV ZONEINFO=/zoneinfo.zip
ENTRYPOINT ["/go/bin/consumer"]