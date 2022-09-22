package stockchart

import (
	"log"
	"math"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb"
	"github.com/sunraylab/rgb/bootstrapcolor.go"
	"github.com/sunraylab/timeline/timeslice"
)

// Drawing a series of Candles
type DrawingCandles struct {
	Drawing
}

// Drawing factory
func NewDrawingCandles(series *DataList, xrange *timeslice.TimeSlice) *DrawingCandles {
	drawing := new(DrawingCandles)
	drawing.Name = "candles"
	drawing.series = series
	drawing.xAxisRange = xrange
	drawing.Drawing.OnRedraw = func(layer *drawingLayer) {
		drawing.OnRedraw(layer)
	}
	drawing.Drawing.OnChangeTimeSelection = func(layer *drawingLayer, timesel timeslice.TimeSlice) {
		drawing.OnChangeTimeSelection(layer, timesel)
	}
	return drawing
}

// OnRedraw redraws all candles inside the xAxisRange of the OHLC series
// The layer should have been cleared before.
func (drawing DrawingCandles) OnRedraw(layer *drawingLayer) {
	if drawing.series == nil || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		log.Printf("OnRedraw %q fails: unable to proceed given data", drawing.Name)
		return
	}

	var (
		greenCandle   = rgb.Color(0x7dce13ff)
		redCandle     = rgb.Color(0xb20016ff)
		neutralCandle = bootstrapcolor.Gray
	)

	// reduce the drawing area
	drawArea := layer.ClipArea.Shrink(0, 5)
	drawArea.Height -= 15
	//fmt.Printf("clip:%s draw:%s\n", layer.clipArea, drawArea) // DEBUG:

	// get xfactor & yfactor according to time selection
	xfactor := float64(drawArea.Width) / float64(*drawing.xAxisRange.Duration())
	yfactor := float64(drawArea.Height) / drawing.series.DataRange(drawing.xAxisRange, 10).Delta()
	lowboundary := drawing.series.DataRange(drawing.xAxisRange, 10).Low()
	//fmt.Printf("xfactor:%f yfactor:%f\n", xfactor, yfactor) // DEBUG:

	// scan all points
	item := drawing.series.Head
	for item != nil {
		// skip items out of range
		if item.TimeStamp.Add(item.Duration).Before(drawing.xAxisRange.From) {
			// scan next item
			item = item.Next
			continue
		}
		if item.TimeStamp.After(drawing.xAxisRange.To) {
			break
		}

		// choose the color
		candleColor := neutralCandle
		if item.Close > item.Open {
			candleColor = greenCandle
		} else if item.Close < item.Open {
			candleColor = redCandle
		}
		layer.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(candleColor.Hexa())})

		// build the OC candle rect inside the drawing areaa
		rcandle := new(Rect)

		// x axis: time
		rcandle.O.X = drawArea.O.X + int(math.Round(xfactor*float64(item.TimeStamp.Sub(drawing.xAxisRange.From))))
		rcandle.Width = imax(1, int(math.Round(xfactor*float64(item.Duration))))

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
			layer.Ctx2D.FillRect(float64(rcandle.O.X), float64(rcandle.O.Y), float64(rcandle.Width), float64(rcandle.Height))
			//fmt.Printf("candle: %v\n", rcandle) // DEBUG:
		}

		// build LH candle-wick rect
		rwick := new(Rect)
		rwick.Width = imax(xpadding, 1)
		xtimerate := drawing.xAxisRange.Progress(item.TimeStamp.Add(item.Duration / 2))
		rwick.O.X = layer.ClipArea.O.X + int(float64(layer.ClipArea.Width)*xtimerate) - rwick.Width/2

		// need to reverse the candle in canvas coordinates
		rwick.O.Y = drawArea.O.Y + drawArea.Height - int(yfactor*(item.Low-lowboundary))
		rwick.Height = -int(yfactor * (item.High - item.Low))
		rwick.FlipPositive()

		// draw LH only if inside the drawing area,same color
		if rwick = drawArea.And(*rwick); rwick != nil {
			layer.Ctx2D.FillRect(float64(rwick.O.X), float64(rwick.O.Y), float64(rwick.Width), float64(rwick.Height))
		}

		// scan next item
		item = item.Next
	}

}

// OnChangeTimeSelection reset the xAxisRange and redraw all.
// The layer should have been cleared before.
func (drawing *DrawingCandles) OnChangeTimeSelection(layer *drawingLayer, timesel timeslice.TimeSlice) {
	*drawing.xAxisRange = timesel
	drawing.OnRedraw(layer)
}
