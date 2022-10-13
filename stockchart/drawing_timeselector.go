package stockchart

import (
	"log"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/css/typedom"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/gowebapi/webapi/html/htmlevent"
	"github.com/sunraylab/rgb/v2/bootstrapcolor.go"
	"github.com/sunraylab/timeline/v2"
)

type DrawingTimeSelector struct {
	Drawing

	buttonFrom               Rect // coordinates of the button within the cliparea
	buttonTo                 Rect // coordinates of the button within the cliparea
	isResizeCursor           bool // is the cursor resize ?
	dragFrom, dragTo, dragIn bool

	timeSelection timeline.TimeSlice // locally updated. The chart.timeSelection is only updated when the drag end

	MinZoomTime time.Duration // minute by defailt, can be changed
}

// Drawing factory
func NewDrawingTimeSelector(series *DataList) *DrawingTimeSelector {
	drawing := new(DrawingTimeSelector)
	drawing.Name = "timeselector"
	drawing.series = series
	drawing.MainColor = bootstrapcolor.Blue.Lighten(0.5)
	drawing.buttonFrom.Width = 8
	drawing.buttonFrom.Height = 30
	drawing.buttonTo.Width = 8
	drawing.buttonTo.Height = 30
	drawing.MinZoomTime = time.Minute

	drawing.Drawing.OnRedraw = func() {
		drawing.OnRedraw()
	}
	drawing.Drawing.OnMouseDown = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.OnMouseDown(xy, event)
	}
	drawing.Drawing.OnMouseUp = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.OnMouseUp(xy, event)
	}
	drawing.Drawing.OnMouseMove = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.OnMouseMove(xy, event)
	}
	drawing.Drawing.OnMouseLeave = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.OnMouseUp(xy, event)
	}
	drawing.Drawing.OnWheel = func(event *htmlevent.WheelEvent) {
		drawing.OnWheel(event)
	}
	// NeddRedraw always false, the time selector is redrawn only by user interecting on it
	return drawing
}

// OnRedrawTimeSelector draws the timeslice selector on top of the navbar.
// Buttons's position are updated to make it easy to catch them during a mouse event.
func (drawing *DrawingTimeSelector) OnRedraw() {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		// log.Printf("OnRedraw %q fails: unable to proceed given data", drawing.Name) // DEBUG:
		return
	}

	// init time sel on first redraw
	// if drawing.timeSelection.IsZero() {
	drawing.timeSelection = drawing.Layer.selectedTimeSlice
	// }

	// get the y center for redrawing the buttons
	ycenter := float64(drawing.ClipArea.O.Y) + float64(drawing.ClipArea.Height)/2.0

	// draw the left selector
	xleftrate := drawing.xAxisRange.Progress(drawing.selectedTimeSlice.From)
	xposleft := float64(drawing.ClipArea.O.X) + float64(drawing.ClipArea.Width)*xleftrate
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Opacify(0.4).Hexa())})
	drawing.Ctx2D.FillRect(float64(drawing.ClipArea.O.X), float64(drawing.ClipArea.O.Y), xposleft, float64(drawing.ClipArea.Height))
	moveButton(drawing, &drawing.buttonFrom, xposleft, ycenter)

	// draw the right selector
	xrightrate := drawing.xAxisRange.Progress(drawing.selectedTimeSlice.To)
	xposright := float64(drawing.ClipArea.O.X) + float64(drawing.ClipArea.Width)*xrightrate
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Opacify(0.4).Hexa())})
	drawing.Ctx2D.FillRect(xposright, float64(drawing.ClipArea.O.Y), float64(drawing.ClipArea.Width)-xposright, float64(drawing.ClipArea.Height))
	moveButton(drawing, &drawing.buttonTo, xposright, ycenter)
}

// moveButton utility
func moveButton(layer *DrawingTimeSelector, button *Rect, xcenter float64, ycenter float64) {
	layer.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(bootstrapcolor.Gray.Hexa())})
	layer.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(bootstrapcolor.Gray.Hexa())})
	layer.Ctx2D.SetLineWidth(1)
	x0 := xcenter - float64(button.Width)/2
	y0 := ycenter - float64(button.Height)/2
	layer.Ctx2D.FillRect(x0, y0, float64(button.Width), float64(button.Height))
	layer.Ctx2D.StrokeRect(x0, y0, float64(button.Width), float64(button.Height))
	button.O.X = int(x0)
	button.O.Y = int(y0)
}

// OnMouseDown starts dragging buttons
func (drawing *DrawingTimeSelector) OnMouseDown(xy Point, event *htmlevent.MouseEvent) {
	//fmt.Printf("%q mousedown xy:%v, frombutton:%v, tobutton:%v\n", drawing.Name, xy, drawing.fromButton, drawing.toButton) //DEBUG:

	// if already dragging and reenter into the canvas
	if drawing.dragFrom || drawing.dragTo {
		drawing.dragIn = true
	} else {
		drawing.dragIn = false
	}

	// do someting only if we're on a button
	if xy.IsIn(drawing.buttonFrom) {
		drawing.dragFrom = true
	} else if xy.IsIn(drawing.buttonTo) {
		drawing.dragTo = true
	}
}

