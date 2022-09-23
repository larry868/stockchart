// date-time axis
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

// StockChart is a 2D chart draw in an HTML5 canvas
type StockChart struct {
	ChartID string       // the identifier of this chart, the canvas id
	masterE *dom.Element // the master element containing the chart

	layers    [6]*drawingLayer // the 6 drawing layers composing a stockchart
	isDrawing bool             // flag signaling a drawing in progress
}

// String interface for StockChart, mainly for debugging purpose
func (chart StockChart) String() string {
	str := fmt.Sprintf("%s\n", chart.ChartID)
	for _, player := range chart.layers {
		if player != nil {
			str += fmt.Sprintf("  layer %22s: %p\n", player.Name, player)
			for _, pdrawing := range player.drawings {
				if pdrawing != nil {
					str += fmt.Sprintf("    drawing %18s: %p series:%p %v\n", pdrawing.Name, pdrawing, pdrawing.series, pdrawing.series)
				}
			}
		}
	}
	return str
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
	chart := &StockChart{ChartID: chartid, masterE: stockchartE}

	// The navbar and the selection shows the full dataset.
	// The selection will be updated but not the nav one.
	// Add 10% at the end to represents the future.
	navXAxisRange := series.TimeSlice()
	navXAxisRange.ToExtend(timeline.Duration(float64(*navXAxisRange.Duration()) * 0.1))
	selXAxisRange := navXAxisRange

	// create master layer layout, white bg and without drawings
	// the master layer covers all the <stockchart> element size, and is build first at the background
	if layer := chart.addnewLayer("0-bg", lAREA_FULL, bgcolor); layer != nil {
		chart.layers[0] = layer
	}

	// the navbar layer
	if layer := chart.addnewLayer("1-navbar", lAREA_NAVBAR, rgb.White); layer != nil {
		layer.AddDrawing(&NewDrawingSeries(series, &navXAxisRange, true).Drawing, rgb.White)
		layer.AddDrawing(&NewDrawingXGrid(series, &navXAxisRange, true).Drawing, rgb.None)
		chart.layers[1] = layer
	}

	// the transparent time selector layer
	if layer := chart.addnewLayer("2-timeselector", lAREA_NAVBAR, rgb.None); layer != nil {
		layer.AddDrawing(&NewDrawingTimeSelector(series, &navXAxisRange).Drawing, rgb.None)
		layer.SetMouseHandlers(chart)
		chart.layers[2] = layer
	}

	// the yscale layer
	if layer := chart.addnewLayer("3-yscale", lAREA_YSCALE, rgb.White); layer != nil {
		layer.AddDrawing(&NewDrawingYGrid(series, &selXAxisRange, true).Drawing, rgb.White)
		chart.layers[3] = layer
	}

	// the chart layer
	if layer := chart.addnewLayer("4-chart", lAREA_GRAPH, rgb.White); layer != nil {
		layer.AddDrawing(&NewDrawingBackground(series).Drawing, rgb.None)
		layer.AddDrawing(&NewDrawingYGrid(series, &selXAxisRange, false).Drawing, rgb.None)
		layer.AddDrawing(&NewDrawingXGrid(series, &selXAxisRange, false).Drawing, rgb.None)
		layer.AddDrawing(&NewDrawingCandles(series, &selXAxisRange).Drawing, rgb.White)
		chart.layers[4] = layer
	}

	// the hover transprent layer
	if layer := chart.addnewLayer("5-hover", lAREA_GRAPH, rgb.None); layer != nil {
		layer.AddDrawing(&NewDrawingHoverCandles(series, &selXAxisRange).Drawing, rgb.White)
		layer.SetMouseHandlers(chart)
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

// addnewLayer create a new canvas, inside the targetE div, then fill its bg color or leave it transparent.
// The layer embed the WEBGL 2D drawing context.
// It's added to the chart stack layers and moved and sized according to layoutStyle and master area
//
// return the chartLayer corresponding to this layer, or nil if an error occurs
func (pchart *StockChart) addnewLayer(id string, larea layerArea, bgcolor rgb.Color) *drawingLayer {
	// create a canvas
	domE := webapi.GetWindow().Document().CreateElement("canvas", &webapi.Union{Value: js.ValueOf("dom.Node")})
	domE.SetId("canvas" + pchart.ChartID + strings.ToLower(strings.Trim(id, " ")))
	newE := pchart.masterE.AppendChild(&domE.Node)
	canvasE := canvas.HTMLCanvasElementFromWrapper(newE)
	canvasE.AttributeStyleMap().Set("position", &typedom.Union{Value: js.ValueOf(`absolute`)})
	canvasE.AttributeStyleMap().Set("border", &typedom.Union{Value: js.ValueOf(`none`)})
	canvasE.AttributeStyleMap().Set("padding", &typedom.Union{Value: js.ValueOf(`0`)})
	canvasE.AttributeStyleMap().Set("margin", &typedom.Union{Value: js.ValueOf(`0`)})

	// create the layer
	layer := &drawingLayer{
		Name:    strings.ToLower(strings.Trim(id, " ")),
		layout:  larea,
		canvasE: canvasE}

	// set default canvas background color or leave it transparent
	if _, _, _, alpha := bgcolor.RGBA(); alpha != 0 {
		canvasE.HTMLElement.AttributeStyleMap().Set("background-color", &typedom.Union{Value: js.ValueOf(bgcolor.Hexa())})
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
		fmt.Printf("chart %q no sizable", pchart.ChartID)
		return
	}

	const (
		sizenav    int = 70
		sizeyscale int = 80
		margin     int = 3
	)

	// relocate and resize every layers according to the master dimensions
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
		area := Rect{O: Point{X: x, Y: y}, Width: w, Height: h}
		layer.Resize(area)
	}
}

// Redraw all the canvas of the stockchart.
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

// ChangeTimeSelection updates all layers to reflect the new timsel.
// It's called by the time selector in the navbar when user navigates,
// but can be called directly from outside of the chart.
func (pchart *StockChart) ChangeTimeSelection(timesel timeline.TimeSlice) {
	if pchart.isDrawing {
		fmt.Println("changeTimeSelection request canceled. drawing still in progress")
		return
	}
	pchart.isDrawing = true
	for _, player := range pchart.layers {
		if player != nil {
			player.ChangeTimeSelection(timesel)
		}
	}
	pchart.isDrawing = false
}

/*
 * Utilities
 */

// getChartElement look for chartid in the DOM and check this is sized element of type <stockchart>
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

func getMousePos(event *htmlevent.MouseEvent) (xy Point) {
	dpr := webapi.GetWindow().DevicePixelRatio()
	dx := float64(event.OffsetX()) * dpr
	dy := float64(event.OffsetY()) * dpr
	xy = Point{X: int(dx), Y: int(dy)}
	return xy
}
