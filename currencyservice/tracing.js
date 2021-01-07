'use strict';
const { logger } = require('@opencensus/core');
const { LogLevel } = require('@opentelemetry/core');
const { NodeTracerProvider } = require('@opentelemetry/node');
const { SimpleSpanProcessor } = require('@opentelemetry/tracing');
const { CollectorTraceExporter } = require('@opentelemetry/exporter-collector-grpc')
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const exportType = process.env.EXPORT_TYPE
const provider = new NodeTracerProvider({
    logLevel: LogLevel.DEBUG,
});

provider.register();

function getExporter(exporterType) {
    switch (exporterType) {
        case 'otlp':
            return new CollectorTraceExporter({
                url: process.env.OTEL_EXPORTER_OTLP_SPAN_ENDPOINT
            })
        case 'jaeger':
        default:
            console.log('Jaeger tracing Initalised');
            return new JaegerExporter({
                serviceName: 'currency',
                JAEGER_AGENT_PORT: 6832,
                JAEGER_AGENT_HOST: "jaeger",
                //endpoint: "http://localdev.logicmonitor.com:8080/rest/traces/ingest",
                //username: "LMSTv2 eI5hW3rVbHCi4b44qkWa",
                //password: "abcd",

            })

    }
}

const exporter = getExporter(exportType)

if (exporter != null) {
    const traceProvider = new NodeTracerProvider()

    traceProvider.addSpanProcessor(
        new SimpleSpanProcessor(exporter)
    )

    traceProvider.register()

}
console.log('tracing initialized');