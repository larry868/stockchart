package stockchart

import (
	"fmt"
	"testing"
)

func TestRectRate(t *testing.T) {

	r := &Rect{O: Point{X: 100, Y: 100}, Width: 100, Height: 100}
	get := r.XRate(150)
	if get != 0.5 {
		t.Errorf("XRate fails: want 0.5, get %v", get)
	}
	get = r.YRate(150)
	if get != 0.5 {
		t.Errorf("YRate fails: want 0.5, get %v", get)
	}

	get = r.XRate(50)
	if get != 0 {
		t.Errorf("XRate fails: want 0, get %v", get)
	}
	get = r.YRate(50)
	if get != 0 {
		t.Errorf("YRate fails: want 0, get %v", get)
	}

	get = r.XRate(200)
	if get != 1 {
		t.Errorf("XRate fails: want 1, get %v", get)
	}
	get = r.YRate(200)
	if get != 1 {
		t.Errorf("YRate fails: want 1, get %v", get)
	}
}

func ExampleRect_And() {
	r1 := &Rect{Width: 100, Height: 100}
	r2 := &Rect{O: Point{X: 20, Y: 20}, Width: 100, Height: 100}
	r3 := r1.And(*r2)
	fmt.Printf("%v\n", r3)
	r3bis := r2.And(*r1)
	fmt.Printf("%v\n", r3bis)
	r4 := r3.And(*r3)
	fmt.Printf("%v\n", r4)
	r5 := &Rect{Width: 10, Height: 10}
	r6 := r5.And(*r3)
	fmt.Printf("%v\n", r6)

	// output:
	// x=20 y=20 width=80 height=80
	// x=20 y=20 width=80 height=80
	// x=20 y=20 width=80 height=80
	// <nil>

}
