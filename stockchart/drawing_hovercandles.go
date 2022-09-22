package stockchart

import (
	"fmt"
	"log"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/gowebapi/webapi/html/htmlevent"
	"github.com/sunraylab/rgb/bootstrapcolor.go"
	"github.com/sunraylab/timeline/timeslice"
)

type DrawingHoverCandles struct {
	Drawing
	hoverData *DataPoint // the data hovered
}

func NewDrawingHoverCandles(series *DataList, xrange *timeslice.TimeSlice) *DrawingHoverCandles {
	drawing := new(DrawingHoverCandles)
	drawing.Name = "hovercandles"
	drawing.series = series
	drawing.xAxisRange = xrange
	drawing.MainColor = *bootstrapcolor.Black.Lighten(0.5)
	drawing.Drawing.OnMouseMove = func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) {
		drawing.OnMouseMove(layer, xy, event)
	}
	drawing.Drawing.OnMouseLeave = func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) {
		drawing.hoverData = nil
		layer.Clear()
	}
	return drawing
}

func (drawing *DrawingHoverCandles) OnMouseMove(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) {
	if drawing.series == nil || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		log.Printf("OnMouseMove %q fails: unable to proceed given data", drawing.Name)
		return
	}
	//fmt.Printf("%q mousemove xy:%v\n", drawing.Name, xy) //DEBUG:

	// get the candle
	trate := layer.ClipArea.XRate(xy.X)
	postime := drawing.xAxisRange.WhatTime(trate)
	hoverData := drawing.series.GetAt(postime)
	if postime.IsZero() || hoverData == nil {
		fmt.Println("no data at this position")
		return
	}

	// do not redraw unchanged hovering datapoint
	if hoverData == drawing.hoverData {
		return
	}
	drawing.hoverData = hoverData

	// remove previous line
	layer.Clear()

	// draw a line at the middle of the selected candle
	xtimerate := drawing.xAxisRange.Progress(hoverData.TimeStamp.Add(hoverData.Duration / 2))
	xpos := layer.ClipArea.O.X + int(float64(layer.ClipArea.Width)*xtimerate)
	layer.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	layer.Ctx2D.FillRect(float64(xpos), float64(layer.ClipArea.O.Y), 1, float64(layer.ClipArea.Height))

	// draw the date in the footer
	strdtefmt := timeslice.MASK_SHORTEST.GetTimeFormat(postime, time.Time{})
	strtime := postime.Format(strdtefmt)

	layer.Ctx2D.SetTextAlign(canvas.CenterCanvasTextAlign)
	layer.Ctx2D.SetTextBaseline(canvas.BottomCanvasTextBaseline)
	layer.Ctx2D.SetFont(`12px 'Roboto', sans-serif`)

	// draw the box
	tm := layer.Ctx2D.MeasureText(strtime)
	txtbox := Rect{
		O:      Point{X: xpos - int(tm.Width())/2},
		Width:  int(tm.Width()) + 6,                                                 // add padding
		Height: int(tm.ActualBoundingBoxAscent()+tm.ActualBoundingBoxDescent()) + 5} // add padding
	txtbox.Box(layer.ClipArea) // ensure the txtbox stays inside the clip area
	txtbox.O.Y = layer.ClipArea.End().Y - txtbox.Height - 5
	
	layer.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(bootstrapcolor.White.Hexa())})
	layer.Ctx2D.FillRect(float64(txtbox.O.X), float64(txtbox.O.Y), float64(txtbox.Width), float64(txtbox.Height))
	
	layer.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	layer.Ctx2D.SetLineWidth(1)
	layer.Ctx2D.StrokeRect(float64(txtbox.O.X)+0.5, float64(txtbox.O.Y)+0.5, float64(txtbox.Width-1), float64(txtbox.Height))

	// draw the text
	layer.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	layer.Ctx2D.FillText(strtime, float64(txtbox.Middle().X), float64(txtbox.End().Y), nil)

}
