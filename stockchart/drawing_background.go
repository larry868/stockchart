package stockchart

import (
	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/sunraylab/rgb/bootstrapcolor.go"
	"github.com/sunraylab/timeline/timeslice"
)

type DrawingBackground struct {
	Drawing
}

func NewDrawingBackground(series *DataList) *DrawingBackground {
	drawing := new(DrawingBackground)
	drawing.Name = "copyright"
	drawing.series = series
	drawing.MainColor = *bootstrapcolor.Gray.Darken(0.5)
	drawing.Drawing.OnRedraw = func(layer *drawingLayer) {
		drawing.OnRedraw(layer)
	}
	drawing.Drawing.OnChangeTimeSelection = func(layer *drawingLayer, timesel timeslice.TimeSlice) {
		drawing.OnRedraw(layer)	
	}
	return drawing
}

func (drawing *DrawingBackground) OnRedraw(layer *drawingLayer) {

	// copyright
	layer.ctx2D.SetTextAlign(canvas.EndCanvasTextAlign)
	layer.ctx2D.SetTextBaseline(canvas.BottomCanvasTextBaseline)
	layer.ctx2D.SetFont(`20px 'Roboto', sans-serif`)
	layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(bootstrapcolor.Gray.Lighten(0.5).Hexa())})
	layer.ctx2D.FillText("@github.com/sunraylab", float64(layer.clipArea.End().X-50), float64(layer.clipArea.End().Y-50), nil)

	// no data
	if drawing.series == nil || drawing.series.TimeSlice().Duration() == nil {
		layer.ctx2D.SetTextAlign(canvas.CenterCanvasTextAlign)
		layer.ctx2D.SetTextBaseline(canvas.MiddleCanvasTextBaseline)
		layer.ctx2D.SetFont(`bold 30px 'Roboto', sans-serif`)
		layer.ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
		layer.ctx2D.FillText("no data", float64(layer.clipArea.Middle().X), float64(layer.clipArea.Middle().Y), nil)
	}
}
