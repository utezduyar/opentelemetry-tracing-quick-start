package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func WeekdayDetector() resource.Detector {
	return &detector{}
}

type detector struct {
}

func (d *detector) Detect(ctx context.Context) (*resource.Resource, error) {
	return resource.NewSchemaless(
		attribute.Key("Weekday").String(time.Now().Weekday().String())), nil
}

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

// textMapCarrier Implements propagation.TextMapCarrier interface which is part
// of the methods of TextMapPropagator interface.
type textMapCarrier struct {
	m map[string]string
}

func newTextCarrier() *textMapCarrier {
	return &textMapCarrier{m: map[string]string{}}
}

func (t *textMapCarrier) Get(key string) string {
	return t.m[key]
}

func (t *textMapCarrier) Set(key string, value string) {
	t.m[key] = value
}

func (t *textMapCarrier) Keys() []string {
	str := make([]string, 0, len(t.m))

	for key := range t.m {
		str = append(str, key)
	}

	return str
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

	// Register TextMapPropagator interface implementers (trace context and baggage) as to
	// be called when we do a context propagation. The sender part of the context propagation calls
	// the Inject method and the receiver side calls Extract.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	// Since we don't have any previous context, we take the background one
	InsertUser(context.Background(), "Foo")
	// This one will generate an error span.
	//InsertUser(context.Background(), "Bar")

	//ExampleContextPropagation()
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
		// Pull attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables
		resource.WithFromEnv(),
		// Discovers process information
		resource.WithProcess(),
		// Discovers OS information
		resource.WithOS(),
		// Discovers container information
		resource.WithContainer(),
		// Discovers host information
		resource.WithHost(),
		// Bring your own external Detector implementation
		resource.WithDetectors(WeekdayDetector()),
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

func ExampleContextPropagation() {
	// As an application developer, you rarely need to interact with the context propagation APIs directly.
	// Very likely there is an API or a library that handles the context propagation. For example
	// go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp helps with context propagation
	// over HTTP with W3C Header traceparent=.

	// This example is not realistic because we could have done a context propagation over the golang
	// context. It tries to show how context is serialized and deserialized with OpenTelemetry APIs.
	// TraceContext and Baggage will be propagated through contexts below.

	// Create a baggage.
	bag, _ := baggage.Parse("foo=bar")
	// Add Baggage to the context.
	ctxSender := baggage.ContextWithBaggage(context.Background(), bag)

	// Create a span for the sending part.
	ctxSender, spanSender := otel.Tracer(name).Start(ctxSender, "Sender")
	defer spanSender.End()

	tmc := newTextCarrier()
	// Serialize all registered TextMapPropagators to the tmc data structure.
	otel.GetTextMapPropagator().Inject(ctxSender, tmc)
	fmt.Printf("Carrier dump [%+v]\n", tmc)

	// Propagation boundary ---------------------------------------------

	// Deserialize all registered TextMapPropagators
	ctxReceiver := otel.GetTextMapPropagator().Extract(context.TODO(), tmc)
	fmt.Printf("Properties of the baggage:\n")
	for _, m := range baggage.FromContext(ctxReceiver).Members() {
		fmt.Printf("%s:%s\n", m.Key(), m.Value())
	}
	// Create a span for the receiving part.
	_, spanReceiver := otel.Tracer(name).Start(ctxReceiver, "Receiver")
	defer spanReceiver.End()
}

func InsertUser(ctx context.Context, user string) error {

	// Let's prepare the structure that will be linked against the span below. In this example
	// we are making up TraceID and SpanID. Multiple Link structure can be linked to the same span
	a := []attribute.KeyValue{attribute.Bool("one", true), attribute.Bool("two", true)}
	l := oteltrace.Link{
		SpanContext: oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
			TraceID:    [16]byte{0x01},
			SpanID:     [8]byte{0x01},
			TraceFlags: 0x1,
		}),
		Attributes: a,
	}

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
	ctx2, span := otel.Tracer(name).Start(ctx, "InsertUser", oteltrace.WithLinks(l))
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
