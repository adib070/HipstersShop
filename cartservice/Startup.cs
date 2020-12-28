using System;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using OpenTelemetry;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;
using cartservice.cartstore;
using cartservice.services;

namespace cartservice
{
    public class Startup
    {
        public Startup(IConfiguration configuration)
        {
            Configuration = configuration;
        }

        public IConfiguration Configuration { get; }

        // This method gets called by the runtime. Use this method to add services to the container.
        // For more information on how to configure your application, visit https://go.microsoft.com/fwlink/?LinkID=398940
        public void ConfigureServices(IServiceCollection services)
        {
            string redisAddress = Configuration["REDIS_ADDR"];
            ICartStore cartStore = null;
            if (!string.IsNullOrEmpty(redisAddress))
            {
                cartStore = new RedisCartStore(redisAddress);
            }
            else
            {
                Console.WriteLine("Redis cache host(hostname+port) was not specified. Starting a cart service using local store");
                Console.WriteLine("If you wanted to use Redis Cache as a backup store, you should provide its address via command line or REDIS_ADDRESS environment variable.");
                cartStore = new LocalCartStore();
            }

            // Initialize the redis store
            cartStore.InitializeAsync().GetAwaiter().GetResult();
            Console.WriteLine("Initialization completed");

            services.AddControllers();
            services.AddGrpc();
            services.AddSingleton<ICartStore>(cartStore);
            services.AddOpenTelemetryTracing(builder => ConfigureOpenTelemetry(builder, cartStore));
       

        }
        private static void ConfigureOpenTelemetry(TracerProviderBuilder builder, ICartStore cartStore)
        {
            builder.AddAspNetCoreInstrumentation();

            if (cartStore is RedisCartStore redisCartStore)
            {
                builder.AddRedisInstrumentation(redisCartStore.ConnectionMultiplexer);
            }

            var exportType = Environment.GetEnvironmentVariable("EXPORT_TYPE") ?? "jaeger";
            var serviceName = "CartService" + (exportType == "newrelic" ? string.Empty : $"-{exportType}");

            builder.SetResourceBuilder(ResourceBuilder.CreateDefault().AddService(serviceName, null, null, false, $"{exportType}-{Guid.NewGuid().ToString()}"));

            switch (exportType)
            {
                case "otlp":
                    var otlpEndpoint = Environment.GetEnvironmentVariable("OTEL_EXPORTER_OTLP_SPAN_ENDPOINT")
                        ?? Environment.GetEnvironmentVariable("OTEL_EXPORTER_OTLP_ENDPOINT");
                    builder
                        .AddOtlpExporter(options => options.Endpoint = otlpEndpoint);
                    break;
                case "jaeger":
                default:
             
                    builder.AddAspNetCoreInstrumentation().AddJaegerExporter();
                    Console.WriteLine("Jaeger Tracing completed");
                    break;
            }
        }
        // This method gets called by the runtime. Use this method to configure the HTTP request pipeline.
        public void Configure(IApplicationBuilder app, IWebHostEnvironment env)
        {
            if (env.IsDevelopment())
            {
                app.UseDeveloperExceptionPage();
            }

            app.UseRouting();

            app.UseEndpoints(endpoints =>
            {
                endpoints.MapGrpcService<CartService>();
                endpoints.MapGrpcService<cartservice.services.HealthCheckService>();

                endpoints.MapGet("/", async context =>
                {
                    await context.Response.WriteAsync("Communication with gRPC endpoints must be made through a gRPC client. To learn how to create a client, visit: https://go.microsoft.com/fwlink/?linkid=2086909");
                });
            });
        }
    }
    }
    