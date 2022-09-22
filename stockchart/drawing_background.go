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
	layer.Ctx2D.SetTextAlign(canvas.EndCanvasTextAlign)
	layer.Ctx2D.SetTextBaseline(canvas.BottomCanvasTextBaseline)
	layer.Ctx2D.SetFont(`20px 'Roboto', sans-serif`)
	layer.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(bootstrapcolor.Gray.Lighten(0.5).Hexa())})
	layer.Ctx2D.FillText("@github.com/sunraylab", float64(layer.ClipArea.End().X-50), float64(layer.ClipArea.End().Y-50), nil)

	// no data
	if drawing.series == nil || drawing.series.TimeSlice().Duration() == nil {
		layer.Ctx2D.SetTextAlign(canvas.CenterCanvasTextAlign)
		layer.Ctx2D.SetTextBaseline(canvas.MiddleCanvasTextBaseline)
		layer.Ctx2D.SetFont(`bold 30px 'Roboto', sans-serif`)
		layer.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.MainColor.Hexa())})
		layer.Ctx2D.FillText("no data", float64(layer.ClipArea.Middle().X), float64(layer.ClipArea.Middle().Y), nil)
	}
}
