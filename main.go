package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
)

// This is an application which happens to include a library (or a module).
// For the sake of simplicity every thing is in the main file. It is important
// to show the library part because there are certain things a library should
// or should not do according to the OpenTelemetry tracing specification.
// ---------------------------------------------------------------------------

func main() {

	// Since we don't have any previous context, we take the background one
	InsertUser(context.Background(), "Foo")
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
	_, span := otel.Tracer(name).Start(ctx, "InsertUser")
	defer span.End()

	time.Sleep(500 * time.Microsecond)
	fmt.Println("User has been inserted!")

	return nil
}