// OnMouseUp Stops Dragging and update the chart.timeSelection.
//
// If the timeselection changes then the event dispatcher will call OnChangeTimeSelection on all drawings of all layers.
func (drawing *DrawingTimeSelector) OnMouseUp(xy Point, event *htmlevent.MouseEvent) {
	//fmt.Printf("%q mouseup xy:%v\n", drawing.Name, xy) //DEBUG:

	// update the chart time selection
	drawing.Layer.selectedTimeSlice = drawing.timeSelection

	// reset drag flag
	drawing.dragFrom = false
	drawing.dragTo = false
}

// OnMouseMove change the cursor when hovering a button, or change the local timeslice selection if dragging the buttons.
//
// If the timeselection changes then the event dispatcher will call OnChangeTimeSelection on all drawings of all layers.
func (drawing *DrawingTimeSelector) OnMouseMove(xy Point, event *htmlevent.MouseEvent) {
	//fmt.Printf("%q mousemove xy:%v\n", pdrawing.Name xy) //DEBUG:
	if drawing.xAxisRange == nil {
		log.Printf("OnMouseMove %q fails: missing data", drawing.Name)
		return
	}

	// change cursor if we start overing a button
	if (xy.IsIn(drawing.buttonFrom) || xy.IsIn(drawing.buttonTo)) && !drawing.isResizeCursor {
		drawing.Ctx2D.Canvas().AttributeStyleMap().Set("cursor", &typedom.Union{Value: js.ValueOf(`col-resize`)})
		drawing.isResizeCursor = true
	}

	// change cursor if we leave a button
	if (!xy.IsIn(drawing.buttonFrom) && !xy.IsIn(drawing.buttonTo)) && drawing.isResizeCursor {
		drawing.Ctx2D.Canvas().AttributeStyleMap().Set("cursor", &typedom.Union{Value: js.ValueOf(`auto`)})
		drawing.isResizeCursor = false
	}

	// change the selector
	if drawing.dragFrom {
		rate := drawing.ClipArea.XRate(xy.X)
		postime := drawing.xAxisRange.WhatTime(rate)
		if postime.Before(drawing.Layer.selectedTimeSlice.To.Add(-drawing.MinZoomTime)) {
			drawing.timeSelection.FromMove(postime, true)
			drawing.Redraw()
		}
	} else if drawing.dragTo {
		rate := drawing.ClipArea.XRate(xy.X)
		postime := drawing.xAxisRange.WhatTime(rate)
		if postime.After(drawing.Layer.selectedTimeSlice.From.Add(drawing.MinZoomTime)) {
			drawing.timeSelection.ToMove(postime, true)
			drawing.Redraw()
		}
	}
}

// OnWheel manage zoom and shifting the time selection
func (drawing *DrawingTimeSelector) OnWheel(event *htmlevent.WheelEvent) {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		return
	}

	tsel := &drawing.timeSelection

	// define a good timestep: 20% of the current duration
	timeStep := timeline.Nanoseconds(float64(tsel.Duration().Duration) * float64(0.2))
	//fmt.Printf("wheel shiftK=%v, dy=%f, timeStep=%v\n", event.ShiftKey(), dy, timeStep) // DEBUG:

	// get the wheel move
	dy := event.DeltaY()
	if event.ShiftKey() {
		// move mode: shift the selection
		if dy < 0 {
			// move to the future
			d := tsel.Duration()
			tsel.To = tsel.To.Add(timeStep.Duration)
			tsel.From = tsel.From.Add(timeStep.Duration)
			if tsel.To.After(drawing.xAxisRange.To) {
				tsel.To = drawing.xAxisRange.To
				tsel.From = tsel.To.Add(-d.Duration)
			}
			if tsel.From.After(drawing.series.Head.From) {
				tsel.From = drawing.series.Head.From
			}
		} else if dy > 0 {
			// move to the past
			d := tsel.Duration()
			tsel.From = tsel.From.Add(-timeStep.Duration)
			tsel.To = tsel.To.Add(-timeStep.Duration)
			if tsel.From.Before(drawing.xAxisRange.From) {
				tsel.From = drawing.xAxisRange.From
				tsel.To = tsel.From.Add(d.Duration)
			}
		}
	} else {
		// zoom mode: move the 'From' time only
		if dy > 0 {
			// zoom+ > reduce the timeslice, but no more than the tsel.To duration
			//fmt.Printf("zoom- mind=%s, at:%s, timeStep=%v\n", mind, at, timeStep) // DEBUG:
			tsel.From = tsel.From.Add(timeStep.Duration)
			if tsel.From.Add(drawing.MinZoomTime).After(tsel.To) {
				tsel.From = tsel.To.Add(-drawing.MinZoomTime)
			}
			if tsel.From.After(drawing.series.Head.From) {
				tsel.From = drawing.series.Head.From
			}
		} else if dy < 0 {
			// zoom- > enlarge the timeslice
			//fmt.Printf("zoom+ timeStep=%v\n", timeStep) // DEBUG:
			tsel.From = tsel.From.Add(-timeStep.Duration)
			if tsel.From.Before(drawing.xAxisRange.From) {
				tsel.From = drawing.xAxisRange.From
			}
		}
	}

	// update the chart time selection
	drawing.Layer.selectedTimeSlice = *tsel

	// redraw the timeselector only !
	//drawing.Redraw() // HACK:
}
