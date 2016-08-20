package main

import "fmt"

const version = "0.4.0"

func cmdVersion() {
	fmt.Fprintf(stdout(), "%s\n", version)
}
