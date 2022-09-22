package stockchart

import (
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

	// setup default text drawing properties
	layer.ctx2D.SetTextAlign(canvas.StartCanvasTextAlign)
	layer.ctx2D.SetTextBaseline(canvas.BottomCanvasTextBaseline)
	layer.ctx2D.SetFont(`10px 'Roboto', sans-serif`)

	// define the grid scale
	const minxstepwidth = 100.0
	maxscans := float64(layer.clipArea.Width) / minxstepwidth
	maskmain := drawing.xAxisRange.GetScanMask(uint(maxscans))
	//fmt.Printf("layer %q, drawing %q, maskmain=%v\n", layer.layerId, drawing.Name, maskmain) // DEBUG:

	// draw the second grid, before the main grid because it can overlay
	if maskmain > timeslice.MASK_SHORTEST {

		// set fillstyle for the grid lines
		layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Lighten(0.5).Hexa())})

		// scan the full xAxisRange every sub-level mask step
		var xtime time.Time
		for drawing.xAxisRange.Scan(&xtime, maskmain-1, true); !xtime.IsZero(); drawing.xAxisRange.Scan(&xtime, maskmain-1, false) {

			// calculate xpos. if the timeslice is a single date then draw a single bar at the middle
			xtimerate := drawing.xAxisRange.Progress(xtime)
			xpos := layer.clipArea.O.X + int(float64(layer.clipArea.Width)*xtimerate)

			// draw the second grid line
			fh := layer.clipArea.Height
			if !drawing.fFullGrid {
				fh = 10
			}
			layer.ctx2D.FillRect(float64(xpos), float64(layer.clipArea.O.Y+layer.clipArea.Height), 1.0, -float64(fh))
		}
	}

	// draw the main grid
	lastlabelend := 0
	var xtime, lastxtime time.Time
	for drawing.xAxisRange.Scan(&xtime, maskmain, true); !xtime.IsZero(); drawing.xAxisRange.Scan(&xtime, maskmain, false) {

		// calculate xpos. if the timeslice is a single date then draw a single bar at the middle
		xtimerate := drawing.xAxisRange.Progress(xtime)
		xpos := layer.clipArea.O.X + int(float64(layer.clipArea.Width)*xtimerate)

		// draw the main grid line
		layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Alpha(50).Hexa())})
		fh := layer.clipArea.Height
		if !drawing.fFullGrid {
			fh = 15
		}
		layer.ctx2D.FillRect(float64(xpos), float64(layer.clipArea.O.Y+layer.clipArea.Height), 1.0, -float64(fh))
		//fmt.Printf("xaxis xpos:%v\n", xpos) // DEBUG:

		// draw time label if not overlapping last label
		if (xpos + 2) > lastlabelend {
			strdtefmt := maskmain.GetTimeFormat(xtime, lastxtime)
			label := xtime.Format(strdtefmt)
			layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(bootstrapcolor.Gray.Darken(0.5).Hexa())})
			layer.ctx2D.FillText(label, float64(xpos+2), float64(layer.clipArea.End().Y)-1, nil)
			lastlabelend = xpos + 2 + int(layer.ctx2D.MeasureText(label).Width())
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
