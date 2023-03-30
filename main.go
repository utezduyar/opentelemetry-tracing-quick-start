package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type yourOwnCustomizedSampler struct{}

// ShouldSample is a method in Sampler's interface. This method is called every time, before
// creating a span with Start(). You can make as advanced analysis as you want before making a decision.
func (s yourOwnCustomizedSampler) ShouldSample(p trace.SamplingParameters) trace.SamplingResult {

	sampleDecision := trace.RecordAndSample
	if false {
		sampleDecision = trace.Drop
	}
	return trace.SamplingResult{
		Decision:   sampleDecision,
		Tracestate: oteltrace.SpanContextFromContext(p.ParentContext).TraceState(),
	}
}

// Description is the other method of the Sampler interface.
func (s yourOwnCustomizedSampler) Description() string {
	return "YourOwnCustomizedSampler (Sample which ever you want. Drop on rest)"
}

func YourOwnCustomizedSampler() trace.Sampler {
	return yourOwnCustomizedSampler{}
}

// This is an application which happens to include a library (or a module).
// For the sake of simplicity every thing is in the main file. It is important
// to show the library part because there are certain things a library should
// or should not do according to the OpenTelemetry tracing specification.
// ---------------------------------------------------------------------------

func main() {

	// Link the application with the OpenTelemetry SDK. We are telling OpenTelemetry
	// layer what to do with the collected spans. Previous no-operation calls become
	// meaningful when the application is linked with the OpenTelemetry SDK.
	// Look at ./OpenTelemetrySDK.jpeg image.
	//
	// Libraries/modules are not allowed to link with the OpenTelemetry SDK
	// according to the specifications.
	//
	// Initialize the tracer and prepare what to do during shutdown like flushing.
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize new trace provider: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("error shutting down TracerProvider: %v", err)
		}
	}()

	// Since we don't have any previous context, we take the background one
	InsertUser(context.Background(), "Foo")
	// This one will generate an error span.
	InsertUser(context.Background(), "Bar")
}

func initTracer() (*trace.TracerProvider, error) {

	// Using the resources, we are adding attributes to all our spans. These
	// attributes are going to be propagated to all the spans that are created.
	res, err := resource.New(context.Background(),
		// Telemetry SDK semantic attributes
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			// semconv package is versioned because there are breaking changes in different versions.
			// As of preparing this material, I have picked the latest one.
			semconv.ServiceNameKey.String("Workshop App"),
			semconv.ServiceVersionKey.String("v1.0.0"),
			attribute.String("foo", "bar"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource.New: %w", err)
	}

	// Create an OTLP (OpenTelemetry Protocol) exporter to send them to OpenTelemetry
	// collector that is running on your machine
	client := otlptracehttp.NewClient(otlptracehttp.WithEndpoint("localhost:4318"), otlptracehttp.WithInsecure())
	otlpexporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	// Create an stdout exporter to show the collected spans out to the stdout.
	stdoutexporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating STDOUT trace exporter: %w", err)
	}

	// Create a trace provide with given options. Look at ./ObjectDiagram.jpg
	// to see how things are tied together.
	tp := trace.NewTracerProvider(
		// There are many built in samplers but if you want more customization, you need to
		// implement the Sampler interface. That is what we do here. You can comment this line
		// out and experiment with other built in samplers shown below.
		trace.WithSampler(YourOwnCustomizedSampler()),

		// Sampling is turned on unconditionally.
		//trace.WithSampler(trace.AlwaysSample()),

		// Sampling is turned off unconditionally.
		//trace.WithSampler(trace.NeverSample()),

		// This is a percentage based sampler. For example 50% of the traces
		// will be sampled. The sample decision must be based on the trace ID
		// because otherwise there would be gaps in the trace. This sampler
		// looks at the trace ID and converts it to a data that can be used
		// to make sampling decision. Run this sampler few many times, you will
		// see that sometimes trace is sampled, sometimes not.
		//trace.WithSampler(trace.TraceIDRatioBased(0.5)),

		// Sometimes you want to make the decision by the consumer of your
		// API. This sampler looks at the parent Span to decide
		// if there should be sampler or not. The argument to this sampler
		// is used if there is no parent span before. Following translates to
		// "let the parent decide, if there is no parent, always sample". You can have
		// more fine granular parent based decision like if the parent is a remote parent etc.
		// https://opentelemetry.io/docs/reference/specification/trace/sdk/#parentbased
		//trace.WithSampler(trace.ParentBased(trace.AlwaysSample())),

		trace.WithBatcher(stdoutexporter),
		trace.WithBatcher(otlpexporter),
		trace.WithResource(res),
	)

	// Set the global trace provider as what we have created
	otel.SetTracerProvider(tp)
	return tp, nil
}

// Below is the library part.
// ---------------------------------------------------------------------------

const name = "module-or-library-name"

func InsertUser(ctx context.Context, user string) error {

	// We are instrumenting the library to create a span. Libraries do this by
	// linking to the OpenTelemetry API which is in module "go.opentelemetry.io/otel".
	// Libraries are not allowed to link with the OpenTelemetry SDK layer.
	//
	// The Start() API creates a span called InsertUser. The first parameter to it is Context.
	// This parameters defines if the created span will be child of a parent span.
	//
	// We stop this span when the method returns.

	// As long as the application is only linked with the OpenTelemetry API, the spans
	// created are almost no-operation. Look at ./OpenTelemetryAPI.jpeg image.
	//
	// Miscellaneous:
	// - otel.Tracer() is short for otel.GetTracerProvider().Tracer()
	// - Span creation should not fail and should not impact the performance according to spec.
	// - The name to the Tracer should be the name of the module or library
	// - The span name should be the most general string that identifies an class of Spans.
	//   InsertUser/foo would be too specific. There are other ways to append specific information
	//   to a span.
	//
	ctx2, span := otel.Tracer(name).Start(ctx, "InsertUser")
	defer span.End()

	// We are annotating the span. Essentially we are adding extra meta data that we can use later on.
	// Attributes, Events (a.k.a Logs) and Status are 3 different types of annotations we will add.
	// There is also Links annotation but we will look at it later.
	//
	// It is important to check for the IsRecording() API before annotating because if this span is
	// not going to be recorded, you don't want to waste any more resources annotating it.
	if span.IsRecording() {

		// You can add what ever attribute you want.
		span.SetAttributes(
			attribute.String("user.username", user),
		)

		// Events are also called span logs. It is very useful to add context information to logs
		// so that you can associate a trace down to it's logs.
		span.AddEvent("Got the mutex lock, doing work...")

		// We are going to mark this span as an error span. An error span shows up differently in UI tools.
		if user == "Bar" {
			err := fmt.Errorf("user %s is trying to hack", user)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}

	time.Sleep(500 * time.Microsecond)

	// We have gotten this context the last time we have created a span.
	// Span corrolation is carried in the context.
	LogOperation(ctx2, user)

	return nil
}

func LogOperation(ctx context.Context, user string) {
	// Just create another span.
	_, span := otel.Tracer(name).Start(ctx, "LogOperation")
	fmt.Printf("User '%s' has been added\n", user)
	defer span.End()
}
