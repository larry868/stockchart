/*
stockchart generate interactive, responsive, and performant chart with HTML5 canvas and web assembly.
*/
package stockchart

import (
	"fmt"
	"log"
	"strings"

	"github.com/sunraylab/rgb/v2"
	"github.com/sunraylab/timeline/v2"

	"github.com/gowebapi/webapi"
	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/css/typedom"
	"github.com/gowebapi/webapi/dom"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/gowebapi/webapi/html/htmlevent"
)

// StockChart is a 2D chart draw embedded into an HTML5 element
type StockChart struct {
	ID string // the identifier of this chart, the canvas id

	masterE   *dom.Element // the master element containing the chart
	layers    [6]*Layer    // the 6 drawing layers composing a stockchart
	isDrawing bool         // flag signaling a drawing in progress

	timeRange     timeline.TimeSlice // the overall time range to display
	timeSelection timeline.TimeSlice // the current time selection
}

// String interface for StockChart, mainly for debugging purpose
func (chart StockChart) String() string {
	str := fmt.Sprintf("%s\n", chart.ID)
	for _, player := range chart.layers {
		if player != nil {
			str += fmt.Sprintf("  layer %22s: %p\n", player.id, player)
			for _, pdrawing := range player.drawings {
				if pdrawing != nil {
					str += fmt.Sprintf("    drawing %18s: %p series:%p %v\n", pdrawing.Name, pdrawing, pdrawing.series, pdrawing.series)
				}
			}
		}
	}
	return str
}

// SetTimeRange defines the overall time range to display.
// Update timeselection if required
func (pchart *StockChart) SetTimeRange(timerange timeline.TimeSlice) {

	fstuck := pchart.timeRange.To.Equal(pchart.timeSelection.To)

	pchart.timeRange.From = timerange.From
	pchart.timeRange.To = timerange.To

	if pchart.timeSelection.From.IsZero() || pchart.timeSelection.From.Before(pchart.timeRange.From) {
		pchart.timeSelection.From = pchart.timeRange.From
	}
	if fstuck || pchart.timeSelection.To.IsZero() || pchart.timeSelection.To.After(pchart.timeRange.To) {
		pchart.timeSelection.To = pchart.timeRange.To
	}
}

