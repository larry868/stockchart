package stockchart

import (
	"fmt"
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

	buttonFrom     Rect // coordinates of the button within the drawarea
	buttonTo       Rect // coordinates of the button within the drawarea
	isCursorResize bool // is the cursor the resize one ?
	isCursorGrab   bool // is the cursor the hand one ?

	dragFrom, dragTo, dragIn, dragShift bool

	dragtimeSelection timeline.TimeSlice // locally updated. The chart.timeSelection is only updated when the drag end
	dragShiftlasttime time.Time

	MinZoomTime time.Duration // 5 minutes by default, can be changed
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
	drawing.MinZoomTime = time.Minute * 5

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
	drawing.Drawing.NeedRedraw = func() bool {
		return drawing.dragtimeSelection != drawing.chart.selectedTimeSlice
	}
	// NeddRedraw always false, the time selector is redrawn only by user interecting on it
	return drawing
}

// OnRedrawTimeSelector draws the timeslice selector on top of the navbar.
// Buttons's position are updated to make it easy to catch them during a mouse event.
func (drawing *DrawingTimeSelector) OnRedraw() {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		Debug(DBG_REDRAW, fmt.Sprintf("%q OnRedraw fails: unable to proceed given data", drawing.Name))
		return
	}

	Debug(DBG_REDRAW, fmt.Sprintf("%q OnRedraw drawarea:%s, xAxisRange:%v\n", drawing.Name, drawing.drawArea, drawing.xAxisRange.String()))

	// take into account the selectedTimeSlice at chatr level
	drawing.dragtimeSelection = drawing.chart.selectedTimeSlice

	drawing.redraw()

}

// local redraw, based on the current drawing selection
func (drawing *DrawingTimeSelector) redraw() {

	// get the y center for redrawing the buttons
	ycenter := float64(drawing.drawArea.O.Y) + float64(drawing.drawArea.Height)/2.0

	// draw the left selector
	xleftrate := drawing.xAxisRange.Progress(drawing.dragtimeSelection.From)
	xposleft := float64(drawing.drawArea.O.X) + float64(drawing.drawArea.Width)*xleftrate
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Opacify(0.4).Hexa())})
	drawing.Ctx2D.FillRect(float64(drawing.drawArea.O.X), float64(drawing.drawArea.O.Y), xposleft, float64(drawing.drawArea.Height))
	moveButton(drawing, &drawing.buttonFrom, xposleft, ycenter)

	// draw the right selector
	xrightrate := drawing.xAxisRange.Progress(drawing.dragtimeSelection.To)
	xposright := float64(drawing.drawArea.O.X) + float64(drawing.drawArea.Width)*xrightrate
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Opacify(0.4).Hexa())})
	drawing.Ctx2D.FillRect(xposright, float64(drawing.drawArea.O.Y), float64(drawing.drawArea.Width)-xposright, float64(drawing.drawArea.Height))
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

// OnMouseDown starts dragging
func (drawing *DrawingTimeSelector) OnMouseDown(xy Point, event *htmlevent.MouseEvent) {

	Debug(DBG_EVENT, fmt.Sprintf("%q OnMouseDown xy:%v, frombutton:%v, tobutton:%v", drawing.Name, xy, drawing.buttonFrom, drawing.buttonTo))

	// if already dragging and reenter into the canvas
	if drawing.dragFrom || drawing.dragTo || drawing.dragShift {
		drawing.dragIn = true
	} else {
		drawing.dragIn = false
	}

	// Start draggin mode
	if xy.IsIn(drawing.buttonFrom) {
		drawing.dragFrom = true
	} else if xy.IsIn(drawing.buttonTo) {
		drawing.dragTo = true
	} else if (xy.X > drawing.buttonFrom.End().X) && (xy.X < drawing.buttonTo.O.X) {
		drawing.dragShift = true

		if !drawing.isCursorGrab {
			drawing.Ctx2D.Canvas().AttributeStyleMap().Set("cursor", &typedom.Union{Value: js.ValueOf(`grab`)})
			drawing.isCursorGrab = true
			xrate := drawing.drawArea.XRate(xy.X)
			postime := drawing.xAxisRange.WhatTime(xrate)
			postime = drawing.xAxisRange.Bound(postime)
			drawing.dragShiftlasttime = postime
		}
	}
}

// OnMouseUp Stops Dragging and update the chart.timeSelection.
//
// If the timeselection changes then the event dispatcher will call OnChangeTimeSelection on all drawings of all layers.
func (drawing *DrawingTimeSelector) OnMouseUp(xy Point, event *htmlevent.MouseEvent) {

	// update the chart time selection
	drawing.chart.selectedTimeSlice = drawing.dragtimeSelection

	Debug(DBG_EVENT, fmt.Sprintf("%q OnMouseUp xy:%v, timeselection:%v", drawing.Name, xy, drawing.dragtimeSelection))

	// reset drag flag
	drawing.dragFrom = false
	drawing.dragTo = false
	drawing.dragShift = false

	// reset the cursor if needed
	if drawing.isCursorGrab {
		drawing.Ctx2D.Canvas().AttributeStyleMap().Set("cursor", &typedom.Union{Value: js.ValueOf(`auto`)})
		drawing.isCursorGrab = false
		drawing.dragShiftlasttime = time.Time{}
	}
}

