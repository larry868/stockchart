package stockchart

import (
	"fmt"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb/v2/bootstrapcolor.go"
	"github.com/sunraylab/timeline/v2"
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
		drawing.OnRedraw()
	}
	drawing.Drawing.NeedRedraw = func() bool {
		fnillast := drawing.lastSelectedData == nil
		fnilnow := drawing.chart.selectedData == nil
		fneed := fnilnow != fnillast // if one is nil and not the other, or vis et versa
		fneed = fneed || (!fnillast && !fnilnow && drawing.lastSelectedData.TimeSlice.Compare(drawing.chart.selectedData.TimeSlice) != timeline.EQUAL)
		// var strlast, strnow string
		// if !fnillast {
		// 	strlast = drawing.lastSelectedData.TimeSlice.String()
		// }
		// if !fnilnow {
		// 	strnow = drawing.Layer.selectedData.TimeSlice.String()
		// }
		// fmt.Printf("series needs redraw:%v, last:%s now:%s\n", fneed, strlast, strnow)
		return fneed
	}
	return drawing
}

func (drawing *DrawingSeries) OnRedraw() {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		Debug(DBG_REDRAW, fmt.Sprintf("%q OnRedraw fails: unable to proceed given data", drawing.Name))
		return
	}

	// reduce the cliping area
	drawArea := drawing.ClipArea.Shrink(0, 5)

	// get xfactor & yfactor according to time selection
	xfactor := float64(drawArea.Width) / float64(drawing.xAxisRange.Duration().Duration)
	yfactor := float64(drawArea.Height) / drawing.series.DataRange(drawing.xAxisRange, 10).Delta()
	lowboundary := drawing.series.DataRange(drawing.xAxisRange, 10).Low()

	Debug(DBG_REDRAW, fmt.Sprintf("%q drawarea:%s, xAxisRange:%v, xfactor:%f yfactor:%f\n", drawing.Name, drawArea, drawing.xAxisRange.String(), xfactor, yfactor))

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
			xopen := drawArea.O.X + int(xfactor*float64(item.TimeSlice.From.Sub(drawing.xAxisRange.From)))
			yopen := drawArea.O.Y + drawArea.Height - int(yfactor*(item.Open-lowboundary))
			drawing.Ctx2D.MoveTo(float64(xopen), float64(yopen))
			drawing.Ctx2D.BeginPath()
			drawing.Ctx2D.LineTo(float64(xopen), float64(yopen))
			x0 = xopen
		}

		xclose = drawArea.O.X + int(xfactor*float64(item.TimeSlice.To.Sub(drawing.xAxisRange.From)))
		yclose := drawArea.O.Y + drawArea.Height - int(yfactor*(item.Close-lowboundary))
		drawing.Ctx2D.LineTo(float64(xclose), float64(yclose))

		// scan next item
		item = item.Next
	}

	// draw the top line
	drawing.Ctx2D.Stroke()

	// then draw the area
	if drawing.fFillArea {
		drawing.Ctx2D.LineTo(float64(xclose), float64(drawing.ClipArea.End().Y))
		drawing.Ctx2D.LineTo(float64(x0), float64(drawing.ClipArea.End().Y))
		drawing.Ctx2D.ClosePath()
		fillrule := canvas.NonzeroCanvasFillRule
		drawing.Ctx2D.Fill(&fillrule)
	}

	// draw selected data if any
	if drawing.chart.selectedData != nil {
		tsemiddle := drawing.chart.selectedData.Middle()
		if drawing.xAxisRange.WhereIs(tsemiddle)&timeline.TS_IN > 0 {
			drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
			drawing.Ctx2D.SetLineWidth(1)
			xseldata := drawArea.O.X + int(xfactor*float64(tsemiddle.Sub(drawing.xAxisRange.From)))
			drawing.Ctx2D.BeginPath()
			drawing.Ctx2D.MoveTo(float64(xseldata)+0.5, float64(drawing.ClipArea.O.Y))
			drawing.Ctx2D.LineTo(float64(xseldata)+0.5, float64(drawing.ClipArea.O.Y+drawing.ClipArea.Height))
			drawing.Ctx2D.Stroke()
		}
	}

	// memorize last sel data
	drawing.lastSelectedData = drawing.chart.selectedData
}