// NewStockChart initialize a stockchart within the <stockchart> HTML element idenfied by chartid.
// An HTML page can have multiple <stockchart> but with different chartid. The layout of the chart is composed of multiples layers which are stacked canvas.
//
// Returns the stockchart created or an error if canvasid is not found.
func NewStockChart(chartid string, bgcolor rgb.Color, series *DataList) (*StockChart, error) {
	// some cleaning
	chartid = strings.ToLower(strings.Trim(chartid, " "))

	// Get the <stockchart> element who will embedd the chart
	stockchartE, err := getChartElement(chartid)
	if err != nil {
		return nil, err
	}

	// build the chart object
	chart := &StockChart{ID: chartid, masterE: stockchartE}

	// by default the timeMaxRange is the full timeslice + 10% to represents the future.
	ts := series.TimeSlice()
	if ts.Duration() != nil {
		ts.ToExtend(timeline.Duration(float64(*ts.Duration()) * 0.1))
	}
	chart.SetTimeRange(ts)

	// create master layer layout, white bg and without drawings
	// the master layer covers all the <stockchart> element size, and is build first at the background
	if layer := chart.addNewLayer("0-bg", lAREA_FULL, bgcolor, nil); layer != nil {
		chart.layers[0] = layer
	}

	// the navbar layer, updated only when navXAxisRange change
	if layer := chart.addNewLayer("1-navbar", lAREA_NAVBAR, rgb.White, &chart.timeRange); layer != nil {
		layer.AddDrawing(&NewDrawingSeries(series, true).Drawing, rgb.White)
		layer.AddDrawing(&NewDrawingXGrid(series, true, false).Drawing, rgb.None)
		chart.layers[1] = layer
	}

	// the transparent time selector layer
	if layer := chart.addNewLayer("2-timeselector", lAREA_NAVBAR, rgb.None, &chart.timeRange); layer != nil {
		layer.AddDrawing(&NewDrawingTimeSelector(series).Drawing, rgb.None)
		layer.SetEventDispatcher()
		chart.layers[2] = layer
	}

	// the yscale layer
	if layer := chart.addNewLayer("3-yscale", lAREA_YSCALE, rgb.White, &chart.timeSelection); layer != nil {
		layer.AddDrawing(&NewDrawingYGrid(series, true).Drawing, rgb.White)
		chart.layers[3] = layer
	}

	// the chart layer
	if layer := chart.addNewLayer("4-chart", lAREA_GRAPH, rgb.White, &chart.timeSelection); layer != nil {
		layer.AddDrawing(&NewDrawingBackground(series).Drawing, rgb.None)
		layer.AddDrawing(&NewDrawingYGrid(series, false).Drawing, rgb.None)
		layer.AddDrawing(&NewDrawingXGrid(series, false, true).Drawing, rgb.None)
		layer.AddDrawing(&NewDrawingCandles(series).Drawing, rgb.White)
		chart.layers[4] = layer
	}

	// the hover transprent layer
	if layer := chart.addNewLayer("5-hover", lAREA_GRAPH, rgb.None, &chart.timeSelection); layer != nil {
		layer.AddDrawing(&NewDrawingHoverCandles(series).Drawing, rgb.White)
		layer.SetEventDispatcher()
		chart.layers[5] = layer
	}

	// Add event listener on resize event
	webapi.GetWindow().AddEventResize(func(event *htmlevent.UIEvent, win *webapi.Window) {
		// resizing the chart will resize and redraw every layers
		chart.resize()
	})

	// size it the first time to force a full redraw
	chart.resize()
	return chart, nil
}

// addNewLayer creates a new canvas, inside the masterE div and add it to the stack of layers within the pchart.
// This new layer is moved and sized according to layoutArea parameter.
// It's background color is setup if any.
//
// The created layer embed the WEBGL 2D drawing context,
//
// Return the created layer, or nil if error.
func (pchart *StockChart) addNewLayer(layerid string, layout layoutT, bgcolor rgb.Color, xrange *timeline.TimeSlice) *Layer {
	// create a canvas
	domE := webapi.GetWindow().Document().CreateElement("canvas", &webapi.Union{Value: js.ValueOf("dom.Node")})
	domE.SetId("canvas" + pchart.ID + strings.ToLower(strings.Trim(layerid, " ")))
	newE := pchart.masterE.AppendChild(&domE.Node)
	canvasE := canvas.HTMLCanvasElementFromWrapper(newE)
	canvasE.AttributeStyleMap().Set("position", &typedom.Union{Value: js.ValueOf(`absolute`)})
	canvasE.AttributeStyleMap().Set("border", &typedom.Union{Value: js.ValueOf(`none`)})
	canvasE.AttributeStyleMap().Set("padding", &typedom.Union{Value: js.ValueOf(`0`)})
	canvasE.AttributeStyleMap().Set("margin", &typedom.Union{Value: js.ValueOf(`0`)})

	// create the layer
	layer := &Layer{
		id:         strings.ToLower(strings.Trim(layerid, " ")),
		layout:     layout,
		chart:      pchart,
		xAxisRange: xrange,
		canvasE:    *canvasE}

	// set canvas background color or leave it transparent
	if bgcolor.Alpha() != 0 {
		layer.canvasE.HTMLElement.AttributeStyleMap().Set("background-color", &typedom.Union{Value: js.ValueOf(bgcolor.Hexa())})
	}

	// to use a canvas we need to get a 2d or 3d contextto enable drawing, here we use a 2d context
	// https://developer.mozilla.org/fr/docs/Web/API/HTMLCanvasElement/getContext
	layer.Ctx2D = canvas.CanvasRenderingContext2DFromWrapper(canvasE.GetContext("2d", "alias:false"))
	if layer.Ctx2D == nil {
		log.Println("unable to get the canvas 2D context for drawing")
		return nil
	}
	return layer
}

