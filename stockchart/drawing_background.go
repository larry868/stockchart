package stockchart

import (
	"github.com/sunraylab/rgb/v2"
)

type DrawingBackground struct {
	Drawing
}

func NewDrawingBackground(series *DataList) *DrawingBackground {
	drawing := new(DrawingBackground)
	drawing.Name = "background & copyright"
	drawing.series = series
	drawing.MainColor = rgb.Silver.Lighten(0.7)
	drawing.Drawing.OnRedraw = func() {
		drawing.OnRedraw()
	}
	return drawing
}

func (drawing *DrawingBackground) OnRedraw() {

	// copyright
	drawing.Ctx2D.SetFont(`20px 'Roboto', sans-serif`)
	drawing.DrawTextBox("@github.com/sunraylab", Point{X: drawing.ClipArea.End().X - 50, Y: drawing.ClipArea.End().Y - 50}, AlignEnd|AlignBottom, drawing.MainColor, 0, 0, 0)

	// no data
	if drawing.series == nil || drawing.series.TimeSlice().Duration() == nil {
		drawing.Ctx2D.SetFont(`bold 30px 'Roboto', sans-serif`)
		drawing.DrawTextBox("no data", Point{X: drawing.ClipArea.Middle().X, Y: drawing.ClipArea.Middle().Y}, AlignCenter, drawing.MainColor, 0, 0, 0)
	}
}
