package stockchart

import (
	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/datarange"
	"github.com/sunraylab/rgb/v2"
)

type DrawingYGrid struct {
	Drawing
	fScale     bool // Draw the scale, otherwise only the lines
	lastyrange datarange.DataRange
}

func NewDrawingYGrid(series *DataList, fscale bool) *DrawingYGrid {
	drawing := new(DrawingYGrid)
	drawing.Name = "ygrid"
	drawing.series = series
	drawing.MainColor = rgb.Gray.Lighten(0.85)
	drawing.fScale = fscale

	drawing.Drawing.OnRedraw = func() {
		drawing.OnRedraw()
	}
	drawing.Drawing.NeedRedraw = func() bool {
		ynewrange := drawing.series.DataRange(&drawing.chart.selectedTimeSlice, 10)
		return !ynewrange.Equal(drawing.lastyrange)
	}
	return drawing
}

// OnRedraw redraw the Y axis
func (drawing *DrawingYGrid) OnRedraw() {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		Debug(DBG_REDRAW, "%q OnRedraw fails: unable to proceed given data", drawing.Name)
		return
	}

//	yrange := drawing.series.DataRange(drawing.xAxisRange, 10)
	yrange := drawing.series.DataRange(&drawing.chart.selectedTimeSlice, 10)

	Debug(DBG_REDRAW, "%q OnRedraw drawarea:%s, xAxisRange:%v, datarange:%v", drawing.Name, drawing.drawArea, drawing.xAxisRange.String(), yrange)

	// setup default text drawing properties
	drawing.Ctx2D.SetTextAlign(canvas.StartCanvasTextAlign)
	drawing.Ctx2D.SetTextBaseline(canvas.MiddleCanvasTextBaseline)
	drawing.Ctx2D.SetFont(`12px 'Roboto', sans-serif`)

	// draw the Y Scale
	for val := yrange.High(); val >= yrange.Low() && yrange.StepSize() > 0; val -= yrange.StepSize() {

		// calculate ypos
		yrate := yrange.Progress(val)
		ypos := float64(drawing.drawArea.End().Y) - yrate*float64(drawing.drawArea.Height)
		ypos = float64(drawing.drawArea.BoundY(int(ypos)))

		// draw the grid line
		drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
		linew := 10.0
		if !drawing.fScale {
			linew = float64(drawing.drawArea.Width)
		}
		drawing.Ctx2D.FillRect(float64(drawing.drawArea.O.X), ypos, linew, 1)

		// draw yscale label
		if drawing.fScale {
			strvalue := datarange.FormatData(val, yrange.StepSize()) // fmt.Sprintf("%.1f", val)
			drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(rgb.Gray.Darken(0.5).Hexa())})
			drawing.Ctx2D.FillText(strvalue, float64(drawing.drawArea.O.Y+7), ypos+1, nil)
		}
	}

	drawing.lastyrange = yrange
}
