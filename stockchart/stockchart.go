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
	"github.com/sunraylab/verbose"

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

	MainSeries        DataList
	timeRange         timeline.TimeSlice // the overall time range to display
	selectedTimeSlice timeline.TimeSlice // the current time slice selected, IsZero if none
	selectedData      *DataStock         // the current data selected, nil if none
	localZone         bool               // Show local zone time, otherwise show UTC time

	NotifySelChangeTimeSlice func(strpair string, ts timeline.TimeSlice) // function called everytime the timeselection change, if not nil
	NotifySelChangeData      func(strpair string, data *DataStock)
}

// String interface for StockChart, mainly for debugging purpose
func (chart StockChart) String() string {
	str := fmt.Sprintf("%s\n", chart.ID)
	for _, player := range chart.layers {
		if player != nil {
			str += fmt.Sprintf("  layer %22s: %p\n", player.Name, player)
			for _, pdrawing := range player.drawings {
				if pdrawing != nil {
					str += fmt.Sprintf("    drawing %18s: %p series:%p %v\n", pdrawing.Name, pdrawing, &pdrawing.series, pdrawing.series)
				}
			}
		}
	}
	return str
}

// SetMainSeries set or reset the MainSeries of the chart and its drawings. Reset the timerange
func (pchart *StockChart) ResetMainSeries(series DataList, extendrate float64, redrawNow bool) {

	// clear subchart series
	for _, player := range pchart.layers {
		if player != nil {
			for _, pdraw := range player.drawings {
				if pdraw != nil {
					if pdraw.series != &pchart.MainSeries {
						pdraw.series = nil
					}
				}
			}
		}
	}

	// change the content
	pchart.MainSeries = series

	// reset time range
	pchart.SetTimeRange(pchart.MainSeries.TimeSlice(), extendrate)

	// Redraw, without subcharts
	if redrawNow {
		pchart.Redraw()
	}
}

// AddSubChart add another drawing to draw within the same X and Y ranges than the main series on the choosen layer.
// Drawing is made just before drawing of the main drawin on the layer.
// This drawing is associated to it's own series of data
func (pchart *StockChart) AddSubChart(layerid int, dr *Drawing) {
	verbose.Assert(layerid >= 0 && layerid <= 5, "wrong layer id")

	pchart.layers[layerid].AddDrawing(dr, rgb.None, false)
	dr.DrawArea = getMainDrawArea
}

// SetTimeRange defines the overall time range to display. Extend the end with extendCoef.
//
//	extendCoef == 0 no extension
//	extendCoef == 0.1 for 10% extention in duration
//
// Update timeselection if required. If the timeselection change the RedrawOnlyNeeds
// Returns the setup timerange
func (pchart *StockChart) SetTimeRange(timerange timeline.TimeSlice, extendrate float64) timeline.TimeSlice {

	Debug(DBG_SELCHANGE, "SetTimeRange %s", timerange)

	if timerange.Duration().IsFinite && extendrate > 0 {
		timerange.ExtendTo(timeline.Nanoseconds(float64(timerange.Duration().Duration) * extendrate).Duration)
	}

	pchart.timeRange.From = timerange.From
	pchart.timeRange.To = timerange.To

	selNeedUpdate := pchart.selectedTimeSlice.IsZero() || pchart.selectedTimeSlice.From.Before(pchart.timeRange.From) || pchart.selectedTimeSlice.To.After(pchart.timeRange.To)
	if selNeedUpdate {
		pchart.DoChangeSelTimeSlice(pchart.MainSeries.Name, pchart.timeRange, false)
	}
	return pchart.timeRange
}

