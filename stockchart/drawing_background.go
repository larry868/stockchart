package stockchart

import (
	"github.com/larry868/rgb"
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
		drawing.onRedraw()
	}
	return drawing
}

func (drawing *DrawingBackground) onRedraw() {
	// copyright
	drawing.Ctx2D.SetFont(`20px 'Roboto', sans-serif`)
	drawing.DrawTextBox("@github.com/larry868", Point{X: drawing.ClipArea.End().X - 100, Y: drawing.ClipArea.End().Y - 100}, AlignEnd|AlignBottom, rgb.None, drawing.MainColor, 0, 0, 0)

	// no data
	if drawing.series.IsEmpty() {
		drawing.Ctx2D.SetFont(`bold 30px 'Roboto', sans-serif`)
		drawing.DrawTextBox("no data", Point{X: drawing.ClipArea.Middle().X, Y: drawing.ClipArea.Middle().Y}, AlignCenter, rgb.White, drawing.MainColor, 0, 0, 0)
	}
}
