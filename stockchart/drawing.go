package stockchart

import (
	"fmt"

	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/html/canvas"
	"github.com/gowebapi/webapi/html/htmlevent"
	"github.com/sunraylab/rgb/v2"
)

// Drawing provides primitives
type Drawing struct {
	Name            string    // drawing name, mainly for debugging, but culd be used in the drawing
	BackgroundColor rgb.Color // optional, fully transparent by default
	MainColor       rgb.Color // optional, fully transparent by default

	*Layer           // the parent layer
	series *DataList // the series of data to draw

	OnRedraw func()

	// optional functions to be defined by upper drawings
	OnMouseDown  func(xy Point, event *htmlevent.MouseEvent)
	OnMouseUp    func(xy Point, event *htmlevent.MouseEvent)
	OnMouseMove  func(xy Point, event *htmlevent.MouseEvent)
	OnMouseEnter func(xy Point, event *htmlevent.MouseEvent)
	OnMouseLeave func(xy Point, event *htmlevent.MouseEvent)
	OnWheel      func(event *htmlevent.WheelEvent)
	NeedRedraw   func() bool
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

// DrawTextBox
//
// xy position is defined in a canvas coordinates and corresponds to the corner defined
// by align, the border, the margin and the padding
//
// Font must be set before
func (drawing *Drawing) DrawTextBox(txt string, xy Point, align Align, color rgb.Color, margin int, border int, padding int) {

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
	//fmt.Printf("cliparea:%v, txtbox:%v, bgbox:%v halfpixel:%v\n", drawing.ClipArea, txtbox, bgbox, halfpix) // DEBUG:

	// draw the box and its frame
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(rgb.None.Hexa())})
	drawing.Ctx2D.FillRect(float64(bgbox.O.X), float64(bgbox.O.Y), float64(bgbox.Width), float64(bgbox.Height))
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(drawing.BackgroundColor.Hexa())})
	drawing.Ctx2D.FillRect(float64(txtbox.O.X)-halfpix, float64(txtbox.O.Y)-halfpix, float64(txtbox.Width), float64(txtbox.Height))
	if border > 0 {
		drawing.Ctx2D.SetStrokeStyle(&canvas.Union{Value: js.ValueOf(color.Hexa())})
		drawing.Ctx2D.SetLineWidth(float64(border))
		drawing.Ctx2D.StrokeRect(float64(txtbox.O.X)-halfpix, float64(txtbox.O.Y)-halfpix, float64(txtbox.Width), float64(txtbox.Height))
	}

	// draw the text
	drawing.Ctx2D.SetFillStyle(&canvas.Union{Value: js.ValueOf(color.Hexa())})
	drawing.Ctx2D.FillText(txt, float64(postxt.X), float64(postxt.Y), nil)
}
