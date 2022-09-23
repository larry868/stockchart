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

	timeSelection timeline.TimeSlice

	MinZoomTime time.Duration // minute by defailt, can be changed
}

// Drawing factory
func NewDrawingTimeSelector(series *DataList, xrange *timeline.TimeSlice) *DrawingTimeSelector {
	drawing := new(DrawingTimeSelector)
	drawing.Name = "timeselector"
	drawing.series = series
	drawing.xAxisRange = xrange
	drawing.timeSelection = *xrange
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
	drawing.Drawing.OnMouseUp = func(xy Point, event *htmlevent.MouseEvent) (timesel *timeline.TimeSlice) {
		return drawing.OnMouseUp(xy, event)
	}
	drawing.Drawing.OnMouseMove = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.OnMouseMove(xy, event)
	}
	drawing.Drawing.OnWheel = func(event *htmlevent.WheelEvent) (timesel *timeline.TimeSlice) {
		return drawing.OnWheel(event)
	}
	return drawing
}

// OnRedrawTimeSelector draws the timeslice selector on top of the navbar.
// Buttons's position are updated to make it easy to catch them during a mouse event.
func (drawing *DrawingTimeSelector) OnRedraw() {
	if drawing.series == nil || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		log.Printf("OnRedraw %s fails: unable to proceed given data", drawing.Name)
		return
	}
	// get the y center for redrawing the buttons
	ycenter := float64(drawing.ClipArea.O.Y) + float64(drawing.ClipArea.Height)/2.0

	// draw the left selector
	xleftrate := drawing.xAxisRange.Progress(drawing.timeSelection.From)
	xposleft := float64(drawing.ClipArea.O.X) + float64(drawing.ClipArea.Width)*xleftrate
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Opacify(0.4).Hexa())})
	drawing.Ctx2D.FillRect(float64(drawing.ClipArea.O.X), float64(drawing.ClipArea.O.Y), xposleft, float64(drawing.ClipArea.Height))
	moveButton(drawing, &drawing.buttonFrom, xposleft, ycenter)

	// draw the right selector
	xrightrate := drawing.xAxisRange.Progress(drawing.timeSelection.To)
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
		drawing.xStartDrag = drawing.timeSelection
	} else if xy.IsIn(drawing.buttonTo) {
		drawing.dragTo = true
		drawing.xStartDrag = drawing.timeSelection
	}
}

// OnMouseUp returns the updated timeslice selected after moving buttons
func (drawing *DrawingTimeSelector) OnMouseUp(xy Point, event *htmlevent.MouseEvent) (timesel *timeline.TimeSlice) {
	//fmt.Printf("%q mouseup xy:%v\n", drawing.Name, xy) //DEBUG:

	if (drawing.dragFrom || drawing.dragTo) && (drawing.dragIn || drawing.xStartDrag.Equal(drawing.timeSelection) == 0) {
		timesel = new(timeline.TimeSlice)
		*timesel = drawing.timeSelection
	}
	drawing.dragFrom = false
	drawing.dragTo = false
	return timesel
}

// OnMouseMove change the cursor when hovering a button, or change the timeslice selection if dragging the buttons
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
		if postime.Before(drawing.timeSelection.To.Add(-drawing.MinZoomTime)) {
			drawing.timeSelection.FromMove(postime, true)
			drawing.Redraw()
		}

	} else if drawing.dragTo {
		rate := drawing.ClipArea.XRate(xy.X)
		postime := drawing.xAxisRange.WhatTime(rate)
		if postime.After(drawing.timeSelection.From.Add(drawing.MinZoomTime)) {
			drawing.timeSelection.ToMove(postime, true)
			drawing.Redraw()
		}
	}
}

// OnWheel manage zoom and shifting the time selection
func (drawing *DrawingTimeSelector) OnWheel(event *htmlevent.WheelEvent) (timesel *timeline.TimeSlice) {
	if drawing.series == nil || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		return
	}

	// define a good timestep
	currentSel := drawing.timeSelection
	timeStep := time.Duration(float64(*drawing.timeSelection.Duration()) * 0.2)
	//fmt.Printf("wheel shiftK=%v, dy=%f, timeStep=%v\n", event.ShiftKey(), dy, timeStep) // DEBUG:

	// get the wheel move
	dy := event.DeltaY()
	if event.ShiftKey() {
		// move mode: shift the selection
		if dy < 0 {
			// move to the future
			d := drawing.timeSelection.Duration()
			drawing.timeSelection.To = drawing.timeSelection.To.Add(timeStep)
			drawing.timeSelection.From = drawing.timeSelection.From.Add(timeStep)
			if drawing.timeSelection.To.After(drawing.xAxisRange.To) {
				drawing.timeSelection.To = drawing.xAxisRange.To
				drawing.timeSelection.From = drawing.timeSelection.To.Add(-time.Duration(*d))
			}
			if drawing.timeSelection.From.After(drawing.series.Head.From) {
				drawing.timeSelection.From = drawing.series.Head.From
			}
		} else if dy > 0 {
			// move to the past
			d := drawing.timeSelection.Duration()
			drawing.timeSelection.From = drawing.timeSelection.From.Add(-timeStep)
			drawing.timeSelection.To = drawing.timeSelection.To.Add(-timeStep)
			if drawing.timeSelection.From.Before(drawing.xAxisRange.From) {
				drawing.timeSelection.From = drawing.xAxisRange.From
				drawing.timeSelection.To = drawing.timeSelection.From.Add(time.Duration(*d))
			}
		}
	} else {
		// zoom mode: move the 'From' time only
		if dy > 0 {
			// zoom+ > reduce the timeslice, but no more than the drawing.timeSelection.To duration
			//fmt.Printf("zoom- mind=%s, at:%s, timeStep=%v\n", mind, at, timeStep) // DEBUG:
			drawing.timeSelection.From = drawing.timeSelection.From.Add(timeStep)
			if drawing.timeSelection.From.Add(drawing.MinZoomTime).After(drawing.timeSelection.To) {
				drawing.timeSelection.From = drawing.timeSelection.To.Add(-drawing.MinZoomTime)
			}
			if drawing.timeSelection.From.After(drawing.series.Head.From) {
				drawing.timeSelection.From = drawing.series.Head.From
			}
		} else if dy < 0 {
			// zoom- > enlarge the timeslice
			//fmt.Printf("zoom+ timeStep=%v\n", timeStep) // DEBUG:
			drawing.timeSelection.From = drawing.timeSelection.From.Add(-timeStep)
			if drawing.timeSelection.From.Before(drawing.xAxisRange.From) {
				drawing.timeSelection.From = drawing.xAxisRange.From
			}
		}
	}

	// redraw only if the time Selection has changed
	if currentSel.Equal(drawing.timeSelection) == 0 {
		timesel = new(timeline.TimeSlice)
		*timesel = drawing.timeSelection
		drawing.Clear()
		drawing.OnRedraw()
	}
	return timesel
}
