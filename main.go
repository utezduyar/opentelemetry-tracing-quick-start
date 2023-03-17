package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

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
}

func initTracer() (*trace.TracerProvider, error) {
	// Create an stdout exporter to show the collected spans out to the stdout.
	stdoutexporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating STDOUT trace exporter: %w", err)
	}

	// Create a trace provide with given options. Look at ./ObjectDiagram.jpg
	// to see how things are tied together.
	tp := trace.NewTracerProvider(
		trace.WithBatcher(stdoutexporter),
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
