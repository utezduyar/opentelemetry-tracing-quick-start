package main

import (
	"fmt"
	"time"
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

	time.Sleep(500 * time.Microsecond)
	fmt.Println("User has been insterted")

	return nil
}
