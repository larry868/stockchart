// this main package contains the web assembly source code.
// It's compiled into a wasm file with "GOOS=js GOARCH=wasm go build -o ../static/main.wasm"
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gowebapi/webapi"
	"github.com/gowebapi/webapi/dom"
	"github.com/gowebapi/webapi/html"
	"github.com/gowebapi/webapi/html/htmlevent"
	"github.com/sunraylab/rgb/v2"
	"github.com/sunraylab/stockchart/stockchart"
	"github.com/sunraylab/timeline/v2"
)

// the main func is required by the GOOS=js GOARCH=wasm go builder
func main() {
	c := make(chan struct{})
	fmt.Println("Go/WASM loaded")

	// build  random dataset for the demo
	datastart := time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)
	dataset := BuildRandomDataset(300, datastart, time.Minute)
	subdataset := BuildRandomDataset(5, datastart, time.Minute*30)

	// Create a new chart
	chart, err := stockchart.NewStockChart("mychart", rgb.Gray.Lighten(0.8), *dataset, 0.1)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	chart.AddSubChart(4, &stockchart.NewDrawingCandles(subdataset).Drawing)


	// handle the button "select a candle"
	fsel := false
	btnseldata := GetButtonById("btnseldata")
	btnseldata.SetOnClick(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
		if !fsel {
			fsel = true
			chart.DoSelChangeData("", dataset.GetDataAt(datastart.Add(29*time.Minute)), false)
			btnseldata.SetInnerText("Unselect candle")
		} else {
			fsel = false
			chart.DoSelChangeData("", nil, false)
			btnseldata.SetInnerText("Select a candle")
		}
	})

	// handle the button "zoom to a period"
	fzoom := false
	btnselzoom := GetButtonById("btnselzoom")
	btnselzoom.SetOnClick(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
		if !fzoom {
			fzoom = true
			chart.DoSelChangeTimeSlice("", timeline.MakeTimeSlice(datastart.Add(2*time.Hour), 1*time.Hour), false)
			btnselzoom.SetInnerText("Unzoom")
		} else {
			fzoom = false
			chart.DoSelChangeTimeSlice("", timeline.TimeSlice{}, false)
			btnselzoom.SetInnerText("Zoom to a specific hour")
		}
	})

	fmt.Println("Go/WASM idling")
	<-c
	fmt.Println("Go/WASM exit")
}

/***************************
 * Helpers
 */

func GetElementById(elementId string) (htmlE *dom.Element) {
	doc := webapi.GetWindow().Document()
	if doc != nil {
		htmlE = doc.GetElementById(elementId)
	}
	if doc == nil || htmlE == nil {
		log.Printf("unable to find html element id=%q\n", elementId)
	}
	return htmlE
}

func GetButtonById(elementId string) (button *html.HTMLButtonElement) {
	htmlE := GetElementById(elementId)
	if htmlE != nil {
		button = html.HTMLButtonElementFromWrapper(htmlE)
		if button == nil {
			log.Printf("element id=%q is not a button\n", elementId)
		}
	}
	return button
}
