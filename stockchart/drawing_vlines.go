package stockchart

import (
	"math"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb/v2"
	"github.com/sunraylab/rgb/v2/bootstrapcolor.go"
	"github.com/sunraylab/timeline/v2"
)

type DrawingVLines struct {
	Drawing

	lastlocalZone bool
}

func NewDrawingVLines(series *DataList, timeSelDependant bool) *DrawingVLines {
	drawing := new(DrawingVLines)
	drawing.Name = "vlines"
	drawing.series = series
	drawing.MainColor = bootstrapcolor.Orange

	drawing.Drawing.OnRedraw = func() {
		drawing.lastlocalZone = drawing.chart.localZone
		drawing.onRedraw()
	}

	drawing.Drawing.NeedRedraw = func() bool {
		return timeSelDependant && (drawing.chart.localZone != drawing.lastlocalZone)
	}
	return drawing
}

// OnRedraw DrawingXGrid
func (drawing DrawingVLines) onRedraw() {
	// get xy factors
	xfactor := float64(drawing.drawArea.Width) / float64(drawing.xAxisRange.Duration().Duration)

	// Debug(DBG_REDRAW, "%q OnRedraw drawarea:%s, xfactor:%f yfactor:%f style:%v", drawing.Name, drawing.drawArea, xfactor, yfactor, drawing.DrawStyle)
	// Debug(DBG_REDRAW, "%q OnRedraw serie:%v seltime:%s, yrange;%s", drawing.Name, drawing.series.String(), drawing.xAxisRange, yrange)
	// Debug(DBG_REDRAW, "%q OnRedraw xAxisRange:%v,", drawing.Name, drawing.xAxisRange.String())

	// drawing style
	drawing.Ctx2D.SetLineWidth(1)
	drawing.Ctx2D.SetLineCap(canvas.ButtCanvasLineCap)
	drawing.Ctx2D.SetLineJoin(canvas.MiterCanvasLineJoin)
	drawing.Ctx2D.SetLineDash([]float64{})
	drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	drawing.Ctx2D.BeginPath()

	var wcf64, xcf64, ycf64 float64
	drawnts := make([]timeline.TimeSlice, 0)

	// scan all points forward !
	drbottomf64 := float64(drawing.drawArea.O.Y + drawing.drawArea.Height)
	item := drawing.series.Tail
	for item != nil {
		// skip items before xAxisRange boundary or without duration.
		// skip items after xAxisRange boundary.
		// Do not break because series are not always sorted chronologicaly
		if item.IsInfinite() || item.Duration().Duration == 0 || item.To.Before(drawing.xAxisRange.From) || item.From.After(drawing.xAxisRange.To) {
			item = item.Next
			continue
		}

		var overlap float64
		for _, k := range drawnts {
			if item.TimeSlice.IsOverlapping(k) {
				overlap++
			}
		}

		// candle size a pos
		wcf64 = fmax(1.0, math.Round(xfactor*float64(item.Duration().Duration)))
		xcf64 = drawing.xTime(item.From)
		ycf64 = drbottomf64 + 5 - 10*overlap

		// draw now
		drawing.Ctx2D.MoveTo(float64(int(xcf64)), float64(int(ycf64))+0.5)
		drawing.Ctx2D.LineTo(float64(int(xcf64+wcf64)), float64(int(ycf64))+0.5)

		// draw the label of the candle
		if item.Label != "" {
			drawing.Ctx2D.SetFont(`8px 'Roboto', sans-serif`)
			drawing.DrawTextBox(item.Label, Point{X: int(xcf64), Y: int(ycf64 - 1)}, AlignStart|AlignBottom, rgb.White.Opacify(0.5), drawing.MainColor, 0, 0, 0)
		}

		// scan next item
		drawnts = append(drawnts, item.TimeSlice)
		item = item.Next
	}
	drawing.Ctx2D.Stroke()

}
