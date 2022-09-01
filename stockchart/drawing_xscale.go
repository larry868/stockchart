package stockchart

import (
	"fmt"
	"log"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb/bootstrapcolor.go"
	"github.com/sunraylab/timeline/timeslice"
)

type DrawingXScale struct {
	Drawing
	fFullGrid bool // draw grids otherwise only labels
}

func NewDrawingXGrid(series *DataList, xrange *timeslice.TimeSlice, fFullGrid bool) *DrawingXScale {
	drawing := new(DrawingXScale)
	drawing.Name = "xgrid"
	drawing.fFullGrid = fFullGrid
	drawing.series = series
	drawing.xAxisRange = xrange
	drawing.MainColor = *bootstrapcolor.Gray.Lighten(0.3)
	drawing.Drawing.OnRedraw = func(layer *drawingLayer) {
		drawing.OnRedraw(layer)
	}
	drawing.Drawing.OnChangeTimeSelection = func(layer *drawingLayer, timesel timeslice.TimeSlice) {
		drawing.OnChangeTimeSelection(layer, timesel)
	}
	return drawing
}

// OnRedraw DrawingXGrid
func (drawing DrawingXScale) OnRedraw(layer *drawingLayer) {
	if drawing.series == nil || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		log.Printf("OnRedraw %s fails: unable to proceed given data", drawing.Name)
		return
	}

	// do not clear the drawing grid, it's transparent

	// setup default text drawing properties
	layer.ctx2D.SetTextAlign(canvas.StartCanvasTextAlign)
	layer.ctx2D.SetTextBaseline(canvas.BottomCanvasTextBaseline)
	layer.ctx2D.SetFont(`10px 'Roboto', sans-serif`)

	// define the grid scale
	const minxstepwidth = 100.0
	maxscans := float64(layer.clipArea.Width) / minxstepwidth
	maskmain := drawing.xAxisRange.GetScanMask(uint(maxscans))
	fmt.Printf("maskmain=%v\n", maskmain)

	// draw the second grid, before the main grid because it can overlay
	if maskmain > timeslice.MASK_SHORTEST {

		// set fillstyle for the grid lines
		layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Lighten(0.5).Hexa())})

		// scan the full xAxisRange every sub-level mask step
		var xtime time.Time
		for drawing.xAxisRange.Scan(&xtime, maskmain-1, true); !xtime.IsZero(); drawing.xAxisRange.Scan(&xtime, maskmain-1, false) {

			// calculate xpos. if the timeslice is a single date then draw a single bar at the middle
			xtimerate := drawing.xAxisRange.Progress(xtime)
			xpos := float64(layer.clipArea.O.X) + float64(layer.clipArea.Width)*xtimerate

			// draw the second grid line
			fh := float64(layer.clipArea.Height)
			if !drawing.fFullGrid {
				fh = 10
			}
			layer.ctx2D.FillRect(xpos, float64(layer.clipArea.O.Y+layer.clipArea.Height), 1, -fh)
		}
	}

	// draw the main grid
	lastlabelend := 0.0
	var xtime, lastxtime time.Time
	for drawing.xAxisRange.Scan(&xtime, maskmain, true); !xtime.IsZero(); drawing.xAxisRange.Scan(&xtime, maskmain, false) {

		// calculate xpos. if the timeslice is a single date then draw a single bar at the middle
		xtimerate := drawing.xAxisRange.Progress(xtime)
		xpos := float64(layer.clipArea.O.X) + float64(layer.clipArea.Width)*xtimerate

		// draw the main grid line
		layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Alpha(50).Hexa())})
		fh := float64(layer.clipArea.Height)
		if !drawing.fFullGrid {
			fh = 15
		}
		layer.ctx2D.FillRect(xpos, float64(layer.clipArea.O.Y+layer.clipArea.Height), 1, -fh)

		// draw time label if not overlapping last label
		if (xpos + 2) > lastlabelend {
			strdtefmt := maskmain.GetTimeFormat(xtime, lastxtime)
			label := xtime.Format(strdtefmt)
			layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(bootstrapcolor.Gray.Darken(0.5).Hexa())})
			layer.ctx2D.FillText(label, xpos+2, float64(layer.clipArea.End().Y)-1, nil)
			lastlabelend = xpos + 2 + layer.ctx2D.MeasureText(label).Width()
		}

		// scan next
		lastxtime = xtime
	}
}

// OnChangeTimeSelection
func (pdrawing *DrawingXScale) OnChangeTimeSelection(layer *drawingLayer, timesel timeslice.TimeSlice) {
	*pdrawing.xAxisRange = timesel
	pdrawing.OnRedraw(layer)
}
