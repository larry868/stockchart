package stockchart

import (
	"math"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb/v2"
)

// Drawing a series of Candles
type DrawingCandles struct {
	Drawing
	alphaFactor float32
}

// Drawing factory
// initAlpha must be within 0 (0% opacity = full transparent) and 1 (100% opacity)
func NewDrawingCandles(series *DataList, initAlpha float32) *DrawingCandles {
	drawing := new(DrawingCandles)
	drawing.Name = "candles"
	drawing.series = series
	drawing.MainColor = rgb.Black.Lighten(0.5)
	drawing.alphaFactor = initAlpha

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
		Debug(DBG_REDRAW, "%q OnRedraw fails: unable to proceed given data", drawing.Name)
		return
	}

	var (
		greenCandle   = rgb.Color(0x7dce13ff)
		redCandle     = rgb.Color(0xb20016ff)
		neutralCandle = rgb.Gray
	)

	// get xfactor & yfactor according to time selection
	yrange := drawing.series.DataRange(drawing.xAxisRange, 10)
	xfactor := float64(drawing.drawArea.Width) / float64(drawing.xAxisRange.Duration().Duration)
	yfactor := float64(drawing.drawArea.Height) / yrange.Delta()

	Debug(DBG_REDRAW, "%q drawarea:%s, xAxisRange:%v, xfactor:%f yfactor:%f", drawing.Name, drawing.drawArea, drawing.xAxisRange.String(), xfactor, yfactor)

	// draw selected data if any
	if drawing.chart.selectedData != nil {
		tsemiddle := drawing.chart.selectedData.Middle()
		drawing.DrawVLine(tsemiddle, drawing.MainColor, true)
	}

	// scan all points
	//var last *DataStock
	var rcandle *Rect
	item := drawing.series.Tail
	for item != nil {
		// skip items before xAxisRange boundary or without duration
		d := float64(item.Duration().Duration)
		if item.To.Before(drawing.xAxisRange.From) || item.IsInfinite() || d == 0.0 {
			item = item.Next
			continue
		}

		// do not draw items after xAxisRange boundary
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
		drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(candleColor.Opacify(drawing.alphaFactor).Hexa())})

		// build the OC candle rect, inside the drawing areaa
		rcandle = new(Rect)

		// x axis: time
		rcandle.O.X = drawing.drawArea.O.X + int(math.Round(xfactor*float64(item.TimeSlice.From.Sub(drawing.xAxisRange.From))))
		rcandle.Width = imax(1, int(math.Round(xfactor*d)))

		// add padding between candles
		// xpadding := int(float64(rcandle.Width) * 0.1)
		xpadding := imax(1, int(math.Round(xfactor*float64(drawing.chart.MainSeries.Precision)*0.1)))
		Debug(DBG_REDRAW, "%q padding:%v, precision:%v", drawing.Name, xpadding, drawing.chart.MainSeries.Precision)
		rcandle.O.X += xpadding
		rcandle.Width -= 2 * xpadding

		// draw the candle only for significant width, otherwise draw the wick only
		if rcandle.Width > 2 {

			// y axis: value
			// need to reverse the candle in canvas coordinates
			rcandle.O.Y = drawing.drawArea.O.Y + drawing.drawArea.Height - int(yfactor*(item.Open-yrange.Low()))
			rcandle.Height = -int(yfactor * (item.Close - item.Open))
			rcandle.FlipPositive()

			// skip candles outside the drawing area
			if rcandle = drawing.drawArea.And(*rcandle); rcandle == nil {
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
		rwick.O.X = drawing.drawArea.O.X + int(float64(drawing.drawArea.Width)*xtimerate) - rwick.Width/2

		// need to reverse the candle in canvas coordinates
		rwick.O.Y = drawing.drawArea.O.Y + drawing.drawArea.Height - int(yfactor*(item.Low-yrange.Low()))
		rwick.Height = -int(yfactor * (item.High - item.Low))
		rwick.FlipPositive()

		// draw LH only if inside the drawing area,same color
		if rwick = drawing.drawArea.And(*rwick); rwick != nil {
			drawing.Ctx2D.FillRect(float64(rwick.O.X), float64(rwick.O.Y), float64(rwick.Width), float64(rwick.Height))
		}

		// scan next item
		//last = item
		item = item.Next
	}

	// draw the label of the series
	// TODO: overwritting when multiple drawings on the same chart
	drawing.Ctx2D.SetFont(`14px 'Roboto', sans-serif`)
	drawing.DrawTextBox(drawing.series.Name, Point{X: 0, Y: 0}, AlignStart|AlignTop, rgb.White, drawing.MainColor, 3, 0, 2)
}
