package stockchart

import (
	"math"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/larry868/rgb"
	bootstrapcolor "github.com/larry868/rgb/bootstrapcolor.go"
	timeline "github.com/larry868/timeline/v2"
)

type DrawStyle int

const (
	DS_Stick DrawStyle = 1
	DS_Bar   DrawStyle = 2
	DS_Area  DrawStyle = 3
	DS_Frame DrawStyle = 4
)

// Drawing a series of Candles
type DrawingCandles struct {
	Drawing

	DrawStyle

	lastSelectedTimeslice timeline.TimeSlice
	lastSelectedData      *DataStock
}

// Drawing factory
// initAlpha must be within 0 (0% opacity = full transparent) and 1 (100% opacity)
func NewDrawingCandles(series *DataList, drawstyle DrawStyle) *DrawingCandles {
	drawing := new(DrawingCandles)
	drawing.Name = "candles"
	drawing.series = series
	drawing.MainColor = rgb.Black.Lighten(0.5)
	drawing.DrawStyle = drawstyle

	// drawing.alphaFactor = alpha
	// drawing.dashstyle = dashstyle

	drawing.Drawing.OnRedraw = func() {
		drawing.lastSelectedTimeslice = drawing.chart.selectedTimeSlice
		drawing.lastSelectedData = drawing.chart.selectedData
		drawing.onRedraw()
	}
	drawing.Drawing.NeedRedraw = func() bool {
		fneedst := drawing.lastSelectedTimeslice.Compare(drawing.chart.selectedTimeSlice) != timeline.EQUAL
		fneedsd := drawing.lastSelectedData != drawing.chart.selectedData
		return fneedst || fneedsd
	}
	return drawing
}

