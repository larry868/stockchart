// this main package contains the web assembly source code.
// It's compiled into a wasm file with "GOOS=js GOARCH=wasm go build -o ../static/main.wasm"
package main

import (
	"fmt"
	"os"

	"github.com/sunraylab/gowebstockchart/stockchart"
	"github.com/sunraylab/rgb/v2"
)

// the main func is required by the GOOS=js GOARCH=wasm go builder
func main() {
	c := make(chan struct{})
	fmt.Println("Go/WASM loaded")

	// build  random dataset for the demo
	dataset := BuildRandomDataset()

	// build a new chart
	_, err := stockchart.NewStockChart("mychart", rgb.Gray.Lighten(0.8), dataset)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Go/WASM idling")
	<-c
	fmt.Println("Go/WASM exit")
}
