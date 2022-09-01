package stockchart

import (
	"log"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/css/typedom"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/gowebapi/webapi/html/htmlevent"

	"github.com/sunraylab/rgb/bootstrapcolor.go"
	"github.com/sunraylab/timeline/timeslice"
)

type DrawingTimeSelector struct {
	Drawing

	fromButton               Rect
	toButton                 Rect
	resizeCursor             bool
	dragFrom, dragTo, dragIn bool
	timeSelection            timeslice.TimeSlice
}

// Drawing factory
func NewDrawingTimeSelector(series *DataList, xrange *timeslice.TimeSlice) *DrawingTimeSelector {
	drawing := new(DrawingTimeSelector)
	drawing.Name = "timeselector"
	drawing.series = series
	drawing.xAxisRange = xrange
	drawing.timeSelection = *xrange
	drawing.MainColor = *bootstrapcolor.Blue.Lighten(0.5)
	drawing.fromButton.Width = 8
	drawing.fromButton.Height = 30
	drawing.toButton.Width = 8
	drawing.toButton.Height = 30
	drawing.Drawing.OnRedraw = func(layer *drawingLayer) {
		drawing.OnRedraw(layer)
	}
	drawing.Drawing.OnMouseDown = func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) {
		drawing.OnMouseDown(layer, xy, event)
	}
	drawing.Drawing.OnMouseUp = func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) (timesel *timeslice.TimeSlice) {
		return drawing.OnMouseUp(layer, xy, event)
	}
	drawing.Drawing.OnMouseMove = func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) (fRedraw bool) {
		return drawing.OnMouseMove(layer, xy, event)
	}
	drawing.Drawing.OnWheel = func(layer *drawingLayer, event *htmlevent.WheelEvent) (timesel *timeslice.TimeSlice) {
		return drawing.OnWheel(layer, event)
	}
	return drawing
}

// OnRedrawTimeSelector draws the timeslice selector on top of the navbar.
// Buttons's position are updated to make it easy to catch them during a mouse event.
func (drawing *DrawingTimeSelector) OnRedraw(layer *drawingLayer) {
	if drawing.series == nil || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		log.Printf("OnRedraw %s fails: unable to proceed given data", drawing.Name)
		return
	}
	// get the y center for redrawing the buttons
	ycenter := float64(layer.clipArea.O.Y) + float64(layer.clipArea.Height)/2.0

	// draw the left selector
	xleftrate := drawing.xAxisRange.Progress(drawing.timeSelection.From)
	xposleft := float64(layer.clipArea.O.X) + float64(layer.clipArea.Width)*xleftrate
	layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Alpha(0.4).Hexa())})
	layer.ctx2D.FillRect(float64(layer.clipArea.O.X), float64(layer.clipArea.O.Y), xposleft, float64(layer.clipArea.Height))
	moveButton(layer, &drawing.fromButton, xposleft, ycenter)

	// draw the right selector
	xrightrate := drawing.xAxisRange.Progress(drawing.timeSelection.To)
	xposright := float64(layer.clipArea.O.X) + float64(layer.clipArea.Width)*xrightrate
	layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Alpha(0.4).Hexa())})
	layer.ctx2D.FillRect(xposright, float64(layer.clipArea.O.Y), float64(layer.clipArea.Width)-xposright, float64(layer.clipArea.Height))
	moveButton(layer, &drawing.toButton, xposright, ycenter)
}

// moveButton utility
func moveButton(layer *drawingLayer, button *Rect, xcenter float64, ycenter float64) {
	layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(bootstrapcolor.Gray.Hexa())})
	layer.ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(bootstrapcolor.Gray.Hexa())})
	layer.ctx2D.SetLineWidth(1)
	x0 := xcenter - float64(button.Width)/2
	y0 := ycenter - float64(button.Height)/2
	layer.ctx2D.FillRect(x0, y0, float64(button.Width), float64(button.Height))
	layer.ctx2D.StrokeRect(x0, y0, float64(button.Width), float64(button.Height))
	button.O.X = int(x0)
	button.O.Y = int(y0)
}

// OnMouseDown starts dragging buttons
func (drawing *DrawingTimeSelector) OnMouseDown(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) {
	//fmt.Printf("%q mousedown xy:%v, frombutton:%v, tobutton:%v\n", drawing.Name, xy, drawing.fromButton, drawing.toButton) //DEBUG:

	// if already dragging and reenter into the canvas
	if drawing.dragFrom || drawing.dragTo {
		drawing.dragIn = true
	} else {
		drawing.dragIn = false
	}

	// do someting only if we're on a button
	if xy.IsIn(drawing.fromButton) {
		drawing.dragFrom = true
		drawing.xStartDrag = drawing.timeSelection
	} else if xy.IsIn(drawing.toButton) {
		drawing.dragTo = true
		drawing.xStartDrag = drawing.timeSelection
	}
}