// NewStockChart initialize a stockchart within the <stockchart> HTML element idenfied by chartid.
// An HTML page can have multiple <stockchart> but with different chartid. The layout of the chart is composed of multiples layers which are stacked canvas.
//
// Returns the stockchart created or an error if canvasid is not found.
func NewStockChart(chartid string, bgcolor rgb.Color, series DataList, extendrate float64) (*StockChart, error) {
	// some cleaning
	chartid = strings.ToLower(strings.Trim(chartid, " "))

	// Get the <stockchart> element who will embedd the chart
	stockchartE, err := getChartElement(chartid)
	if err != nil {
		return nil, err
	}

	// build the chart object
	chart := &StockChart{
		ID:         chartid,
		masterE:    stockchartE,
		MainSeries: series}

	// by default the timeMaxRange is the full timeslice + 10% to represents the future.
	chart.SetTimeRange(chart.MainSeries.TimeSlice(), extendrate)

	// create master layer layout, white bg and without drawings
	// the master layer covers all the <stockchart> element size, and is build first at the background
	if layer := chart.addNewLayer("0-bg", lAREA_FULL, bgcolor, nil); layer != nil {
		chart.layers[0] = layer
	}

	// the navbar layer, updated only when navXAxisRange change
	if layer := chart.addNewLayer("1-navbar", lAREA_NAVBAR, rgb.White, &chart.timeRange); layer != nil {
		dr := layer.AddDrawing(&NewDrawingSeries(&chart.MainSeries, true).Drawing, rgb.White, true)
		dr.DrawArea = func(cliparea Rect) Rect {
			area := cliparea
			area.O.Y += 5
			area.Height -= 5
			return area
		}
		layer.AddDrawing(&NewDrawingXGrid(&chart.MainSeries, true, false).Drawing, rgb.None, true)
		chart.layers[1] = layer
	}

	// the transparent time selector layer
	if layer := chart.addNewLayer("2-timeselector", lAREA_NAVBAR, rgb.None, &chart.timeRange); layer != nil {
		layer.AddDrawing(&NewDrawingTimeSelector(&chart.MainSeries).Drawing, rgb.None, true)
		layer.SetEventDispatcher()
		chart.layers[2] = layer
	}

	// the yscale layer
	if layer := chart.addNewLayer("3-yscale", lAREA_YSCALE, rgb.White, &chart.selectedTimeSlice); layer != nil {
		dr := layer.AddDrawing(&NewDrawingYGrid(&chart.MainSeries, true).Drawing, rgb.White, true)
		dr.DrawArea = getMainDrawArea
		chart.layers[3] = layer
	}

	// the chart layer
	if layer := chart.addNewLayer("4-chart", lAREA_GRAPH, rgb.White, &chart.selectedTimeSlice); layer != nil {
		// the background
		layer.AddDrawing(&NewDrawingBackground(&chart.MainSeries).Drawing, rgb.None, true)

		// the YGrid
		dr := layer.AddDrawing(&NewDrawingYGrid(&chart.MainSeries, false).Drawing, rgb.None, true)
		dr.DrawArea = getMainDrawArea

		// The XGrid
		layer.AddDrawing(&NewDrawingXGrid(&chart.MainSeries, false, true).Drawing, rgb.None, true)

		// The volume bars
		dr = layer.AddDrawing(&NewDrawingBars(&chart.MainSeries).Drawing, rgb.None, true)
		dr.DrawArea = func(cliparea Rect) Rect {
			area := cliparea.Shrink(0, 5)
			h := int(float64(area.Height) * 0.15) // draw bars at the bottom of the cliparea
			area.O.Y = area.O.Y + area.Height - h - 15
			area.Height = h
			return area
		}

		// The candles
		dr = layer.AddDrawing(&NewDrawingCandles(&chart.MainSeries, 1, false).Drawing, rgb.White, true)
		dr.DrawArea = getMainDrawArea

		chart.layers[4] = layer
	}

	// the hover transprent layer
	if layer := chart.addNewLayer("5-hover", lAREA_GRAPH, rgb.None, &chart.selectedTimeSlice); layer != nil {
		layer.AddDrawing(&NewDrawingHoverCandles(&chart.MainSeries).Drawing, rgb.White, true)
		layer.SetEventDispatcher()
		chart.layers[5] = layer
	}

	// Add event listener on resize event
	webapi.GetWindow().AddEventResize(func(event *htmlevent.UIEvent, win *webapi.Window) {
		// resizing the chart will resize and redraw every layers
		chart.Resize()
	})

	// size it the first time to force a full redraw
	//	chart.Resize()

	return chart, nil
}

func getMainDrawArea(cliparea Rect) Rect {
	area := cliparea.Shrink(0, 5)
	area.Height -= 15
	return area
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
	layer := NewLayer(strings.ToLower(strings.Trim(layerid, " ")), pchart, layout, xrange, *canvasE)

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
func (pchart *StockChart) Resize() {

	// get the masterE dimensions
	cr := pchart.masterE.GetBoundingClientRect()
	masterx := int(cr.X())
	mastery := int(cr.Y())
	masterw := int(cr.Width())
	masterh := int(cr.Height())
	if cr.Width() <= 0 || cr.Height() <= 0 {
		log.Printf("chart %q no sizable", pchart.ID)
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
		log.Println("/!\\ Redraw request canceled. drawing still in progress")
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

// Redraw all layers (canvas) of the stockchart.
//
// Do not need to be called after a resize as layers automatically redrawn themselves
func (pchart *StockChart) RedrawOnlyNeeds() {
	if pchart.isDrawing {
		log.Println("/!\\ RedrawOnlyNeeds request canceled. drawing still in progress")
		return
	}
	pchart.isDrawing = true
	for _, player := range pchart.layers {
		if player != nil {
			player.RedrawOnlyNeeds()
		}
	}
	pchart.isDrawing = false
}

// DoChangeSelTimeSlice updates all drawings to reflect the new timsel.
//
// It's called by the time selector in the navbar when user navigates,
// but can be called directly outside of the chart.
//
// call OnDoChangeTimeSelection if setup
func (pchart *StockChart) DoChangeSelTimeSlice(strpair string, newts timeline.TimeSlice, fNotify bool) {

	if newts.IsZero() {
		newts = pchart.timeRange
	} else {
		pchart.timeRange.BoundIn(&newts)
	}
	pchart.selectedTimeSlice = newts

	Debug(DBG_SELCHANGE, "DoChangeSelTimeSlice: %s", newts.String())

	pchart.RedrawOnlyNeeds()

	if pchart.NotifySelChangeTimeSlice != nil && fNotify {
		pchart.NotifySelChangeTimeSlice(strpair, newts)
	}
}

func (pchart *StockChart) DoChangeSelData(strpair string, newdata *DataStock, fNotify bool) {
	pchart.selectedData = newdata

	Debug(DBG_SELCHANGE, "DoChangeSelData: %p %s", newdata, newdata.String())

	pchart.RedrawOnlyNeeds()

	if pchart.NotifySelChangeData != nil && fNotify {
		pchart.NotifySelChangeData(strpair, pchart.selectedData)
	}
}

func (pchart *StockChart) DoChangeTimeZone(localZone bool) {
	pchart.localZone = localZone

	Debug(DBG_SELCHANGE, "DoChangeTimeZone: localzone:%v", localZone)

	pchart.RedrawOnlyNeeds()
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
