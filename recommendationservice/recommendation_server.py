import os
import random
import time
import traceback
from concurrent import futures
# Google Cloud Debugger not supported for Python>3.8
#import googleclouddebugger
#import googlecloudprofiler
#from google.auth.exceptions import DefaultCredentialsError
import grpc
'''from opencensus.trace.exporters import print_exporter
from opencensus.trace.exporters import stackdriver_exporter
from opencensus.trace.ext.grpc import server_interceptor
from opencensus.common.transports.async_ import AsyncTransport
from opencensus.trace.samplers import always_on'''
import demo_pb2
import demo_pb2_grpc
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc
from opentelemetry import trace
from opentelemetry.exporter import jaeger
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchExportSpanProcessor
#gRPC OTEL
from opentelemetry.instrumentation.grpc import GrpcInstrumentorServer, server_interceptor
#from opentelemetry.instrumentation.grpc.grpcext import intercept_server
# create a JaegerSpanExporter
jaeger_exporter = jaeger.JaegerSpanExporter(
    service_name='recommendation-service',
    # configure agent
    agent_host_name='jaeger',
    agent_port=6831,
    # optional: configure also collector
    #collector_host_name='jaeger',
    #collector_port=14268,
    #collector_endpoint='/api/traces?format=jaeger.thrift',
    # collector_protocol='http',
    # username=xxxx, # optional
    # password=xxxx, # optional
)
# Create a BatchExportSpanProcessor and add the exporter to it
print("batch")
# Create a BatchExportSpanProcessor and add the exporter to it
span_processor = BatchExportSpanProcessor(jaeger_exporter)
print("tracer")
# add to the tracer
trace.set_tracer_provider(TracerProvider())
trace.get_tracer_provider().add_span_processor(span_processor)
grpc_server_instrumentor = GrpcInstrumentorServer()
grpc_server_instrumentor.instrument()
#from logger import getJSONLogger
#logger = getJSONLogger(‘recommendationservice-server’)


class RecommendationService(demo_pb2_grpc.RecommendationServiceServicer):
    def ListRecommendations(self, request, context):
        max_responses = 5
        # fetch list of products from product catalog stub
        cat_response = product_catalog_stub.ListProducts(demo_pb2.Empty())
        product_ids = [x.id for x in cat_response.products]
        filtered_products = list(set(product_ids)-set(request.product_ids))
        num_products = len(filtered_products)
        num_return = min(max_responses, num_products)
        # sample list of indicies to return
        indices = random.sample(range(num_products), num_return)
        # fetch product ids from indices
        prod_list = [filtered_products[i] for i in indices]
        #logger.info(“[Recv ListRecommendations] product_ids={}“.format(prod_list))
        # build and return response
        response = demo_pb2.ListRecommendationsResponse()
        response.product_ids.extend(prod_list)
        return response
    def Check(self, request, context):
        return health_pb2.HealthCheckResponse(
            status=health_pb2.HealthCheckResponse.SERVING)
    def Watch(self, request, context):
        return health_pb2.HealthCheckResponse(
            status=health_pb2.HealthCheckResponse.UNIMPLEMENTED)
if __name__ == "__main__":
    #logger.info(“initializing recommendationservice”)
    port = "8080"
    catalog_addr = "productcatlog:4000"

    #logger.info(“product catalog address: ” + catalog_addr)
    channel = grpc.insecure_channel(catalog_addr)
    product_catalog_stub = demo_pb2_grpc.ProductCatalogServiceStub(channel)
    # create gRPC server
    server = grpc.server(futures.ThreadPoolExecutor())
    # add class to gRPC server
    service = RecommendationService()
    demo_pb2_grpc.add_RecommendationServiceServicer_to_server(service, server)
    health_pb2_grpc.add_HealthServicer_to_server(service, server)
    # start server
    #logger.info(“listening on port: ” + port)
    server.add_insecure_port('[::]:'+port)
    server.start()
    # keep alive
    try:
         while True:
            time.sleep(10000)
    except KeyboardInterrupt:
            server.stop(0)