// OnMouseUp returns the updated timeslice selected after moving buttons
func (drawing *DrawingTimeSelector) OnMouseUp(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) (timesel *timeslice.TimeSlice) {
	//fmt.Printf("%q mouseup xy:%v\n", drawing.Name, xy) //DEBUG:

	if (drawing.dragFrom || drawing.dragTo) && (drawing.dragIn || drawing.xStartDrag.Equal(drawing.timeSelection) == 0) {
		timesel = new(timeslice.TimeSlice)
		*timesel = drawing.timeSelection
	}
	drawing.dragFrom = false
	drawing.dragTo = false
	return timesel
}

// OnMouseMove change the cursor when hovering a button, or change the timeslice selection if dragging the buttons
func (drawing *DrawingTimeSelector) OnMouseMove(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) (fRedraw bool) {
	//fmt.Printf("%q mousemove xy:%v\n", pdrawing.Name xy) //DEBUG:
	if drawing.xAxisRange == nil {
		log.Printf("OnMouseMove %q fails: missing data", drawing.Name)
		return
	}

	// change cursor if we start overing a button
	if (xy.IsIn(drawing.fromButton) || xy.IsIn(drawing.toButton)) && !drawing.resizeCursor {
		layer.ctx2D.Canvas().AttributeStyleMap().Set("cursor", &typedom.Union{Value: js.ValueOf(`col-resize`)})
		drawing.resizeCursor = true
	}

	// change cursor if we leave a button
	if (!xy.IsIn(drawing.fromButton) && !xy.IsIn(drawing.toButton)) && drawing.resizeCursor {
		layer.ctx2D.Canvas().AttributeStyleMap().Set("cursor", &typedom.Union{Value: js.ValueOf(`auto`)})
		drawing.resizeCursor = false
	}

	// change the selector
	if drawing.dragFrom {
		rate := layer.clipArea.XRate(xy.X)
		fromtime := drawing.xAxisRange.WhatTime(rate)
		// HACK: cap flag not working!
		if fromtime.Before(drawing.timeSelection.To) {
			drawing.timeSelection.MoveFrom(fromtime, true)
			fRedraw = true
		}

	} else if drawing.dragTo {
		rate := layer.clipArea.XRate(xy.X)
		totime := drawing.xAxisRange.WhatTime(rate)
		// HACK: cap flag not working!
		if totime.After(drawing.timeSelection.From) {
			drawing.timeSelection.MoveTo(totime, true)
			fRedraw = true
		}
	}
	return fRedraw
}

// OnWheel manage zoom and shifting the time selection
func (drawing *DrawingTimeSelector) OnWheel(layer *drawingLayer, event *htmlevent.WheelEvent) (timesel *timeslice.TimeSlice) {
	if drawing.series == nil || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		return
	}

	// define a good timestep
	currentSel := drawing.timeSelection
	timeStep := time.Duration(float64(*drawing.timeSelection.Duration()) * 0.2)
	//fmt.Printf("wheel shiftK=%v, dy=%f, timeStep=%v\n", event.ShiftKey(), dy, timeStep) // DEBUG:

	// get the wheel move
	dy := event.DeltaY()

	// move the selection
	if event.ShiftKey() {
		if dy < 0 {
			// move to the future
			d := drawing.timeSelection.Duration()
			drawing.timeSelection.To = drawing.timeSelection.To.Add(timeStep)
			drawing.timeSelection.From = drawing.timeSelection.From.Add(timeStep)
			if drawing.timeSelection.To.After(drawing.xAxisRange.To) {
				drawing.timeSelection.To = drawing.xAxisRange.To
				drawing.timeSelection.From = drawing.timeSelection.To.Add(-time.Duration(*d))
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
	} else { // zoom, moving the from date only
		if dy > 0 {
			// reduce the timeslice, but no more than the drawing.timeSelection.To duration
			at := drawing.series.GetAt(drawing.timeSelection.To)
			mind := at.Duration
			drawing.timeSelection.From = drawing.timeSelection.From.Add(timeStep)
			//fmt.Printf("zoom- mind=%s, at:%s, timeStep=%v\n", mind, at, timeStep) // DEBUG:
			if drawing.timeSelection.From.Add(mind).After(drawing.timeSelection.To) {
				drawing.timeSelection.From = at.TimeStamp
			}
		} else if dy < 0 {
			// enlarge the timeslice
			//fmt.Printf("zoom+ timeStep=%v\n", timeStep) // DEBUG:
			drawing.timeSelection.From = drawing.timeSelection.From.Add(-timeStep)
			if drawing.timeSelection.From.Before(drawing.xAxisRange.From) {
				drawing.timeSelection.From = drawing.xAxisRange.From
			}
		}
	}

	// redraw only if the time Selection has changed
	if currentSel.Equal(drawing.timeSelection) == 0 {
		timesel = new(timeslice.TimeSlice)
		*timesel = drawing.timeSelection
		layer.Clear()
		drawing.OnRedraw(layer)
	}
	return timesel
}
