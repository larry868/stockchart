package stockchart

import (
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb/v2"
	"github.com/sunraylab/timeline/v2"
)

type DrawingXGrid struct {
	Drawing
	fFullGrid bool // draw grids otherwise only labels
}

func NewDrawingXGrid(series *DataList, fFullGrid bool, timeSelDependant bool) *DrawingXGrid {
	drawing := new(DrawingXGrid)
	drawing.Name = "xgrid"
	drawing.fFullGrid = fFullGrid
	drawing.series = series
	drawing.MainColor = rgb.Gray
	drawing.Drawing.OnRedraw = func() {
		drawing.OnRedraw()
	}
	drawing.Drawing.NeedRedraw = func() bool {
		return timeSelDependant
	}
	return drawing
}

// OnRedraw DrawingXGrid
func (drawing DrawingXGrid) OnRedraw() {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		Debug(DBG_REDRAW, "%q OnRedraw fails: unable to proceed given data", drawing.Name)
		return
	}

	// define the grid scale
	const minxstepwidth = 100.0
	maxscans := float64(drawing.drawArea.Width) / minxstepwidth
	maskmain := drawing.xAxisRange.GetScanMask(uint(maxscans))

	Debug(DBG_REDRAW, "%q OnRedraw drawarea:%s, xAxisRange:%v, maskmain=%v\n", drawing.Name, drawing.drawArea, drawing.xAxisRange.String(), maskmain)

	// setup default text drawing properties
	drawing.Ctx2D.SetTextAlign(canvas.StartCanvasTextAlign)
	drawing.Ctx2D.SetTextBaseline(canvas.BottomCanvasTextBaseline)
	drawing.Ctx2D.SetFont(`10px 'Roboto', sans-serif`)

	// set fillstyle for the grid lines
	gMainColor := drawing.MainColor.Opacify(0.4)
	gSecondColor := drawing.MainColor.Lighten(0.5).Opacify(0.4)
	gLabelColor := drawing.MainColor.Darken(0.5)
	if !drawing.fFullGrid {
		gMainColor = gSecondColor
	}

	// draw the second grid, before the main grid because it can overlay
	if maskmain > timeline.MASK_SHORTEST {

		drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(gSecondColor.Hexa())})

		fh := -float64(drawing.drawArea.Height)
		if !drawing.fFullGrid {
			fh = -10.0
		}

		// scan the full xAxisRange every sub-level mask step
		var xtime time.Time
		for drawing.xAxisRange.Scan(&xtime, maskmain-1, true); !xtime.IsZero(); drawing.xAxisRange.Scan(&xtime, maskmain-1, false) {

			// calculate xpos. if the timeslice is a single date then draw a single bar at the middle
			xtimerate := drawing.xAxisRange.Progress(xtime)
			xpos := drawing.drawArea.O.X + int(float64(drawing.drawArea.Width)*xtimerate)

			// draw the grid
			drawing.Ctx2D.FillRect(float64(xpos), float64(drawing.drawArea.O.Y+drawing.drawArea.Height), 1.0, fh)
		}
	}

	// draw the main grid
	lastlabelend := 0
	var xtime, lastxtime time.Time
	for drawing.xAxisRange.Scan(&xtime, maskmain, true); !xtime.IsZero(); drawing.xAxisRange.Scan(&xtime, maskmain, false) {

		// calculate xpos. if the timeslice is a single date then draw a single bar at the middle
		xtimerate := drawing.xAxisRange.Progress(xtime)
		xpos := drawing.drawArea.O.X + int(float64(drawing.drawArea.Width)*xtimerate)

		// draw the main grid line
		drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(gMainColor.Hexa())})
		drawing.Ctx2D.FillRect(float64(xpos), float64(drawing.drawArea.O.Y+drawing.drawArea.Height), 1.0, -float64(drawing.drawArea.Height))

		// draw time label if not overlapping last label
		if (xpos + 2) > lastlabelend {
			strdtefmt := maskmain.GetTimeFormat(xtime, lastxtime)
			label := xtime.Format(strdtefmt)
			drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(gLabelColor.Hexa())})
			drawing.Ctx2D.FillText(label, float64(xpos+2), float64(drawing.drawArea.End().Y)-1, nil)
			lastlabelend = xpos + 2 + int(drawing.Ctx2D.MeasureText(label).Width())
		}

		// scan next
		lastxtime = xtime
	}
}
