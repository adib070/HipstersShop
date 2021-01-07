FROM mcr.microsoft.com/dotnet/sdk:5.0.100 as builder
WORKDIR /app
COPY . .
RUN dotnet publish -r linux-musl-x64 --self-contained true -p:PublishTrimmed=True -p:TrimMode=Link -c release -o /cartservice
# PublishSingleFile=true causes this issue when using the OTLP exporter https://github.com/grpc/grpc/issues/24266
# RUN dotnet publish -p:PublishSingleFile=true -r linux-musl-x64 --self-contained true -p:PublishTrimmed=True -p:TrimMode=Link -c release -o /cartservice

FROM mcr.microsoft.com/dotnet/runtime-deps:5.0.0-alpine3.12-amd64
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.5 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe
WORKDIR /app
COPY --from=builder /cartservice .

# Work-around to make gRPC native library work on alpine containers
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.8/main' >> /etc/apk/repositories && \
    apk update --no-cache && \
    apk add --no-cache bash libc6-compat=1.1.19-r11

ARG NEW_RELIC_API_KEY
ENV NEW_RELIC_API_KEY=${NEW_RELIC_API_KEY}

ARG NEW_RELIC_TRACE_URL
ENV NEW_RELIC_TRACE_URL=${NEW_RELIC_TRACE_URL}

ARG NEW_RELIC_METRIC_URL
ENV NEW_RELIC_METRIC_URL=${NEW_RELIC_METRIC_URL}

ARG NEW_RELIC_DEMO_EXPORT_TYPE
ENV NEW_RELIC_DEMO_EXPORT_TYPE=${NEW_RELIC_DEMO_EXPORT_TYPE}

ARG OTEL_EXPORTER_OTLP_ENDPOINT
ENV OTEL_EXPORTER_OTLP_ENDPOINT=${OTEL_EXPORTER_OTLP_ENDPOINT}

ARG OTEL_EXPORTER_OTLP_SPAN_ENDPOINT
ENV OTEL_EXPORTER_OTLP_SPAN_ENDPOINT=${OTEL_EXPORTER_OTLP_SPAN_ENDPOINT}

ENTRYPOINT ["/app/cartservice"]