// OnRedraw redraws all candles inside the xAxisRange of the OHLC series
// The layer should have been cleared before.
// update the drawing layer title area
func (drawing *DrawingCandles) onRedraw() {
	// get xfactor & yfactor according to time selection
	yrange := drawing.chart.yAxisRange
	xfactor := float64(drawing.drawArea.Width) / float64(drawing.xAxisRange.Duration().Duration)
	yfactor := float64(drawing.drawArea.Height) / yrange.Delta()

	// Debug(DBG_REDRAW, "%q OnRedraw drawarea:%s, xfactor:%f yfactor:%f style:%v", drawing.Name, drawing.drawArea, xfactor, yfactor, drawing.DrawStyle)
	// Debug(DBG_REDRAW, "%q OnRedraw serie:%v seltime:%s, yrange;%s", drawing.Name, drawing.series.String(), drawing.xAxisRange, yrange)
	//Debug(DBG_REDRAW, "%q OnRedraw xAxisRange:%v,", drawing.Name, drawing.xAxisRange.String())

	// draw a vertical line for the selected data if any
	if drawing.chart.selectedData != nil {
		middletime := drawing.chart.selectedData.Middle()
		drawing.DrawVLine(middletime, drawing.MainColor, true)
	}

	var wcf64, hcf64, xcf64, ycf64 float64
	var wwickf64, hwickf64, xwickf64, ywickf64 float64
	var wpatf64, hpatf64, xpatf64, ypatf64 float64
	const ypat = 10.0

	// scan all points forward !
	drbottomf64 := float64(drawing.drawArea.O.Y + drawing.drawArea.Height)
	item := drawing.series.Tail
	for item != nil {
		// skip items before xAxisRange boundary or without duration
		// skip items after xAxisRange boundary.
		// Do not break because series are not always sorted chronologicaly
		if item.IsInfinite() || item.Duration().Duration == 0 || item.To.Before(drawing.xAxisRange.From) || item.From.After(drawing.xAxisRange.To) {
			item = item.Next
			continue
		}

		// candle width, in px
		wcf64 = fmax(1.0, math.Round(xfactor*float64(item.Duration().Duration)))

		// choose the color
		candleColor := item.CandleColor()
		patternColor := bootstrapcolor.Purple

		// force bar style if width is too small
		style := drawing.DrawStyle
		if wcf64 <= 3.0 {
			wcf64 = 1
			style = DS_Bar
		}
		switch style {
		case DS_Bar:
			// colorfull
			drawing.Ctx2D.SetLineWidth(0)
			drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(candleColor.Hexa())})

			// width & xpos
			wcf64 = 1.0
			xcf64 = drawing.xTime(item.TimeSlice.Middle())

			// height & ypos
			hcf64 = yfactor * (item.High - item.Low)
			ycf64 = drbottomf64 - yfactor*(item.High-yrange.Low())
			hcf64 = fmax(hcf64, 1.0)

			// draw now
			drawing.Ctx2D.FillRect(float64(int(xcf64)), float64(int(ycf64)), float64(int(wcf64)), float64(int(hcf64)))

			// patternif any
			wpatf64 = 1.0
			hpatf64 = 1.0
			xpatf64 = xcf64
			ypatf64 = drbottomf64 - ypat + 2

		case DS_Stick:
			// colorfull
			drawing.Ctx2D.SetLineWidth(0)
			drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(candleColor.Hexa())})

			// width, padding & xpos
			xpaddingf64 := fmax(0.5, wcf64/15)
			wcf64 -= 2 * xpaddingf64
			xcf64 = drawing.xTime(item.From) + xpaddingf64

			// height & ypos
			hcf64 = yfactor * (item.Close - item.Open)
			ycf64 = drbottomf64 - yfactor*(item.Close-yrange.Low())
			if hcf64 < 0 {
				hcf64 = -hcf64
				ycf64 = drbottomf64 - yfactor*(item.Open-yrange.Low())
			}
			hcf64 = fmax(hcf64, 1.0)

			// draw body
			drawing.Ctx2D.FillRect(float64(int(xcf64)), float64(int(ycf64)), float64(int(wcf64)), float64(int(hcf64)))

			// wick
			wwickf64 = fmax(xpaddingf64, 1.0)
			xwickf64 = xcf64 + (wcf64-wwickf64)/2.0

			hwickf64 = fmax(yfactor*(item.High-item.Low), 1.0)
			ywickf64 = drbottomf64 - yfactor*(item.High-yrange.Low())
			drawing.Ctx2D.FillRect(float64(int(xwickf64)), float64(int(ywickf64)), float64(int(wwickf64)), float64(int(hwickf64)))

			// patterns, if any
			wpatf64 = fmin(wcf64, 5.0)
			hpatf64 = wpatf64
			xpatf64 = xcf64 + (wcf64-wpatf64)/2.0
			ypatf64 = drbottomf64 - ypat - 5.0

		case DS_Area:
			// transparent
			drawing.Ctx2D.SetLineWidth(0)
			drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(candleColor.Opacify(0.1).Hexa())})

			// width & xpos
			xcf64 = drawing.xTime(item.From)

			// height & ypos
			hcf64 = yfactor * (item.Close - item.Open)
			ycf64 = drbottomf64 - yfactor*(item.Close-yrange.Low())
			if hcf64 < 0 {
				hcf64 = -hcf64
				ycf64 = drbottomf64 - yfactor*(item.Open-yrange.Low())
			}
			hcf64 = fmax(hcf64, 1.0)

			// draw now
			drawing.Ctx2D.FillRect(float64(int(xcf64))+0.5, float64(int(ycf64))-0.5, float64(int(wcf64)), float64(int(hcf64)))

			// wick
			wwickf64 = wcf64
			xwickf64 = math.Round(drawing.xTime(item.Middle()) - wwickf64/2.0)

			hwickf64 = yfactor * (item.High - item.Low)
			ywickf64 = drbottomf64 - yfactor*(item.High-yrange.Low())
			drawing.Ctx2D.FillRect(float64(int(xwickf64)), float64(int(ywickf64)), float64(int(wwickf64)), float64(int(hwickf64)))

			// pattern if any
			wpatf64 = fmin(wcf64, 5.0)
			hpatf64 = wpatf64
			xpatf64 = xcf64 + (wcf64-wpatf64)/2.0
			ypatf64 = drbottomf64 - ypat - 10

		case DS_Frame:
			// transparent dash
			drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(candleColor.Opacify(0.5).Hexa())})
			drawing.Ctx2D.SetLineDash([]float64{5.0, 5.0})
			drawing.Ctx2D.SetLineWidth(1)

			// width & xpos
			xcf64 = drawing.xTime(item.From)

			// height & ypos
			hcf64 = yfactor * (item.Close - item.Open)
			ycf64 = drbottomf64 - yfactor*(item.Close-yrange.Low())
			if hcf64 < 0 {
				hcf64 = -hcf64
				ycf64 = drbottomf64 - yfactor*(item.Open-yrange.Low())
			}
			hcf64 = fmax(hcf64, 1.0)

			// draw now
			drawing.Ctx2D.StrokeRect(float64(int(xcf64))+0.5, float64(int(ycf64))-0.5, float64(int(wcf64)), float64(int(hcf64)))

			// wick
			wwickf64 = wcf64
			xwickf64 = math.Round(drawing.xTime(item.Middle())) - wwickf64/2.0

			hwickf64 = yfactor * (item.High - item.Low)
			ywickf64 = drbottomf64 - yfactor*(item.High-yrange.Low())
			drawing.Ctx2D.StrokeRect(float64(int(xwickf64))+0.5, float64(int(ywickf64))-0.5, float64(int(wwickf64)), float64(int(hwickf64)))

			// pattern if any
			wpatf64 = fmin(wcf64, 5.0)
			hpatf64 = wpatf64
			xpatf64 = xcf64 + (wcf64-wpatf64)/2.0
			ypatf64 = drbottomf64 - ypat - 15
		}

		// draw has patterns
		if item.HasPatterns > 0 {
			drawing.Ctx2D.SetLineWidth(1)
			drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(patternColor.Hexa())})
			drawing.Ctx2D.StrokeRect(float64(int(xpatf64))-0.5, float64(int(ypatf64))-0.5, float64(int(wpatf64)), float64(int(hpatf64)))
		}

		// scan next item
		item = item.Next
	}

	// draw the label of the series
	var tarea Rect
	if len(drawing.Layer.TitleAreas) > 0 {
		tarea = drawing.Layer.TitleAreas[len(drawing.Layer.TitleAreas)-1]
		tarea.O.Y += drawing.drawArea.O.Y + 15
	}
	drawing.Ctx2D.SetFont(`14px 'Roboto', sans-serif`)
	rtitle := drawing.DrawTextBox(drawing.series.Name, Point{X: 0, Y: tarea.O.Y}, AlignStart|AlignTop, rgb.White.Opacify(0.8), drawing.MainColor, 3, 0, 2)
	drawing.Layer.TitleAreas = append(drawing.Layer.TitleAreas, rtitle)
}
