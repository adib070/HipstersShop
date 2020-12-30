# HipstersShop



## Development Guide 

This doc explains how to build and run the Hipstershop source code locally.  

## Prerequisites 
- Jaeger
- [Docker for Desktop](https://www.docker.com/products/docker-desktop).
- JDK 11
- Installation of Go
- Installation of Python
- Visual Studio

## Steps

1. Currency Service (Node.js)

    ```sh
    cd currencyservice
    npm i 
    node -r ./tracing server.js tracing initialized
    
    ```
2. Cart Service (C#)
      
    ```sh
    Opent and Run the cart service in Visual Studio.
   
    ```
  
3. Payment Service(Node.js)
  
    ```sh
    cd paymentservice
    node -r ./tracing index.js  
    
    ```
    
4. Recommendation Service (Python)
  
    ```sh
    cd recommendationservice
    pip install -r requirements.txt
    opentelemetry-instrument -e none python3 recommendation_server.py
    
    ```
    
5. Shipping Service(Go)
  
    ```sh
    cd shippingservice
    go run . .
    
    ```
    
6. ProductCatlog Service (Go)
  
    ```sh
    cd productcatalogservice
    go run . .
    
    ```
    
7. Checkout Service (Go)
  
    ```sh
    cd checkoutservice
    go run . .
    
    ```
8. Email Service (Python)

    ```sh
    cd emailservice
    export OTEL_PYTHON_TRACER_PROVIDER=sdk_tracer_provider
    opentelemetry-instrument -e none python3 email_server.py
    
    ```
    
9. Ad Service (Java)
 > Note : Download opentelemetry-javaagent-all.jar : https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v0.8.0/opentelemetry-javaagent-all.jar and copy the jar file in folder adservice/tracinglib

    ```sh
    cd adservice
    gradle build
    java -javaagent:tracinglib/opentelemetry-javaagent-all.jar\
    -Dotel.exporter=jaeger \
    -Dotel.exporter.jaeger.service.name=adService \
    -Dotel.exporter.jaeger.endpoint=localhost:14250 \
    -jar build/libs/hipstershop-0.1.0-SNAPSHOT-fat.jar
     
    ```

 10.  Access the web frontend through your browser 
  
  - Once run above all steps you can access frontend service at  http://localhost:8081
  - start the jaeger either using binary file or using docker desktop http://localhost:16686/ , You will see traces in jaeger
    
