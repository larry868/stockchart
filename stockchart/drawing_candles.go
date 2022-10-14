package stockchart

import (
	"fmt"
	"math"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb/v2"
	"github.com/sunraylab/timeline/v2"
)

// Drawing a series of Candles
type DrawingCandles struct {
	Drawing
}

// Drawing factory
func NewDrawingCandles(series *DataList) *DrawingCandles {
	drawing := new(DrawingCandles)
	drawing.Name = "candles"
	drawing.series = series
	drawing.MainColor = rgb.Black.Lighten(0.5)
	drawing.Drawing.OnRedraw = func() {
		drawing.OnRedraw()
	}
	drawing.Drawing.NeedRedraw = func() bool {
		return true
	}
	return drawing
}

// OnRedraw redraws all candles inside the xAxisRange of the OHLC series
// The layer should have been cleared before.
func (drawing DrawingCandles) OnRedraw() {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		Debug(DBG_REDRAW, fmt.Sprintf("%q OnRedraw fails: unable to proceed given data", drawing.Name))
		return
	}

	var (
		greenCandle   = rgb.Color(0x7dce13ff)
		redCandle     = rgb.Color(0xb20016ff)
		neutralCandle = rgb.Gray
	)

	// reduce the drawing area
	drawArea := drawing.ClipArea.Shrink(0, 5)
	drawArea.Height -= 15
	
	// get xfactor & yfactor according to time selection
	xfactor := float64(drawArea.Width) / float64(drawing.xAxisRange.Duration().Duration)
	yfactor := float64(drawArea.Height) / drawing.series.DataRange(drawing.xAxisRange, 10).Delta()
	lowboundary := drawing.series.DataRange(drawing.xAxisRange, 10).Low()

	Debug(DBG_REDRAW, fmt.Sprintf("%q drawarea:%s, xAxisRange:%v, xfactor:%f yfactor:%f\n", drawing.Name, drawArea, drawing.xAxisRange.String(), xfactor, yfactor))

	// draw selected data if any
	if drawing.chart.selectedData != nil {
		tsemiddle := drawing.chart.selectedData.Middle()
		if drawing.xAxisRange.WhereIs(tsemiddle)&timeline.TS_IN > 0 {
			drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
			drawing.Ctx2D.SetLineWidth(1)
			xseldata := drawArea.O.X + int(xfactor*float64(tsemiddle.Sub(drawing.xAxisRange.From)))
			drawing.Ctx2D.BeginPath()
			drawing.Ctx2D.MoveTo(float64(xseldata)+0.5, float64(drawing.ClipArea.O.Y))
			drawing.Ctx2D.LineTo(float64(xseldata)+0.5, float64(drawing.ClipArea.O.Y+drawing.ClipArea.Height))
			drawing.Ctx2D.Stroke()
		}
	}

	// scan all points
	var last *DataStock
	var rcandle *Rect
	item := drawing.series.Tail
	for item != nil {
		// skip items out of range
		d := float64(item.Duration().Duration)
		if item.To.Before(drawing.xAxisRange.From) || item.IsInfinite() || d == 0.0 {
			// scan next item
			item = item.Next
			continue
		}
		if item.TimeSlice.From.After(drawing.xAxisRange.To) {
			break
		}

		// choose the color
		candleColor := neutralCandle
		if item.Close > item.Open {
			candleColor = greenCandle
		} else if item.Close < item.Open {
			candleColor = redCandle
		}
		drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(candleColor.Hexa())})

		// build the OC candle rect inside the drawing areaa
		rcandle = new(Rect)

		// x axis: time
		rcandle.O.X = drawArea.O.X + int(math.Round(xfactor*float64(item.TimeSlice.From.Sub(drawing.xAxisRange.From))))
		rcandle.Width = imax(1, int(math.Round(xfactor*d)))

		// add padding between candles
		xpadding := int(float64(rcandle.Width) * 0.1)
		rcandle.O.X += xpadding
		rcandle.Width -= 2 * xpadding

		// draw the candle only for significant width, otherwise draw the wick only
		if rcandle.Width > 2 {

			// y axis: value
			// need to reverse the candle in canvas coordinates
			rcandle.O.Y = drawArea.O.Y + drawArea.Height - int(yfactor*(item.Open-lowboundary))
			rcandle.Height = -int(yfactor * (item.Close - item.Open))
			rcandle.FlipPositive()

			// skip candles outside the drawing area
			if rcandle = drawArea.And(*rcandle); rcandle == nil {
				item = item.Next
				continue
			}

			// draw OC
			drawing.Ctx2D.FillRect(float64(rcandle.O.X), float64(rcandle.O.Y), float64(rcandle.Width), float64(rcandle.Height))
		}

		// build LH candle-wick rect
		rwick := new(Rect)
		rwick.Width = imax(xpadding, 1)
		xtimerate := drawing.xAxisRange.Progress(item.TimeSlice.Middle())
		rwick.O.X = drawing.ClipArea.O.X + int(float64(drawing.ClipArea.Width)*xtimerate) - rwick.Width/2

		// need to reverse the candle in canvas coordinates
		rwick.O.Y = drawArea.O.Y + drawArea.Height - int(yfactor*(item.Low-lowboundary))
		rwick.Height = -int(yfactor * (item.High - item.Low))
		rwick.FlipPositive()

		// draw LH only if inside the drawing area,same color
		if rwick = drawArea.And(*rwick); rwick != nil {
			drawing.Ctx2D.FillRect(float64(rwick.O.X), float64(rwick.O.Y), float64(rwick.Width), float64(rwick.Height))
		}

		// scan next item
		last = item
		item = item.Next
	}

	// draw an ending line at the end of the series
	if last != nil && rcandle != nil {

		drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Lighten(0.2).Opacify(0.5).Hexa())})
		drawing.Ctx2D.FillRect(float64(rcandle.End().X), float64(drawing.ClipArea.O.Y), 1.0, float64(drawing.ClipArea.Height))

		// draw the ending date
		strdtefmt := timeline.MASK_SHORTEST.GetTimeFormat(last.To, time.Time{})
		strtime := last.To.Format(strdtefmt)
		drawing.Ctx2D.SetFont(`10px 'Roboto', sans-serif`)
		drawing.DrawTextBox(strtime, Point{X: rcandle.End().X + 1, Y: drawing.ClipArea.O.Y + drawing.ClipArea.Height}, AlignStart|AlignBottom, drawing.MainColor, 0, 0, 2)

	}

	// draw the label of the series
	drawing.Ctx2D.SetFont(`14px 'Roboto', sans-serif`)
	drawing.DrawTextBox(drawing.series.Name, Point{X: 0, Y: 0}, AlignStart|AlignTop, drawing.MainColor, 3, 0, 2)

}
