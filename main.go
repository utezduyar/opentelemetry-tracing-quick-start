package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
)

// ------------------------------------------
// BELOW IS THE APPLICATION PART
// -------------------------------------------

func main() {

	InsertUser("Foo")
}

// ------------------------------------------
// BELOW IS THE LIBRARY PART
// -------------------------------------------

const name = "name-of-my-library"

func InsertUser(user string) error {

	// Start instrumenting. It should never fail according to the spec
	// This is short for GetTracerProvider().Tracer(name, opts...)
	_, span := otel.Tracer(name).Start(context.Background(), "InsertUser")
	defer span.End()

	time.Sleep(500 * time.Microsecond)
	fmt.Println("User has been insterted")

	return nil
}
