package main

import (
	"math/rand"
	"time"

	"github.com/sunraylab/gowebstockchart/stockchart"
	"github.com/sunraylab/timeline/v2"
)

func BuildRandomDataset() *stockchart.DataList {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	const candleDuration = time.Minute
	dataset := &stockchart.DataList{Name: "BTC/EUR"}
	last := time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)
	open := 20000.0

	for i := 0; i < 300; i++ {
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
		p.TimeSlice = timeline.MakeTimeslice(last, candleDuration)

		// skip a data point for the sample
		if i != 10 {
			dataset.Append(p)
		}

		// change the timeslice of a data point for the sample
		if i == 25 {
			p.TimeSlice.ToExtend(timeline.Duration(candleDuration * 3.0))
		}

		last = p.TimeSlice.To
		open = p.Close
		//fmt.Println(p.String()) // DEBUG:
	}
	return dataset
}
