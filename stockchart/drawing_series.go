package stockchart

import (
	"log"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb/bootstrapcolor.go"
	"github.com/sunraylab/timeline/timeslice"
)

type DrawingSeries struct {
	Drawing
	fFillArea bool // fullfil the area or draw only the line
}

func NewDrawingSeries(series *DataList, xrange *timeslice.TimeSlice, fFillArea bool) *DrawingSeries {
	drawing := new(DrawingSeries)
	drawing.Name = "series"
	drawing.series = series
	drawing.xAxisRange = xrange
	drawing.fFillArea = fFillArea
	drawing.MainColor = *bootstrapcolor.Blue.Lighten(0.5)
	drawing.Drawing.OnRedraw = func(layer *drawingLayer) {
		drawing.OnRedraw(layer)
	}
	drawing.Drawing.OnChangeTimeSelection = func(layer *drawingLayer, timesel timeslice.TimeSlice) {
		drawing.OnChangeTimeSelection(layer, timesel)
	}
	return drawing
}

func (drawing DrawingSeries) OnRedraw(layer *drawingLayer) {
	if drawing.series == nil || drawing.xAxisRange == nil || drawing.xAxisRange.Duration() == nil || time.Duration(*drawing.xAxisRange.Duration()).Seconds() < 0 {
		log.Printf("OnRedraw %s fails: unable to proceed given data", drawing.Name)
		return
	}

	// setup drawing tools
	layer.ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
	if drawing.fFillArea {
		layer.ctx2D.SetLineWidth(3)
	} else {
		layer.ctx2D.SetLineWidth(2)
	}
	layer.ctx2D.SetLineJoin(canvas.RoundCanvasLineJoin)
	layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Lighten(0.8).Hexa())})

	// reduce the cliping area
	drawArea := layer.clipArea.Shrink(0, 5)
	//fmt.Printf("clip:%s draw:%s\n", layer.clipArea, drawArea) // DEBUG:

	// get xfactor & yfactor according to time selection
	xfactor := float64(drawArea.Width) / time.Duration(*drawing.xAxisRange.Duration()).Seconds()
	yfactor := float64(drawArea.Height) / drawing.series.DataRange(drawing.xAxisRange, 10).Delta()
	lowboundary := drawing.series.DataRange(drawing.xAxisRange, 10).Low()
	//fmt.Printf("xfactor:%f yfactor:%f\n", xfactor, yfactor) // DEBUG:

	// scan all points
	var x0, xclose int
	first := true
	item := drawing.series.Head
	for item != nil {
		// skip items out of range
		if item.TimeStamp.Before(drawing.xAxisRange.From) {
			// scan next item
			item = item.Next
			continue
		}
		if item.TimeStamp.Add(item.Duration).After(drawing.xAxisRange.To) {
			break
		}

		// draw the path
		if first {
			first = false
			xopen := drawArea.O.X + int(xfactor*item.TimeStamp.Sub(drawing.xAxisRange.From).Seconds())
			yopen := drawArea.O.Y + drawArea.Height - int(yfactor*(item.Open-lowboundary))
			layer.ctx2D.MoveTo(float64(xopen), float64(yopen))
			layer.ctx2D.BeginPath()
			layer.ctx2D.LineTo(float64(xopen), float64(yopen))
			x0 = xopen
			// fmt.Printf("xopen=%d yopen=%d\n", xopen, yopen) // DEBUG:
		}

		xclose = drawArea.O.X + int(xfactor*item.TimeStamp.Add(item.Duration).Sub(drawing.xAxisRange.From).Seconds())
		yclose := drawArea.O.Y + drawArea.Height - int(yfactor*(item.Close-lowboundary))
		layer.ctx2D.LineTo(float64(xclose), float64(yclose))
		//fmt.Printf("xclose=%d yclose=%d\n", xclose, yclose)  // DEBUG:

		// scan next item
		item = item.Next
	}

	// draw the top line
	layer.ctx2D.Stroke()

	// then draw the area
	if drawing.fFillArea {
		layer.ctx2D.LineTo(float64(xclose), float64(layer.clipArea.End().Y))
		layer.ctx2D.LineTo(float64(x0), float64(layer.clipArea.End().Y))
		layer.ctx2D.ClosePath()
		fillrule := canvas.NonzeroCanvasFillRule
		layer.ctx2D.Fill(&fillrule)
	}
}

// OnChangeTimeSelection
func (pdrawing *DrawingSeries) OnChangeTimeSelection(layer *drawingLayer, timesel timeslice.TimeSlice) {
	*pdrawing.xAxisRange = timesel
	pdrawing.OnRedraw(layer)
}
