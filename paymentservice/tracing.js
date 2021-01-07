//tracing.js
'use strict';
const { NodeTracerProvider } = require('@opentelemetry/node');
const { SimpleSpanProcessor } = require("@opentelemetry/tracing");
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const { CollectorTraceExporter } = require('@opentelemetry/exporter-collector-grpc')

const exportType = process.env.EXPORT_TYPE || "jaeger"
console.log("Exporter type Set To :" + exportType)

const OTLP_HEADERS = {
    'Api-Key': process.env.OTLOP_API_KEY,
}

function getExporter(exporterType) {
    switch (exporterType) {
        case 'otlp':
            return new CollectorTraceExporter({
                headers: new OTLP_HEADERS,
                url: process.env.OTEL_EXPORTER_OTLP_SPAN_ENDPOINT
            })
        case 'jaeger':
        default:
            console.log("Jaeger Set  ")
            return new JaegerExporter({
                serviceName: process.env.SERVICE_NAME || "Payment",
                endpoint: process.env.ENDPOINT || "jaeger:14250",
                username: process.env.USER_NAME,
                password: process.env.PASSWORD

            })
    }
}

const exporter = getExporter(exportType)
const traceProvider = new NodeTracerProvider()
traceProvider.addSpanProcessor(
    new SimpleSpanProcessor(exporter)
)

traceProvider.register()

console.log("tracing initialized");