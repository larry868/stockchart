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
		drawing.OnMouseMove(xy, event)
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

func (drawing *DrawingHoverCandles) OnMouseMove(xy Point, event *htmlevent.MouseEvent) {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		// log.Printf("OnMouseMove %q fails: unable to proceed given data", drawing.Name) DEBUG:
		return
	}
	//fmt.Printf("%q mousemove xy:%v\n", drawing.Name, xy) //DEBUG:

	// get the candle
	trate := drawing.ClipArea.XRate(xy.X)
	postime := drawing.xAxisRange.WhatTime(trate)
	hoverData := drawing.series.GetDataAt(postime)
	if postime.IsZero() || hoverData == nil {
		// fmt.Println("no data at this position") // DEBUG//
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
	xtimerate := drawing.xAxisRange.Progress(hoverData.TimeSlice.Middle())
	xpos := drawing.ClipArea.O.X + int(float64(drawing.ClipArea.Width)*xtimerate)
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	drawing.Ctx2D.FillRect(float64(xpos), float64(drawing.ClipArea.O.Y), 1, float64(drawing.ClipArea.Height))

	// draw the date in the footer
	strdtefmt := timeline.MASK_SHORTEST.GetTimeFormat(postime, time.Time{})
	strtime := postime.Format(strdtefmt)
	drawing.Ctx2D.SetFont(`12px 'Roboto', sans-serif`)
	drawing.DrawTextBox(strtime, Point{X: xpos, Y: drawing.ClipArea.O.Y + drawing.ClipArea.Height}, AlignCenter|AlignBottom, drawing.MainColor, 5, 1, 1)

}
