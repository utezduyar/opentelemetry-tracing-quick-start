package main

import (
	"fmt"
	"time"
)

// This is an application which happens to include a library (or a module).
// For the sake of simplicity every thing is in the main file. It is important
// to show the library part because there are certain things a library should
// or should not do according to the OpenTelemetry tracing specification.
// ---------------------------------------------------------------------------

func main() {

	InsertUser("Foo")
}

// Below is the library part.
// ---------------------------------------------------------------------------

func InsertUser(user string) error {

	time.Sleep(500 * time.Microsecond)
	fmt.Println("User has been inserted!")

	return nil
}
