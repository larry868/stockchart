package stockchart

import (
	"log"

	"github.com/gowebapi/webapi/html/htmlevent"
	"github.com/sunraylab/rgb/bootstrapcolor.go"
	"github.com/sunraylab/timeline/timeslice"
)

type DrawingHoverCandles struct {
	Drawing
	hoverData *DataPoint // the data hovered
}

func NewDrawingHoverCandles(series *DataList, xrange *timeslice.TimeSlice) *DrawingHoverCandles {
	drawing := new(DrawingHoverCandles)
	drawing.Name = "hovrcandles"
	drawing.series = series
	drawing.xAxisRange = xrange
	drawing.MainColor = bootstrapcolor.Pink
	/*drawing.Drawing.OnRedraw = func(layer *drawingLayer) {
		drawing.OnRedraw(layer)
	}*/
	drawing.Drawing.OnMouseMove = func(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) (fRedraw bool) {
		return drawing.OnMouseMove(layer, xy, event)
	}

	return drawing
}

func (pdrawing *DrawingHoverCandles) OnMouseMove(layer *drawingLayer, xy Point, event *htmlevent.MouseEvent) (fRedraw bool) {
	if pdrawing.xAxisRange == nil {
		log.Printf("OnMouseMove %q fails: unable to proceed given data", pdrawing.Name)
		return false
	}
	//fmt.Printf("%q mousemove xy:%v\n", pdrawing.Name xy) //DEBUG:
	//fmt.Printf("buttons from:%v to:%v\n", pchart.fromButton, pchart.toButton) //DEBUG:

	// get the candle
	rate := layer.clipArea.XRate(xy.X)
	postime := pdrawing.xAxisRange.WhatTime(rate)
	hoverData := pdrawing.series.GetAt(postime)

	layer.Clear()
	if hoverData != nil {
		// draw a line at the selected

	}

	// change cursor if we start overing a button
	/*if (xy.IsIn(pdrawing.fromButton) || xy.IsIn(pdrawing.toButton)) && !pdrawing.resizeCursor {
		layer.ctx2D.Canvas().AttributeStyleMap().Set("cursor", &typedom.Union{Value: js.ValueOf(`col-resize`)})
		pdrawing.resizeCursor = true
	}*/

	return true
}
