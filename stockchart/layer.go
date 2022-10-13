package stockchart

import (
	"fmt"
	"math"
	"strings"

	"github.com/gowebapi/webapi"
	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/css/typedom"
	"github.com/gowebapi/webapi/html"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/gowebapi/webapi/html/htmlevent"
	"github.com/sunraylab/rgb/v2"
	"github.com/sunraylab/timeline/v2"
)

type layoutT int

const (
	lAREA_FULL layoutT = iota
	lAREA_GRAPH
	lAREA_YSCALE
	lAREA_NAVBAR
)

// A Layer correspond to a single canvas with an html5 2D drawing context.
// It's embedding a stack of drawings.
type Layer struct {
	ClipArea Rect
	Ctx2D    *canvas.CanvasRenderingContext2D

	Name      string                   // the name of the layer, for debugging purpose
	chart   *StockChart              // the parent chart
	canvasE canvas.HTMLCanvasElement // the <canvas> element of this layer
	layout  layoutT                  // The layout of this layer within the chart

	xAxisRange *timeline.TimeSlice // the timeslice to show and draw on this layer
	drawings   []*Drawing          // stack of drawings
}

func NewLayer(id string, chart *StockChart, layout layoutT, xaxisrange *timeline.TimeSlice, canvasE canvas.HTMLCanvasElement) *Layer {
	layer := new(Layer)
	layer.Name = strings.ToLower(strings.Trim(id, " "))
	layer.chart = chart
	layer.layout = layout
	layer.xAxisRange = xaxisrange
	layer.canvasE = canvasE
	return layer
}

// AddDrawing add a new drawing to the stack of drawings appearing on this layer.
//
// The bgcolor is only used by the drawing.
func (layer *Layer) AddDrawing(d *Drawing, bgcolor rgb.Color) {
	d.Layer = layer
	d.BackgroundColor = bgcolor
	layer.drawings = append(layer.drawings, d)
}

// Default string interface
func (layer Layer) String() string {
	str := fmt.Sprintf("%q canvasE:%p, ctx2D:%p area:{%v} nb drawings:%d", layer.Name, &layer.canvasE, layer.Ctx2D, layer.ClipArea, len(layer.drawings))
	return str
}

// SetEventDispatcher activates mouse event handler on the canvas of the layer.
// When setup, if the mouse is located in the cliparea of the layer,
// the event is propagated to all drawings having defined their own MouseEvent func.
//
// Usually SetEventDispatcher is called by the chart factory.
func (layer *Layer) SetEventDispatcher() {

	// update changed time selection
	var oldselts timeline.TimeSlice
	var oldseldata *DataStock
	processSelChange := func() {
		if oldselts.Compare(layer.chart.selectedTimeSlice) == timeline.DIFFERENT {
			layer.chart.DoSelChangeTimeSlice(layer.chart.MainSeries.Name, layer.chart.selectedTimeSlice, true)
		}
		if oldseldata != layer.chart.selectedData {
			layer.chart.DoSelChangeData(layer.chart.MainSeries.Name, layer.chart.selectedData, true)
		}
	}

	// Define functions to capture mouse events on this layer,
	// only if the layer contains at least one mouse function on its drawings
	hme := layer.HandledEvents()

	Debug(DBG_EVENT, fmt.Sprintf("%q layer, SetEventDispatcher event handled=%08b ", layer.Name, hme))

	if (hme & evt_MouseDown) != 0 {
		layer.canvasE.SetOnMouseDown(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			if xy.IsIn(layer.ClipArea) {
				oldselts = layer.chart.selectedTimeSlice
				oldseldata = layer.chart.selectedData
				for _, drawing := range layer.drawings {
					if drawing.OnMouseDown != nil {
						drawing.OnMouseDown(xy, event)
					}
				}
				processSelChange()
			}
		})
	}

	if (hme & evt_MouseUp) != 0 {
		layer.canvasE.SetOnMouseUp(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			oldselts = layer.chart.selectedTimeSlice
			oldseldata = layer.chart.selectedData
			for _, drawing := range layer.drawings {
				if drawing.OnMouseUp != nil {
					drawing.OnMouseUp(xy, event)
				}
			}
			processSelChange()
		})
	}

	if (hme & evt_MouseMove) != 0 {
		layer.canvasE.SetOnMouseMove(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			oldselts = layer.chart.selectedTimeSlice
			oldseldata = layer.chart.selectedData
			for _, drawing := range layer.drawings {
				if drawing.OnMouseMove != nil {
					drawing.OnMouseMove(xy, event)
				}
			}
			processSelChange()
		})
	}

	if (hme & evt_MouseEnter) != 0 {
		layer.canvasE.SetOnMouseEnter(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			oldselts = layer.chart.selectedTimeSlice
			oldseldata = layer.chart.selectedData
			for _, drawing := range layer.drawings {
				if drawing.OnMouseEnter != nil {
					drawing.OnMouseEnter(xy, event)
				}
			}
			processSelChange()
		})
	}

	if (hme & evt_MouseLeave) != 0 {
		layer.canvasE.SetOnMouseLeave(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			oldselts = layer.chart.selectedTimeSlice
			oldseldata = layer.chart.selectedData
			for _, drawing := range layer.drawings {
				if drawing.OnMouseLeave != nil {
					drawing.OnMouseLeave(xy, event)
				}
			}
			processSelChange()
		})
	}

	if (hme & evt_Wheel) != 0 {
		layer.canvasE.SetOnWheel(func(event *htmlevent.WheelEvent, currentTarget *html.HTMLElement) {
			oldselts = layer.chart.selectedTimeSlice
			oldseldata = layer.chart.selectedData
			for _, drawing := range layer.drawings {
				if drawing.OnWheel != nil {
					drawing.OnWheel(event)
				}
			}
			Debug(DBG_SELCHANGE, fmt.Sprintf("OnWheel dispatcher: last %s, new %s", oldselts.String(), layer.chart.selectedTimeSlice.String() ))
			processSelChange()
		})
	}

	if (hme & evt_Click) != 0 {
		layer.canvasE.SetOnClick(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			oldselts = layer.chart.selectedTimeSlice
			oldseldata = layer.chart.selectedData
			for _, drawing := range layer.drawings {
				if drawing.OnClick != nil {
					drawing.OnClick(xy, event)
				}
			}
			processSelChange()
		})
	}
}

