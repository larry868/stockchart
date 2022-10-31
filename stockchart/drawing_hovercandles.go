package stockchart

import (
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
		drawing.onMouseMove(xy, event)
	}
	drawing.Drawing.OnClick = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.onClick(xy, event)
	}
	drawing.Drawing.OnMouseLeave = func(xy Point, event *htmlevent.MouseEvent) {
		drawing.hoverData = nil
		drawing.Clear()
	}
	return drawing
}

// draw the line over the candle where the mouse is
func (drawing *DrawingHoverCandles) onMouseMove(xy Point, event *htmlevent.MouseEvent) {

	// get the candle
	trate := drawing.drawArea.XRate(xy.X)
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
	xpos := drawing.drawArea.O.X + int(float64(drawing.drawArea.Width)*xtimerate)
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	drawing.Ctx2D.FillRect(float64(xpos), float64(drawing.drawArea.O.Y), 1, float64(drawing.drawArea.Height))

	// draw the date in the footer
	strdtefmt := timeline.MASK_SHORTEST.GetTimeFormat(middletime, time.Time{})
	if drawing.chart.localZone {
		middletime = middletime.Local()
	} else {
		middletime = middletime.UTC()
	}
	strtime := middletime.Format(strdtefmt)
	drawing.Ctx2D.SetFont(`12px 'Roboto', sans-serif`)
	drawing.DrawTextBox(strtime, Point{X: xpos, Y: drawing.drawArea.O.Y + drawing.drawArea.Height}, AlignCenter|AlignBottom, rgb.White, drawing.MainColor, 5, 1, 1)
}

// select a candle
func (drawing *DrawingHoverCandles) onClick(xy Point, event *htmlevent.MouseEvent) {

	if event.ShiftKey() {
		drawing.chart.selectedData = nil
		return 
	}

	// get the candle
	trate := drawing.drawArea.XRate(xy.X)
	postime := drawing.xAxisRange.WhatTime(trate)
	drawing.chart.selectedData = drawing.series.GetDataAt(postime)
	if postime.IsZero() || drawing.chart.selectedData == nil {
		Debug(DBG_EVENT, "%q OnClick xy:%v ==> no data found at this position", drawing.Name, xy)
	} else {
		Debug(DBG_EVENT, "%q OnClick xy:%v ==> %s", drawing.Name, xy, drawing.chart.selectedData.String())
	}
}
