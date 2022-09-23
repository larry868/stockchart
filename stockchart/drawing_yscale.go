package stockchart

import (
	"log"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/datarange"
	"github.com/sunraylab/rgb/v2"
	"github.com/sunraylab/timeline/v2"
)

type DrawingYGrid struct {
	Drawing
	fScale bool // Draw the scale, otherwise only the lines
}

func NewDrawingYGrid(series *DataList, xrange *timeline.TimeSlice, fscale bool) *DrawingYGrid {
	drawing := new(DrawingYGrid)
	drawing.Name = "yscale"
	drawing.series = series
	drawing.xAxisRange = xrange
	drawing.MainColor = rgb.Gray.Lighten(0.85)
	drawing.fScale = fscale
	drawing.Drawing.OnRedraw = func() {
		drawing.OnRedraw()
	}
	drawing.Drawing.OnChangeTimeSelection = func(timesel timeline.TimeSlice) {
		drawing.OnChangeTimeSelection(timesel)
	}
	return drawing
}

// OnRedraw redraw the Y axis
func (drawing *DrawingYGrid) OnRedraw() {
	if drawing.series == nil || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		log.Printf("OnRedraw %q fails: unable to proceed given data", drawing.Name)
		return
	}
	//fmt.Printf("OnRedraw %q starting...\n", drawing) //DEBUG:

	// setup default text drawing properties
	drawing.Ctx2D.SetTextAlign(canvas.StartCanvasTextAlign)
	drawing.Ctx2D.SetTextBaseline(canvas.MiddleCanvasTextBaseline)
	drawing.Ctx2D.SetFont(`12px 'Roboto', sans-serif`)

	// reduce the cliping area
	clipArea := drawing.ClipArea.Shrink(0, 5)
	clipArea.Height -= 15

	// draw the Y Scale
	yrange := drawing.series.DataRange(drawing.xAxisRange, 10)
	//fmt.Printf("DataRange %v\n", yrange) //DEBUG:

	for val := yrange.High(); val >= yrange.Low() && yrange.StepSize() > 0; val -= yrange.StepSize() {

		// calculate ypos
		yrate := yrange.Progress(val)
		ypos := float64(clipArea.End().Y) - yrate*float64(clipArea.Height)
		ypos = float64(clipArea.BoundY(int(ypos)))

		// draw the grid line
		drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
		linew := 10.0
		if !drawing.fScale {
			linew = float64(clipArea.Width)
		}
		drawing.Ctx2D.FillRect(float64(clipArea.O.X), ypos, linew, 1)

		// draw yscale label
		if drawing.fScale {
			strvalue := datarange.FormatData(val, yrange.StepSize()) // fmt.Sprintf("%.1f", val)
			drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(rgb.Gray.Darken(0.5).Hexa())})
			drawing.Ctx2D.FillText(strvalue, float64(clipArea.O.Y+7), ypos+1, nil)
		}
	}
}

// OnChangeTimeSelection
func (pdrawing *DrawingYGrid) OnChangeTimeSelection(timesel timeline.TimeSlice) {
	*pdrawing.xAxisRange = timesel
	pdrawing.OnRedraw()
}
