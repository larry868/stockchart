package stockchart

import (
	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	bootstrapcolor "github.com/larry868/rgb/bootstrapcolor.go"
)

type DrawingSeries struct {
	Drawing
	fFillArea bool // fullfil the area or draw only the line

	lastSelectedData *DataStock
}

func NewDrawingSeries(series *DataList, fFillArea bool) *DrawingSeries {
	drawing := new(DrawingSeries)
	drawing.Name = "series"
	drawing.series = series
	drawing.fFillArea = fFillArea
	drawing.MainColor = bootstrapcolor.Blue.Lighten(0.5)

	drawing.Drawing.OnRedraw = func() {
		// memorize last sel data
		drawing.lastSelectedData = drawing.chart.selectedData
		drawing.onRedraw()
	}

	drawing.Drawing.NeedRedraw = func() bool {
		return drawing.lastSelectedData != drawing.chart.selectedData
	}
	return drawing
}

func (drawing *DrawingSeries) onRedraw() {

	// get xfactor & yfactor according to time selection
	xfactor := float64(drawing.drawArea.Width) / float64(drawing.xAxisRange.Duration().Duration)

	yrange := drawing.series.DataRange(drawing.xAxisRange, 10)
	yfactor := float64(drawing.drawArea.Height) / yrange.Delta()

	// Debug(DBG_REDRAW, "%q drawarea:%s, xAxisRange:%v, xfactor:%f yfactor:%f", drawing.Name, drawing.drawArea, drawing.xAxisRange.String(), xfactor, yfactor)

	// setup drawing tools
	drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	if drawing.fFillArea {
		drawing.Ctx2D.SetLineWidth(3)
	} else {
		drawing.Ctx2D.SetLineWidth(2)
	}
	drawing.Ctx2D.SetLineJoin(canvas.RoundCanvasLineJoin)
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Lighten(0.8).Hexa())})

	// scan all points
	var x0, xclose int
	first := true
	item := drawing.series.Tail
	for item != nil {
		// skip items out of range
		if item.TimeSlice.From.Before(drawing.xAxisRange.From) {
			// scan next item
			item = item.Next
			continue
		}
		if item.TimeSlice.To.After(drawing.xAxisRange.To) {
			break
		}

		// draw the path
		if first {
			first = false
			xopen := drawing.drawArea.O.X + int(xfactor*float64(item.TimeSlice.From.Sub(drawing.xAxisRange.From)))
			yopen := drawing.drawArea.O.Y + drawing.drawArea.Height - int(yfactor*(item.Open-yrange.Low()))
			drawing.Ctx2D.MoveTo(float64(xopen), float64(yopen))
			drawing.Ctx2D.BeginPath()
			drawing.Ctx2D.LineTo(float64(xopen), float64(yopen))
			x0 = xopen
		}

		xclose = drawing.drawArea.O.X + int(xfactor*float64(item.TimeSlice.To.Sub(drawing.xAxisRange.From)))
		yclose := drawing.drawArea.O.Y + drawing.drawArea.Height - int(yfactor*(item.Close-yrange.Low()))
		drawing.Ctx2D.LineTo(float64(xclose), float64(yclose))

		// scan next item
		item = item.Next
	}

	// draw the top line
	drawing.Ctx2D.Stroke()

	// then draw the area
	if drawing.fFillArea {
		drawing.Ctx2D.LineTo(float64(xclose), float64(drawing.drawArea.End().Y))
		drawing.Ctx2D.LineTo(float64(x0), float64(drawing.drawArea.End().Y))
		drawing.Ctx2D.ClosePath()
		fillrule := canvas.NonzeroCanvasFillRule
		drawing.Ctx2D.Fill(&fillrule)
	}

	// draw selected data if any
	if drawing.chart.selectedData != nil {
		tsemiddle := drawing.chart.selectedData.Middle()
		drawing.DrawVLine(tsemiddle, drawing.MainColor, true)
	}
}
