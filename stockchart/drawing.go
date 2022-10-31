package stockchart

import (
	"fmt"
	"math"
	"time"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/gowebapi/webapi/html/htmlevent"
	"github.com/sunraylab/rgb/v2"
	"github.com/sunraylab/timeline/v2"
)

// Drawing provides primitives
type Drawing struct {
	Name            string    // drawing name, mainly for debugging, but culd be used in the drawing
	MainColor       rgb.Color // optional, fully transparent by default
	BackgroundColor rgb.Color // optional, fully transparent by default
	drawArea        Rect      // the drawing area of this drawing, the clip area of the layer by default

	*Layer           // the parent layer
	series *DataList // the series of data to draw

	OnRedraw func()

	// optional functions to be defined by upper drawings
	DrawArea func(clipArea Rect) Rect

	OnMouseDown  func(xy Point, event *htmlevent.MouseEvent)
	OnMouseUp    func(xy Point, event *htmlevent.MouseEvent)
	OnMouseMove  func(xy Point, event *htmlevent.MouseEvent)
	OnMouseEnter func(xy Point, event *htmlevent.MouseEvent)
	OnMouseLeave func(xy Point, event *htmlevent.MouseEvent)
	OnWheel      func(event *htmlevent.WheelEvent)
	OnClick      func(xy Point, event *htmlevent.MouseEvent)
	NeedRedraw   func() bool
}

func (drawing Drawing) hasNonEmptySeries() bool {
	return drawing.series != nil && !drawing.series.IsEmpty()
}

func (drawing Drawing) String() string {
	strxrange := "missing"
	if drawing.xAxisRange != nil {
		strxrange = drawing.xAxisRange.String()
	}
	str := fmt.Sprintf("%q xrange:%s", drawing.Name, strxrange)
	return str
}

type Align uint

const (
	AlignCenter Align = 0b00000000
	AlignStart  Align = 0b00000001
	AlignEnd    Align = 0b00000010
	AlignTop    Align = 0b00000100
	AlignBottom Align = 0b00001000
)

// SetMainSeries set or reset the MainSeries of the chart and its drawings. Reset the timerange.
// Need to redraw
func (drawing *Drawing) ResetSubSeries(series *DataList, redrawNow bool) {
	// change the content
	drawing.series = series

	if redrawNow {
		drawing.Redraw()
	}
}

// DrawTextBox draw a text within a box.
//
// xy position is defined in a canvas coordinates and corresponds to the corner defined
// by align, the border, the margin and the padding.
//
// Font must be set before
func (drawing *Drawing) DrawTextBox(txt string, xy Point, align Align, backgroundcolor rgb.Color, textcolor rgb.Color, margin int, border int, padding int) {

	var postxt Point

	// size and position
	tm := drawing.Ctx2D.MeasureText(txt)
	bgbox := Rect{
		Width:  int(tm.Width()) + 2*(margin+border+padding),
		Height: int(tm.ActualBoundingBoxAscent()+tm.ActualBoundingBoxDescent()) + 2*(margin+border+padding)}

	// pixel perfect
	halfpix := 0.0
	if border%2 != 0 {
		halfpix = 0.5
	}

	// X axis
	if (align & AlignStart) > 0 {
		drawing.Ctx2D.SetTextAlign(canvas.StartCanvasTextAlign)
		bgbox.O.X = xy.X
		postxt.X = margin + border + padding

	} else if (align & AlignEnd) > 0 {
		drawing.Ctx2D.SetTextAlign(canvas.EndCanvasTextAlign)
		bgbox.O.X = xy.X - bgbox.Width - 1
		postxt.X = bgbox.Width + 1 - (+margin + border + padding)

	} else {
		drawing.Ctx2D.SetTextAlign(canvas.CenterCanvasTextAlign)
		bgbox.O.X = xy.X - bgbox.Width/2
		postxt.X = bgbox.Width / 2
	}

	// Y axis
	if (align & AlignTop) > 0 {
		drawing.Ctx2D.SetTextBaseline(canvas.TopCanvasTextBaseline)
		bgbox.O.Y = xy.Y
		postxt.Y = margin + border + padding

	} else if (align & AlignBottom) > 0 {
		drawing.Ctx2D.SetTextBaseline(canvas.BottomCanvasTextBaseline)
		bgbox.O.Y = xy.Y + 1 - int(2.0*halfpix) - bgbox.Height
		postxt.Y = +bgbox.Height - (margin + border + padding)

	} else {
		drawing.Ctx2D.SetTextBaseline(canvas.MiddleCanvasTextBaseline)
		bgbox.O.Y = xy.Y - bgbox.Height/2
		postxt.Y = bgbox.Height / 2
	}

	// ensure the box is within the cliparea
	bgbox.Box(drawing.ClipArea)
	postxt.X += bgbox.O.X
	postxt.Y += bgbox.O.Y

	// build the txtbox
	txtbox := bgbox.Shrink(margin+border/2.0, margin+border/2.0)

	// draw the box and its frame
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(rgb.None.Hexa())})
	drawing.Ctx2D.FillRect(float64(bgbox.O.X), float64(bgbox.O.Y), float64(bgbox.Width), float64(bgbox.Height))
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(backgroundcolor.Hexa())})
	drawing.Ctx2D.FillRect(float64(txtbox.O.X)-halfpix, float64(txtbox.O.Y)-halfpix, float64(txtbox.Width), float64(txtbox.Height))
	if border > 0 {
		drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(textcolor.Hexa())})
		drawing.Ctx2D.SetLineWidth(float64(border))
		drawing.Ctx2D.StrokeRect(float64(txtbox.O.X)-halfpix, float64(txtbox.O.Y)-halfpix, float64(txtbox.Width), float64(txtbox.Height))
	}

	// draw the text
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(textcolor.Hexa())})
	drawing.Ctx2D.FillText(txt, float64(postxt.X), float64(postxt.Y), nil)
}

// return the x position of a specific time withing the drawing area and accoring to the xAxisRange
func (drawing *Drawing) xTime(at time.Time) (xpos float64) {

	dx := float64(at.Sub(drawing.xAxisRange.From))
	drange := float64(drawing.xAxisRange.Duration().Duration)
	f := math.Round(float64(drawing.drawArea.Width) * dx / drange)
	return float64(drawing.drawArea.O.X) + f
}

// Draw Vertical Line.
// returns -1 if at is out of range
func (drawing *Drawing) DrawVLine(at time.Time, color rgb.Color, full bool) (xpos float64) {

	if drawing.xAxisRange.WhereIs(at)&timeline.TS_IN > 0 {

		xpos = drawing.xTime(at)

		drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(color.Hexa())})
		drawing.Ctx2D.SetLineWidth(1)
		drawing.Ctx2D.SetLineDash([]float64{})

		drawing.Ctx2D.BeginPath()
		if full {
			drawing.Ctx2D.MoveTo(float64(xpos)+0.5, float64(drawing.ClipArea.O.Y))
			drawing.Ctx2D.LineTo(float64(xpos)+0.5, float64(drawing.ClipArea.O.Y+drawing.ClipArea.Height))
		} else {
			drawing.Ctx2D.MoveTo(float64(xpos)+0.5, float64(drawing.drawArea.O.Y))
			drawing.Ctx2D.LineTo(float64(xpos)+0.5, float64(drawing.drawArea.O.Y+drawing.drawArea.Height))
		}
		drawing.Ctx2D.Stroke()
	} else {
		return -1.0
	}

	return xpos
}
