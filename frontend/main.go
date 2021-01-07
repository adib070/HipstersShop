// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	exporttrace "go.opentelemetry.io/otel/sdk/export/trace"
	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	port            = "8081"
	defaultCurrency = "USD"
	cookieMaxAge    = 60 * 60 * 48

	cookiePrefix    = "shop_"
	cookieSessionID = cookiePrefix + "session-id"
	cookieCurrency  = cookiePrefix + "currency"

	serviceName = "Frontend"
)

var (
	whitelistedCurrencies = map[string]bool{
		"USD": true,
		"EUR": true,
		"CAD": true,
		"JPY": true,
		"GBP": true,
		"TRY": true}
)

type ctxKeySessionID struct{}

type frontendServer struct {
	productCatalogSvcAddr string
	productCatalogSvcConn *grpc.ClientConn

	currencySvcAddr string
	currencySvcConn *grpc.ClientConn

	cartSvcAddr string
	cartSvcConn *grpc.ClientConn

	recommendationSvcAddr string
	recommendationSvcConn *grpc.ClientConn

	checkoutSvcAddr string
	checkoutSvcConn *grpc.ClientConn

	shippingSvcAddr string
	shippingSvcConn *grpc.ClientConn

	adSvcAddr string
	adSvcConn *grpc.ClientConn
}

func frontendserverConstructor(productCatalogSvcAddr string, currencySvcAddr string, cartSvcAddr string, recommendationSvcAddr string, checkoutSvcAddr string, shippingSvcAddr string, adSvcAddr string) *frontendServer {
	obj := new(frontendServer)

	obj.productCatalogSvcAddr = productCatalogSvcAddr
	obj.currencySvcAddr = currencySvcAddr
	obj.cartSvcAddr = cartSvcAddr
	obj.recommendationSvcAddr = recommendationSvcAddr
	obj.checkoutSvcAddr = checkoutSvcAddr
	obj.shippingSvcAddr = shippingSvcAddr
	obj.adSvcAddr = adSvcAddr

	return obj
}

func main() {
	ctx := context.Background()
	log := logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout

	var traceExporter exporttrace.SpanExporter
	var addr1 = "http://localhost:14268/api/traces"
	log.Infof("exporting to Jaeger endpoint at %s", addr1)
	traceExporter, _ = jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint(addr1),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: serviceName,
		}),
	)

	if os.Getenv("DISABLE_TRACING") == "" {
		log.Info("Tracing enabled.")
		initTracing(log, traceExporter)
	} else {
		log.Info("Tracing disabled.")
	}

	srvPort := port

	addr := os.Getenv("LISTEN_ADDR")

	var PRODUCT_CATALOG_SERVICE_ADDR = "localhost:4000"
	var CURRENCY_SERVICE_ADDR = "localhost:9000"
	var CART_SERVICE_ADDR = "localhost:7528"
	var RECOMMENDATION_SERVICE_ADDR = "localhost:8280"
	var CHECKOUT_SERVICE_ADDR = "localhost:5050"
	var SHIPPING_SERVICE_ADDR = "localhost:50051"
	var AD_SERVICE_ADDR = "localhost:9555"

	svc := frontendserverConstructor(PRODUCT_CATALOG_SERVICE_ADDR, CURRENCY_SERVICE_ADDR, CART_SERVICE_ADDR, RECOMMENDATION_SERVICE_ADDR, CHECKOUT_SERVICE_ADDR, SHIPPING_SERVICE_ADDR, AD_SERVICE_ADDR)

	mustConnGRPC(ctx, &svc.currencySvcConn, svc.currencySvcAddr)
	mustConnGRPC(ctx, &svc.productCatalogSvcConn, svc.productCatalogSvcAddr)
	mustConnGRPC(ctx, &svc.cartSvcConn, svc.cartSvcAddr)
	mustConnGRPC(ctx, &svc.recommendationSvcConn, svc.recommendationSvcAddr)
	mustConnGRPC(ctx, &svc.shippingSvcConn, svc.shippingSvcAddr)
	mustConnGRPC(ctx, &svc.checkoutSvcConn, svc.checkoutSvcAddr)
	mustConnGRPC(ctx, &svc.adSvcConn, svc.adSvcAddr)

	r := mux.NewRouter()
	r.Use(MuxMiddleware(), otelmux.Middleware(serviceName))
	r.HandleFunc("/", svc.homeHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/product/{id}", svc.productHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/cart", svc.viewCartHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/cart", svc.addToCartHandler).Methods(http.MethodPost)
	r.HandleFunc("/cart/empty", svc.emptyCartHandler).Methods(http.MethodPost)
	r.HandleFunc("/setCurrency", svc.setCurrencyHandler).Methods(http.MethodPost)
	r.HandleFunc("/logout", svc.logoutHandler).Methods(http.MethodGet)
	r.HandleFunc("/cart/checkout", svc.placeOrderHandler).Methods(http.MethodPost)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "User-agent: *\nDisallow: /") })
	r.HandleFunc("/_healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })

	var handler http.Handler = r
	handler = &logHandler{log: log, next: handler} // add logging
	handler = ensureSessionID(handler)             // add session ID
	log.Infof("starting server on " + addr + ":" + srvPort)
	log.Fatal(http.ListenAndServe(addr+":"+srvPort, handler))
}

func initTracing(log logrus.FieldLogger, syncer exporttrace.SpanExporter) {
	tp := trace.NewTracerProvider(
		trace.WithConfig(
			trace.Config{
				DefaultSampler: trace.AlwaysSample(),
				Resource:       res,
			},
		),
		trace.WithSyncer(syncer),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func mustConnGRPC(ctx context.Context, conn **grpc.ClientConn, addr string) {
	var err error

	*conn, err = grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		panic(errors.Wrapf(err, "grpc: failed to connect %s", addr))
	}
}
