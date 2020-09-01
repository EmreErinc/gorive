package main

import (
	"flag"
	"fmt"
	. "gorive/pkg"
	"gorive/pkg/auth"
	"gorive/pkg/http"
)

func main() {
	Banner()
	service := auth.Authorize()

	countPtr := flag.Int64("count", 20, "fetch item count")
	flag.Parse()

	fmt.Printf("...Fetching %d items...\n", *countPtr)

	http.Fetch(service, *countPtr, "")
}