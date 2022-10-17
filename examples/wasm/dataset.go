package main

import (
	"math/rand"
	"time"

	"github.com/sunraylab/stockchart/stockchart"
	"github.com/sunraylab/timeline/v2"
)

func BuildRandomDataset(name string, nbdata int, from time.Time, candleDuration time.Duration) *stockchart.DataList {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	dataset := &stockchart.DataList{Name: name, Precision:candleDuration }
	last := from
	open := 20000.0

	for i := 0; i < nbdata; i++ {
		// create a data point
		p := new(stockchart.DataStock)

		// set random but consistent OHLCV
		delta := (r1.Float64() - 0.5) * 100
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
		p.Volume = r1.Float64() * 1000000.0

		// set time
		p.TimeSlice = timeline.MakeTimeSlice(last, candleDuration)

		// skip a data point for the sample
		if i != 10 {
			dataset.Append(p)
		}

		// change the timeslice of a data point for the sample
		if i == 25 {
			p.TimeSlice.ExtendTo(candleDuration * 3.0)
		}

		last = p.TimeSlice.To
		open = p.Close
	}
	return dataset
}
