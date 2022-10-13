package stockchart

import (
	"fmt"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/gowebapi/webapi/html/htmlevent"
	"github.com/sunraylab/rgb/v2"
	"github.com/sunraylab/timeline/v2"
)

type DrawingHoverCandles struct {
	Drawing
	hoverData *DataStock // the data hovered
}

func NewDrawingHoverCandles(series *DataList) *DrawingHoverCandles {
	drawing := new(DrawingHoverCandles)
	drawing.Name = "hovercandles"
	drawing.series = series
	drawing.MainColor = rgb.Black.Lighten(0.5)
	drawing.Drawing.OnMouseMove = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.OnMouseMove(xy, event)
	}
	drawing.Drawing.OnClick = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.OnClick(xy, event)
	}
	drawing.Drawing.OnMouseLeave = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.hoverData = nil
		drawing.Clear()
	}
	drawing.Drawing.NeedRedraw = func() bool {
		return true
	}
	return drawing
}

// draw the line over the candle where the mouse is
func (drawing *DrawingHoverCandles) OnMouseMove(xy Point, event *htmlevent.MouseEvent) {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		Debug(DBG_EVENT, fmt.Sprintf("%q OnMouseMove fails, missing data", drawing.Name))
		return
	}

	// get the candle
	trate := drawing.ClipArea.XRate(xy.X)
	postime := drawing.xAxisRange.WhatTime(trate)
	hoverData := drawing.series.GetDataAt(postime)
	if postime.IsZero() || hoverData == nil {
		return
	}

	// do not redraw unchanged hovering datapoint
	if hoverData == drawing.hoverData {
		return
	}
	drawing.hoverData = hoverData

	// remove previous line
	drawing.Clear()

	// draw a line at the middle of the selected candle
	middletime := hoverData.TimeSlice.Middle()
	xtimerate := drawing.xAxisRange.Progress(middletime)
	xpos := drawing.ClipArea.O.X + int(float64(drawing.ClipArea.Width)*xtimerate)
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	drawing.Ctx2D.FillRect(float64(xpos), float64(drawing.ClipArea.O.Y), 1, float64(drawing.ClipArea.Height))

	// draw the date in the footer
	strdtefmt := timeline.MASK_SHORTEST.GetTimeFormat(middletime, time.Time{})
	strtime := middletime.Format(strdtefmt)
	drawing.Ctx2D.SetFont(`12px 'Roboto', sans-serif`)
	drawing.DrawTextBox(strtime, Point{X: xpos, Y: drawing.ClipArea.O.Y + drawing.ClipArea.Height}, AlignCenter|AlignBottom, drawing.MainColor, 5, 1, 1)

}

// select a candle
func (drawing *DrawingHoverCandles) OnClick(xy Point, event *htmlevent.MouseEvent) {

	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		drawing.chart.selectedData = nil
		Debug(DBG_EVENT, fmt.Sprintf("%q OnClick fails, missing data", drawing.Name))
		return
	}

	// get the candle
	trate := drawing.ClipArea.XRate(xy.X)
	postime := drawing.xAxisRange.WhatTime(trate)
	drawing.chart.selectedData = drawing.series.GetDataAt(postime)
	if postime.IsZero() || drawing.chart.selectedData == nil {
		Debug(DBG_EVENT, fmt.Sprintf("%q OnClick xy:%v ==> no data found at this position", drawing.Name, xy))
	} else {
		Debug(DBG_EVENT, fmt.Sprintf("%q OnClick xy:%v ==> %s", drawing.Name, xy, drawing.chart.selectedData.String()))
	}
}
