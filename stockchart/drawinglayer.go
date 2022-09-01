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
	"github.com/sunraylab/timeline/timeslice"
)

type layerArea int

const (
	lAREA_FULL layerArea = iota
	lAREA_GRAPH
	lAREA_YSCALE
	lAREA_NAVBAR
)

// A drawingLayer correspond to a single canvas with an html5 2D drawing context
// embedding one or many drawings
type drawingLayer struct {
	layerId  string
	canvasE  *canvas.HTMLCanvasElement
	layout   layerArea
	ctx2D    *canvas.CanvasRenderingContext2D
	clipArea Rect

	drawings []*Drawing //TimeSeriesDrawer
}

func (layer drawingLayer) String() string {
	str := fmt.Sprintf("%q, canvasE:%p, ctx2D:%p area:%v drawings:%d", layer.layerId, layer.canvasE, layer.ctx2D, layer.clipArea, len(layer.drawings))
	return str
}

func (layer *drawingLayer) SetMouseHandlers(pchart *StockChart) {

	// Define functions to capture mouse events on this layer,
	// only if the layer contains at least one mouse function on its drawings
	he := layer.HandledEvents()
	fmt.Printf("layer %q he=%b\n", layer, he)
	if (he & evt_MouseDown) != 0 {
		layer.canvasE.SetOnMouseDown(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMousePos(event)
			if xy.IsIn(layer.clipArea) {
				for _, drawing := range layer.drawings {
					if drawing.OnMouseDown != nil {
						drawing.OnMouseDown(layer, xy, event)
					}
				}
			}
		})
	}

	if (he & evt_MouseUp) != 0 {
		layer.canvasE.SetOnMouseUp(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMousePos(event)
			var timesel *timeslice.TimeSlice
			for _, drawing := range layer.drawings {
				if drawing.OnMouseUp != nil {
					tsel := drawing.OnMouseUp(layer, xy, event)
					if tsel != nil {
						timesel = tsel
					}
				}
			}
			// propagate change in time selection to all layers and all drawings
			if timesel != nil {
				pchart.ChangeTimeSelection(*timesel)
			}
		})
	}

	if (he & evt_MouseMove) != 0 {

		layer.canvasE.SetOnMouseMove(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
			xy := getMousePos(event)
			fRedraw := false
			for _, drawing := range layer.drawings {
				if drawing.OnMouseMove != nil {
					if drawing.OnMouseMove(layer, xy, event) {
						fRedraw = true
					}
				}
			}
			if fRedraw {
				layer.Redraw()
			}
		})
	}

	if (he & evt_Wheel) != 0 {
		layer.canvasE.SetOnWheel(func(event *htmlevent.WheelEvent, currentTarget *html.HTMLElement) {
			var timesel *timeslice.TimeSlice
			for _, drawing := range layer.drawings {
				if drawing.OnMouseUp != nil {
					tsel := drawing.OnWheel(layer, event)
					if tsel != nil {
						timesel = tsel
					}
				}
			}
			// propagate change in time selection to all layers and all drawings
			if timesel != nil {
				pchart.ChangeTimeSelection(*timesel)
			}
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
func (layer *drawingLayer) Resize(area Rect) {
	// TODO: do not redraw if the size has not changed

	// resize the canvas HTML element
	stylemap := layer.canvasE.HTMLElement.AttributeStyleMap()
	stylemap.Set("left", &typedom.Union{Value: js.ValueOf(fmt.Sprintf("%dpx", area.O.X))})
	stylemap.Set("top", &typedom.Union{Value: js.ValueOf(fmt.Sprintf("%dpx", area.O.Y))})
	stylemap.Set("width", &typedom.Union{Value: js.ValueOf(fmt.Sprintf("%dpx", area.Width))})
	stylemap.Set("height", &typedom.Union{Value: js.ValueOf(fmt.Sprintf("%dpx", area.Height))})

	// resize the drawing buffer of the canvas
	dpr := webapi.GetWindow().DevicePixelRatio()
	dw := math.Abs(float64(area.Width) * dpr)
	dh := math.Abs(float64(area.Height) * dpr)
	dbuffwidth := int(dw)
	dbuffheight := int(dh)
	layer.canvasE.SetWidth(uint(dbuffwidth))
	layer.canvasE.SetHeight(uint(dbuffheight))

	// update the cliparea
	layer.clipArea.O.X = 0
	layer.clipArea.O.X = 0
	layer.clipArea.Width = dbuffwidth
	layer.clipArea.Height = dbuffheight

	// fmt.Printf("Resizing layer %v ", player.String()) //DEBUG:
	// fmt.Printf("dpr=%f drawbuffw=%v, drawbuffh=%v\n", dpr, dbuffwidth, dbuffheight) //DEBUG:
	layer.Redraw()
}

// Clear the layer
func (layer *drawingLayer) Clear() {
	layer.ctx2D.ClearRect(float64(layer.clipArea.O.X), float64(layer.clipArea.O.Y), float64(layer.clipArea.Width), float64(layer.clipArea.Height))
}

// Clear the layer and Redraw all drawings
func (layer *drawingLayer) Redraw() {
	layer.Clear()
	for _, drawing := range layer.drawings {
		if drawing.OnRedraw != nil {
			//fmt.Printf("layer:%15s drawing:%15s OnRedraw\n", layer.layerId, drawing.Name) // DEBUG:
			drawing.OnRedraw(layer)
		}
	}
}

// Update chart selection to all drawings on the layer
func (layer *drawingLayer) ChangeTimeSelection(timesel timeslice.TimeSlice) {
	if layer.layout != lAREA_NAVBAR {
		layer.Clear()
		for _, drawing := range layer.drawings {
			if drawing.OnChangeTimeSelection != nil {
				//fmt.Printf("layer:%15s drawing:%15s OnSelectionChange--> Xsel=%v\n", layer.layerId, drawing.Name, timesel) // DEBUG:
				drawing.OnChangeTimeSelection(layer, timesel)
			}
		}
	}
}

type evtHandler int

const (
	evt_MouseUp   evtHandler = 0b00000001
	evt_MouseDown evtHandler = 0b00000010
	evt_MouseMove evtHandler = 0b00000100
	evt_Wheel     evtHandler = 0b00001000
)

func (layer *drawingLayer) HandledEvents() evtHandler {
	var e evtHandler
	for _, drawing := range layer.drawings {
		if drawing.OnMouseUp != nil {
			e |= evt_MouseUp
		}
		if drawing.OnMouseDown != nil {
			e |= evt_MouseDown
		}
		if drawing.OnMouseMove != nil {
			e |= evt_MouseMove
		}
		if drawing.OnWheel != nil {
			e |= evt_Wheel
		}
	}
	return e
}
