package stockchart

import (
	"math"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/larry868/rgb"
	timeline "github.com/larry868/timeline/v2"
)

// Drawing a series of bars like for volumes
type DrawingBars struct {
	Drawing

	lastSelectedTimeslice timeline.TimeSlice
}

// Drawing factory
func NewDrawingBars(series *DataList) *DrawingBars {
	drawing := new(DrawingBars)
	drawing.Name = "bars"
	drawing.series = series
	drawing.MainColor = rgb.Gray

	drawing.Drawing.OnRedraw = func() {
		drawing.lastSelectedTimeslice = drawing.chart.selectedTimeSlice
		drawing.onRedraw()
	}

	drawing.Drawing.NeedRedraw = func() bool {
		return drawing.lastSelectedTimeslice.Compare(drawing.chart.selectedTimeSlice) != timeline.EQUAL
	}
	return drawing
}

// OnRedraw redraws all bars inside the xAxisRange of the OHLC series
// The layer should have been cleared before.
func (drawing DrawingBars) onRedraw() {

	// get xfactor & yfactor according to time selection
	yrange := drawing.series.VolumeDataRange(drawing.xAxisRange, 0)
	if yrange.Delta() == 0 {
		yrange.ResetBoundaries(0, yrange.High())
	}
	xfactor := float64(drawing.drawArea.Width) / float64(drawing.xAxisRange.Duration().Duration)
	yfactor := float64(drawing.drawArea.Height) / yrange.Delta()

	// Debug(DBG_REDRAW, "%q OnRedraw drawarea:%s, xAxisRange:%v, yrange:%v, xfactor:%f yfactor:%f", drawing.Name, drawing.drawArea, drawing.xAxisRange.String(), yrange.String(), xfactor, yfactor)

	// scan all points
	var rbar *Rect
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
		barcolor := rgb.Gray.Lighten(0.7)
		drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(barcolor.Hexa())})

		// build the BAR rect, inside the drawing areaa
		rbar = new(Rect)

		// x axis: time
		rbar.O.X = drawing.drawArea.O.X + int(math.Round(xfactor*float64(item.TimeSlice.From.Sub(drawing.xAxisRange.From))))
		rbar.Width = imax(1, int(math.Round(xfactor*d)))

		// add padding between bars
		// TODO:
		xpadding := int(float64(rbar.Width) * 0.1)
		rbar.O.X += xpadding
		rbar.Width -= 2 * xpadding

		// y axis: value
		// need to reverse the bar in canvas coordinates
		rbar.O.Y = drawing.drawArea.O.Y + drawing.drawArea.Height - int(yfactor*yrange.Low())
		rbar.Height = -int(yfactor * item.Volume)
		rbar.FlipPositive()

		// skip bars outside the drawing area
		if rbar = drawing.drawArea.And(*rbar); rbar == nil {
			item = item.Next
			continue
		}

		// draw bar
		drawing.Ctx2D.FillRect(float64(rbar.O.X), float64(rbar.O.Y), float64(rbar.Width), float64(rbar.Height))

		// scan next item
		item = item.Next
	}
}
