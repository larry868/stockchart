package main

import (
	"math/rand"
	"time"

	"github.com/sunraylab/gowebstockchart/stockchart"
)

func BuildRandomDataset() *stockchart.DataList {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	dataset := &stockchart.DataList{Unit: "EUR"}
	var prev *stockchart.DataPoint
	last := time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)
	open := 10000.0

	for i := 0; i < 150; i++ {
		// add a data point
		p := new(stockchart.DataPoint)

		// get a random value
		delta := (r1.Float64() - 0.5) * 50
		p.Open = open
		p.Close = p.Open + delta
		if p.Close < 0 {
			p.Close = -p.Close
		}
		if p.Open > p.Close {
			p.Low = p.Close - r1.Float64()*20
			p.High = p.Open + r1.Float64()*20
		} else {
			p.Low = p.Open - r1.Float64()*20
			p.High = p.Close + r1.Float64()*20
		}
		p.Val = p.Open
		p.Duration = time.Minute

		p.TimeStamp = last.Add(p.Duration)
		p.Next = nil
		p.Prev = prev
		if prev != nil {
			prev.Next = p
		}
		if dataset.Head == nil {
			dataset.Head = p
		}
		dataset.Tail = p
		last = p.TimeStamp
		prev = p

		open = p.Close

		//fmt.Println(p.String())
	}
	return dataset
}