// resize all layers according to the master element dimensions.
func (pchart *StockChart) resize() {

	// get the masterE dimensions
	cr := pchart.masterE.GetBoundingClientRect()
	masterx := int(cr.X())
	mastery := int(cr.Y())
	masterw := int(cr.Width())
	masterh := int(cr.Height())
	if cr.Width() <= 0 || cr.Height() <= 0 {
		fmt.Printf("chart %q no sizable", pchart.ID)
		return
	}

	const (
		sizenav    int = 70
		sizeyscale int = 80
		margin     int = 3
	)

	// relocate and resize every layers according to the master dimensions and their layout
	for _, layer := range pchart.layers {
		if layer == nil {
			continue
		}
		var x, y, w, h int
		switch layer.layout {
		case lAREA_FULL:
			x = masterx
			y = mastery
			w = masterw
			h = masterh

		case lAREA_NAVBAR:
			x = masterx
			y = mastery + masterh - sizenav
			w = masterw - sizeyscale
			h = int(sizenav)

		case lAREA_YSCALE:
			x = masterx + masterw - sizeyscale
			y = mastery
			w = int(sizeyscale)
			h = masterh - sizenav - margin

		case lAREA_GRAPH:
			x = masterx
			y = mastery
			w = masterw - sizeyscale
			h = masterh - sizenav - margin
		}
		newarea := Rect{O: Point{X: x, Y: y}, Width: w, Height: h}
		layer.Resize(newarea)
	}
}

// Redraw all layers (canvas) of the stockchart.
//
// Do not need to be called after a resize as layers automatically redrawn themselves
func (pchart *StockChart) Redraw() {
	if pchart.isDrawing {
		fmt.Println("new redraw request canceled. drawing still in progress")
		return
	}
	pchart.isDrawing = true
	for _, player := range pchart.layers {
		if player != nil {
			player.Redraw()
		}
	}
	pchart.isDrawing = false
}

// DoChangeTimeSelection updates all drawings to reflect the new timsel.
//
// It's called by the time selector in the navbar when user navigates,
// but can be called directly outside of the chart.
func (pchart *StockChart) DoChangeTimeSelection() {
	if pchart.isDrawing {
		fmt.Println("changeTimeSelection request canceled. drawing still in progress")
		return
	}

	pchart.isDrawing = true
	for _, player := range pchart.layers {
		if player != nil {
			player.OnChangeTimeSelection()
		}
	}
	pchart.isDrawing = false
}

/*
 * Utilities
 */

// getChartElement looks for chartid in the DOM and check it's type <stockchart>
func getChartElement(chartid string) (*dom.Element, error) {
	doc := webapi.GetWindow().Document()
	if doc == nil {
		return nil, fmt.Errorf("unable to access the html page content")
	}

	element := doc.GetElementById(chartid)
	if element == nil {
		return nil, fmt.Errorf("unable to find %q drawing area element in your html page", chartid)
	}

	strname := strings.ToLower(element.NodeName())
	if strname != "stockchart" {
		return nil, fmt.Errorf("drawing area element is not a <stockchart>, it should: %s\n", strname)
	}

	return element, nil
}

// get Mouse position taking into account DevicePixelRatio in case of browser zoom
func getMouseXY(event *htmlevent.MouseEvent) (xy Point) {
	dpr := webapi.GetWindow().DevicePixelRatio()
	dx := float64(event.OffsetX()) * dpr
	dy := float64(event.OffsetY()) * dpr
	xy = Point{X: int(dx), Y: int(dy)}
	return xy
}
