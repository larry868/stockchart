package stockchart

import (
	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb/v2/bootstrapcolor.go"
)

type DrawingSeries struct {
	Drawing
	fFillArea bool // fullfil the area or draw only the line
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
	return drawing
}

func (drawing DrawingSeries) OnRedraw() {
	if drawing.series.IsEmpty() || drawing.xAxisRange == nil || !drawing.xAxisRange.Duration().IsFinite || drawing.xAxisRange.Duration().Seconds() < 0 {
		//log.Printf("serie size: %v, xAxisRange:%v", drawing.series.Size(), drawing.xAxisRange.String())
		//log.Printf("OnRedraw %q fails: unable to proceed given data", drawing.Name) // DEBUG:
		return
	}

	// setup drawing tools
	drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	if drawing.fFillArea {
		drawing.Ctx2D.SetLineWidth(3)
	} else {
		drawing.Ctx2D.SetLineWidth(2)
	}
	drawing.Ctx2D.SetLineJoin(canvas.RoundCanvasLineJoin)
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Lighten(0.8).Hexa())})

	// reduce the cliping area
	drawArea := drawing.ClipArea.Shrink(0, 5)
	//fmt.Printf("clip:%s draw:%s\n", drawing.clipArea, drawArea) // DEBUG:

	// get xfactor & yfactor according to time selection
	xfactor := float64(drawArea.Width) / float64(drawing.xAxisRange.Duration().Duration)
	yfactor := float64(drawArea.Height) / drawing.series.DataRange(drawing.xAxisRange, 10).Delta()
	lowboundary := drawing.series.DataRange(drawing.xAxisRange, 10).Low()
	//fmt.Printf("xfactor:%f yfactor:%f\n", xfactor, yfactor) // DEBUG:

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
			// fmt.Printf("xopen=%d yopen=%d\n", xopen, yopen) // DEBUG:
		}

		xclose = drawArea.O.X + int(xfactor*float64(item.TimeSlice.To.Sub(drawing.xAxisRange.From)))
		yclose := drawArea.O.Y + drawArea.Height - int(yfactor*(item.Close-lowboundary))
		drawing.Ctx2D.LineTo(float64(xclose), float64(yclose))
		//fmt.Printf("xclose=%d yclose=%d\n", xclose, yclose)  // DEBUG:

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
}