// OnMouseMove change the cursor when hovering a button, or change the local timeslice selection if dragging the buttons.
//
// If the local timeselection changes then redraw this drawing only
// the event dispatcher won't call DoChange as the chart selection has not changed
func (drawing *DrawingTimeSelector) OnMouseMove(xy Point, event *htmlevent.MouseEvent) {
	if drawing.xAxisRange == nil {
		Debug(DBG_EVENT, fmt.Sprintf("%q OnMouseMove fails, missing data", drawing.Name))
		return
	}

	//	Debug(DBG_EVENT, fmt.Sprintf("%q OnMouseMove xy:%v", drawing.Name, xy))

	// change cursor if we start overing a button
	if (xy.IsIn(drawing.buttonFrom) || xy.IsIn(drawing.buttonTo)) && !drawing.isCursorResize {
		drawing.Ctx2D.Canvas().AttributeStyleMap().Set("cursor", &typedom.Union{Value: js.ValueOf(`col-resize`)})
		drawing.isCursorResize = true
	}

	// change cursor if we leave a button
	if (!xy.IsIn(drawing.buttonFrom) && !xy.IsIn(drawing.buttonTo)) && drawing.isCursorResize {
		drawing.Ctx2D.Canvas().AttributeStyleMap().Set("cursor", &typedom.Union{Value: js.ValueOf(`auto`)})
		drawing.isCursorResize = false
	}

	if drawing.dragFrom || drawing.dragTo || drawing.dragShift {
		fRedraw := false
		// change the boundary of the selector
		// get and bound the position of the cursor within xAxisRange
		xrate := drawing.drawArea.XRate(xy.X)
		postime := drawing.xAxisRange.WhatTime(xrate)
		postime = drawing.xAxisRange.Bound(postime)

		if drawing.dragFrom {
			// ensure the position is not over the right boundary minus the MinZoomTime
			if postime.Before(drawing.dragtimeSelection.To.Add(-drawing.MinZoomTime)) {
				drawing.dragtimeSelection.MoveFromAt(postime)
				fRedraw = true
			}
		} else if drawing.dragTo {
			// ensure the position is not over the left boundary plus the MinZoomTime
			if postime.After(drawing.dragtimeSelection.From.Add(drawing.MinZoomTime)) {
				drawing.dragtimeSelection.MoveToAt(postime)
				fRedraw = true
			}
		} else if drawing.dragShift {
			// shift the selection according to the position of the grab selector
			if !postime.Equal(drawing.dragShiftlasttime) {
				shiftdur := postime.Sub(drawing.dragShiftlasttime)
				drawing.dragShiftlasttime = postime
				drawing.dragtimeSelection.ShiftIn(shiftdur, *drawing.xAxisRange)
				fRedraw = true
			}
		}

		if fRedraw {
			drawing.Clear()
			drawing.redraw()

		}
	}
}

// OnWheel manage zoom and shifting the time selection
func (drawing *DrawingTimeSelector) OnWheel(event *htmlevent.WheelEvent) {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		Debug(DBG_EVENT, fmt.Sprintf("%q OnWheel fails, missing data", drawing.Name))
		return
	}

	// get the wheel move
	dir := time.Duration(0)
	dy := event.DeltaY()
	if dy > 0 {
		dir = -1
	} else if dy < 0 {
		dir = 1
	}
	if dir == 0 {
		return
	}

	// define a good timestep: 20% of the current duration
	timeStep := drawing.dragtimeSelection.Duration().Adjust(0.2).Duration
	Debug(DBG_EVENT, fmt.Sprintf("%q OnWheel, shiftKey:%v, dy:%f, timeStep:%v", drawing.Name, event.ShiftKey(), dy, timeStep))

	if event.ShiftKey() {
		// shift mode, shift to the future or to the past according to dir
		drawing.dragtimeSelection.ShiftIn(timeStep*dir, *drawing.xAxisRange)
	} else {
		// zoom mode
		oldfrom := drawing.dragtimeSelection.From
		drawing.dragtimeSelection.ExtendFrom(timeStep * dir)
		drawing.dragtimeSelection.From = drawing.xAxisRange.Bound(drawing.dragtimeSelection.From)
		if drawing.dragtimeSelection.Duration().Duration < drawing.MinZoomTime {
			drawing.dragtimeSelection.From = drawing.dragtimeSelection.To.Add(-drawing.MinZoomTime)
		} else if drawing.dragtimeSelection.From.Equal(oldfrom) {
			// if from is at the limit then try to move to
			drawing.dragtimeSelection.ExtendTo(timeStep * -dir)
			drawing.dragtimeSelection.To = drawing.xAxisRange.Bound(drawing.dragtimeSelection.To)
			if drawing.dragtimeSelection.Duration().Duration < drawing.MinZoomTime {
				drawing.dragtimeSelection.To = drawing.dragtimeSelection.To.Add(drawing.MinZoomTime)
			}
		}

	}

	Debug(DBG_EVENT, fmt.Sprintf("%q OnWheel, newsel=%v", drawing.Name, drawing.dragtimeSelection))

	// update the chart selected timeslice
	drawing.chart.selectedTimeSlice = drawing.dragtimeSelection

	// the event dispatrcher will not propagate the change here because we need to keep the dragtimeSelection
	// so force a clear and a local redraw
	drawing.Clear()
	drawing.redraw()
}
