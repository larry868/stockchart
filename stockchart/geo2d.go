package stockchart

import (
	"fmt"
	"log"
	"math"
)

// A simple point in a 2D plan
type Point struct {
	X int
	Y int
}

// IsIn returns true if the point is inside the rect
func (pt Point) IsIn(r Rect) bool {
	inx := (pt.X >= r.O.X) && (pt.X <= (r.O.X + int(r.Width)))
	iny := (pt.Y >= r.O.Y) && (pt.Y <= (r.O.Y + int(r.Height)))
	return inx && iny
}

var Origin = Point{X: 0, Y: 0}

// A simple rect in a 2d plan
type Rect struct {
	O      Point // the orgin of the rect, can be negative
	Width  int
	Height int
}

// String Interface for r
func (r Rect) String() string {
	return fmt.Sprintf("x=%v y=%v width=%v height=%v", r.O.X, r.O.Y, r.Width, r.Height)
}

// Equal
func (pr1 *Rect) Equal(pr2 *Rect) bool {
	if pr1 == nil && pr2 == nil {
		return true
	}
	if pr1 == nil || pr2 == nil {
		return false
	}
	return (pr1.O.X == pr2.O.X) && (pr1.O.Y == pr2.O.Y) && (pr1.Width == pr2.Width) && (pr1.Height == pr2.Height)
}

// returns the Middle of the rect
func (r Rect) Middle() Point {
	return Point{X: r.O.X + int(math.Round(float64(r.Width)/2.0)), Y: r.O.Y + int(math.Round(float64(r.Height)/2.0))}
}

// returns the endpoint of the rect
func (r Rect) End() Point {
	return Point{X: r.O.X + r.Width, Y: r.O.Y + r.Height}
}

func (r Rect) Shift(xy Point) {
	r.O.X += xy.X
	r.O.Y += xy.Y
}

// FlipPositive flip the rect to make it positive in both dimensions
// with at least a height and and width of 1
func (pr *Rect) FlipPositive() {
	if pr.Height == 0 {
		pr.Height = 1
	} else if pr.Height < 0 {
		pr.O.Y = pr.O.Y + pr.Height
		pr.Height = -pr.Height
	}

	if pr.Width == 0 {
		pr.Width = 1
	} else if pr.Width < 0 {
		pr.O.X = pr.O.X + pr.Width
		pr.Width = -pr.Width
	}
}

// returns where is y inside the rect
//
//	if y is over EndPoint.Y then return 1
//	if y is before O.Y then return 0
//	if YSize is zero then return 0
func (r Rect) YRate(y int) float64 {
	if r.Height == 0 {
		return 0
	}
	rate := float64(y-r.O.Y) / float64(r.Height)
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		rate = 1
	}
	return rate
}

// returns where is y inside the rect
//
//	if x is over End.Y then return 1
//	if x is before O.X then return 0
//	if XSize is zero then return 0
func (r Rect) XRate(x int) float64 {
	if r.Width == 0 {
		return 0
	}
	rate := float64(x-r.O.X) / float64(r.Width)
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		rate = 1
	}
	return rate
}

// returns a Shrinked rect.
//
//	if x or y are <0 then the rect is expanded
//	x and y are expressed in Rect unit, not in rate
//	if x or y are greater than the size then rect become a single point with zero size
func (r Rect) Shrink(x int, y int) *Rect {
	xo := r.O.X + x
	if xo < 0 {
		xo = 0
	}
	yo := r.O.Y + y
	if yo < 0 {
		yo = 0
	}

	o := &Point{X: xo, Y: yo}

	xs := int(r.Width) - x - x
	if xs < 0 {
		xs = 0
	}
	ys := int(r.Height) - y - y
	if ys < 0 {
		ys = 0
	}

	return &Rect{O: *o, Width: xs, Height: ys}
}

// And returns a rect which correspond to common area of thisRect and anotherRect.
// rteurns nul if there's no common area
func (thisRect Rect) And(anotherRect Rect) (AndRect *Rect) {
	xo := imax(anotherRect.O.X, thisRect.O.X)
	if xo > anotherRect.End().X || xo > thisRect.End().X {
		return nil
	}
	xe := imin(anotherRect.End().X, thisRect.End().X)
	if xe < anotherRect.O.X || xe < thisRect.O.X {
		return nil
	}
	yo := imax(anotherRect.O.Y, thisRect.O.Y)
	if yo > anotherRect.End().Y || yo > thisRect.End().Y {
		return nil
	}
	ye := imin(anotherRect.End().Y, thisRect.End().Y)
	if ye < anotherRect.O.Y || ye < thisRect.O.Y {
		return nil
	}
	AndRect = &Rect{O: Point{X: xo, Y: yo}, Width: xe - xo, Height: ye - yo}
	if !anotherRect.Equal(AndRect) {
		log.Printf("thisRect:%s, another:%s And=%s\n", thisRect, anotherRect, AndRect) // DEBUG:
	}
	return AndRect
}

// BoundX returns X bounded by the rect boundaries
func (r Rect) BoundX(x int) int {
	if x < r.O.X {
		x = r.O.X
	} else if x > r.O.X+int(r.Width) {
		x = r.O.X + int(r.Width)
	}
	return x
}

// BoundY returns y bounded by the rect boundaries
func (r Rect) BoundY(y int) int {
	if y < r.O.Y {
		y = r.O.Y
	} else if y > r.O.Y+int(r.Height) {
		y = r.O.Y + int(r.Height)
	}
	return y
}
