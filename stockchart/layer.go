package stockchart

import (
	"fmt"
	"math"

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

	id      string                   // the id of the layer, for debugging purpose
	chart   *StockChart              // the parent chart
	canvasE canvas.HTMLCanvasElement // the <canvas> element of this layer
	layout  layoutT                  // The layout of this layer within the chart

	xAxisRange *timeline.TimeSlice // the timeslice to show and draw on this layer
	drawings   []*Drawing          // stack of drawings
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
	str := fmt.Sprintf("%q canvasE:%p, ctx2D:%p area:{%v} nb drawings:%d", layer.id, &layer.canvasE, layer.Ctx2D, layer.ClipArea, len(layer.drawings))
	return str
}

// SetEventDispatcher activates mouse event handler on the canvas of the layer.
// When setup, if the mouse is located in the cliparea of the layer,
// the event is propagated to all drawings having defined their own MouseEvent func.
//
// Usually SetEventDispatcher is called by the chart factory.
func (layer *Layer) SetEventDispatcher() {

	// update changed time selection
	var oldtimesel timeline.TimeSlice
	processTimeSelection := func() {
		if oldtimesel.Compare(layer.chart.timeSelection) == timeline.DIFFERENT {
			layer.chart.DoChangeTimeSelection(layer.chart.MainSeries.Name)
		}
	}

	var olddataSelected *DataStock
	processSelectData := func() {
		if olddataSelected != layer.chart.dataSelected {
			layer.chart.DoSelectData(layer.chart.MainSeries.Name)
		}
	}

	// Define functions to capture mouse events on this layer,
	// only if the layer contains at least one mouse function on its drawings
	hme := layer.HandledEvents()
	fmt.Printf("layer %q, mouse event handled=%08b\n", layer.id, hme) // DEBUG:

	if (hme & evt_MouseDown) != 0 {
		layer.canvasE.SetOnMouseDown(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			if xy.IsIn(layer.ClipArea) {
				oldtimesel = layer.chart.timeSelection
				for _, drawing := range layer.drawings {
					if drawing.OnMouseDown != nil {
						drawing.OnMouseDown(xy, event)
					}
				}
				processTimeSelection()
			}
		})
	}

	if (hme & evt_MouseUp) != 0 {
		layer.canvasE.SetOnMouseUp(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			oldtimesel = layer.chart.timeSelection
			for _, drawing := range layer.drawings {
				if drawing.OnMouseUp != nil {
					drawing.OnMouseUp(xy, event)
				}
			}
			processTimeSelection()
		})
	}

	if (hme & evt_MouseMove) != 0 {
		layer.canvasE.SetOnMouseMove(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			oldtimesel = layer.chart.timeSelection
			for _, drawing := range layer.drawings {
				if drawing.OnMouseMove != nil {
					drawing.OnMouseMove(xy, event)
				}
			}
			processTimeSelection()
		})
	}

	if (hme & evt_MouseEnter) != 0 {
		layer.canvasE.SetOnMouseEnter(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			oldtimesel = layer.chart.timeSelection
			for _, drawing := range layer.drawings {
				if drawing.OnMouseEnter != nil {
					drawing.OnMouseEnter(xy, event)
				}
			}
			processTimeSelection()
		})
	}

	if (hme & evt_MouseLeave) != 0 {
		layer.canvasE.SetOnMouseLeave(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			oldtimesel = layer.chart.timeSelection
			for _, drawing := range layer.drawings {
				if drawing.OnMouseLeave != nil {
					drawing.OnMouseLeave(xy, event)
				}
			}
			processTimeSelection()
		})
	}

	if (hme & evt_Wheel) != 0 {
		layer.canvasE.SetOnWheel(func(event *htmlevent.WheelEvent, currentTarget *html.HTMLElement) {
			for _, drawing := range layer.drawings {
				oldtimesel = layer.chart.timeSelection
				if drawing.OnMouseUp != nil {
					drawing.OnWheel(event)
				}
			}
			processTimeSelection()
		})
	}

	if (hme & evt_Click) != 0 {
		layer.canvasE.SetOnClick(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMouseXY(event)
			olddataSelected = layer.chart.dataSelected
			for _, drawing := range layer.drawings {
				olddataSelected = layer.chart.dataSelected
				if drawing.OnClick != nil {
					drawing.OnClick(xy, event)
				}
			}
			processSelectData()
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

	// fmt.Printf("Resizing layer %v ", player.String()) //DEBUG:
	// fmt.Printf("dpr=%f drawbuffw=%v, drawbuffh=%v\n", dpr, dbuffwidth, dbuffheight) //DEBUG:

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
func (layer *Layer) OnChangeTimeSelection() {
	for _, drawing := range layer.drawings {
		if drawing.NeedRedraw != nil {
			need := drawing.NeedRedraw()
			//fmt.Printf(" drawing:%s, NeedRedraw:%v\n", drawing.id, need) // DEBUG:
			if need {
				layer.Redraw()
				break
			}
		}
	}
}

func (layer *Layer) OnSelectData() {
	for _, drawing := range layer.drawings {
		if drawing.NeedRedraw != nil {
			need := drawing.NeedRedraw()
			//fmt.Printf(" drawing:%s, NeedRedraw:%v\n", drawing.id, need) // DEBUG:
			if need {
				layer.Redraw()
				break
			}
		}
	}
}