// Resize resize the drawing buffer according to the canvas element size.
// We need to extend the drawingbuffer to the same size of the canvas HTML element
// to avoid blurry effect, and take into account the DevicePixelRatio.
// It's more accurate to use GetBoundingClientRect than ClientWidth(),
// unfortunately thers'no easy way to take into account border size and padding withing the canvas element,
// so when resizing we assume there's no border, no padding and no margin within the canvas.
// https://webglfundamentals.org/webgl/lessons/webgl-resizing-the-canvas.html
//
// Resize calls automatically redraw
func (layer *Layer) Resize(newarea Rect) {
	// resize the canvas HTML element
	stylemap := layer.canvasE.HTMLElement.AttributeStyleMap()
	stylemap.Set("left", &typedom.Union{Value: js.ValueOf(fmt.Sprintf("%dpx", newarea.O.X))})
	stylemap.Set("top", &typedom.Union{Value: js.ValueOf(fmt.Sprintf("%dpx", newarea.O.Y))})
	stylemap.Set("width", &typedom.Union{Value: js.ValueOf(fmt.Sprintf("%dpx", newarea.Width))})
	stylemap.Set("height", &typedom.Union{Value: js.ValueOf(fmt.Sprintf("%dpx", newarea.Height))})

	// resize the drawing buffer of the canvas
	dpr := webapi.GetWindow().DevicePixelRatio()
	dw := math.Abs(float64(newarea.Width) * dpr)
	dh := math.Abs(float64(newarea.Height) * dpr)
	dbuffwidth := int(dw)
	dbuffheight := int(dh)
	layer.canvasE.SetWidth(uint(dbuffwidth))
	layer.canvasE.SetHeight(uint(dbuffheight))

	// update the cliparea
	layer.ClipArea.O.X = 0
	layer.ClipArea.O.X = 0
	layer.ClipArea.Width = dbuffwidth
	layer.ClipArea.Height = dbuffheight

	
	Debug(DBG_RESIZE, fmt.Sprintf("%q layer, Resize dpr=%f drawbuffw=%v, drawbuffh=%v", layer.Name, dpr, dbuffwidth, dbuffheight))

	// TODO: do not redraw if the size has not changed
	layer.Redraw()
}

// Clear the layer
func (layer *Layer) Clear() {
	layer.Ctx2D.ClearRect(float64(layer.ClipArea.O.X), float64(layer.ClipArea.O.Y), float64(layer.ClipArea.Width), float64(layer.ClipArea.Height))
}

// Clear the layer and redraw all drawings.
func (layer *Layer) Redraw() {
	layer.Clear()
	for _, drawing := range layer.drawings {
		if drawing.OnRedraw != nil {
			drawing.OnRedraw()
		}
	}
}

// Redraw the layer if at least one drawings need to be redrawn.
// Then update selection on the layer
func (layer *Layer) RedrawOnlyNeeds() {
	for _, drawing := range layer.drawings {
		if drawing.NeedRedraw != nil {
			need := drawing.NeedRedraw()

			Debug(DBG_SELCHANGE | DBG_REDRAW, fmt.Sprintf("%q/%q layer/drawing, RedrawOnlyNeeds NeedRedraw:%v", layer.Name, drawing.Name, need))

			if need {
				layer.Redraw()
				break
			}
		}
	}
}
