package stockchart

import (
	"fmt"

	"github.com/gowebapi/webapi/html/htmlevent"
	"github.com/sunraylab/rgb"
	"github.com/sunraylab/timeline/timeslice"
)

type Drawing struct {
	Name      string    // drawing name, mainly for debugging, but culd be used in the drawing
	MainColor rgb.Color // optional

	series     *DataList            // the series of data to draw
	xAxisRange *timeslice.TimeSlice // the timeslice to show and draw
	xStartDrag timeslice.TimeSlice  // drag and drop data that can be used by mouse events

	OnRedraw              func(layer *drawingLayer)
	OnMouseDown           func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent)
	OnMouseUp             func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) (timesel *timeslice.TimeSlice)
	OnMouseMove           func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent)
	OnMouseLeave          func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent)
	OnChangeTimeSelection func(layer *drawingLayer, timesel timeslice.TimeSlice)
	OnWheel               func(layer *drawingLayer, event *htmlevent.WheelEvent) (timesel *timeslice.TimeSlice)
}

func (drawing Drawing) String() string {
	strxrange := "missing"
	if drawing.xAxisRange != nil {
		strxrange = drawing.xAxisRange.String()
	}
	str := fmt.Sprintf("%q xrange:%s", drawing.Name, strxrange)
	return str
}
