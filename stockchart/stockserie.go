package stockchart

import (
	"fmt"
	"time"

	"github.com/sunraylab/datarange"
	"github.com/sunraylab/timeline/timeslice"
)

// DataPoint is a value at a given timestamp.
// It's linked with previous and following datapoint
type DataPoint struct {
	Val   float64
	Open  float64
	Low   float64
	High  float64
	Close float64

	TimeStamp time.Time
	Duration  time.Duration
	Next      *DataPoint
	Prev      *DataPoint
}

func (dp DataPoint) String() string {
	str := fmt.Sprintf("o=%f l=%f h=%f c=%f from:%s\n", dp.Open, dp.Low, dp.High, dp.Close, dp.TimeStamp)
	return str
}

// DataList is a chained list of DataPoint.
// We assume that ordered points are linked in chronological order
type DataList struct {
	Unit string
	Head *DataPoint
	Tail *DataPoint
}

// return the dataPoint at t time, nil if no points found
func (l DataList) GetAt(t time.Time) (data *DataPoint) {
	item := l.Head
	for item != nil {
		if (t.Equal(item.TimeStamp) || t.After(item.TimeStamp)) && (t.Equal(item.TimeStamp.Add(item.Duration)) || t.Before(item.TimeStamp.Add(item.Duration))) {
			return item
		}
		item = item.Next
	}
	return data
}

// TimeSlice returns the time boundaries of the DataList, between the Head and the Tail.
//
// returns an empty timeslice if the list is empty or if missing head or tail
func (l DataList) TimeSlice() timeslice.TimeSlice {
	var ts timeslice.TimeSlice

	if l.Head != nil {
		ts.From = l.Head.TimeStamp
		if l.Tail != nil {
			ts.To = l.Tail.TimeStamp.Add(l.Tail.Duration)
		}
	}
	return ts
}

// DataRange returns the data boundaries of the DataList, scanning all datapoint between the Head and the Tail:
//
//   - if maxSteps == 0 // the returned datarange doesn't have any stepzise.
//   - if maxSteps > 0 // the returned datarange gets a stepzise and boudaries are rounded.
//
// returns an empty datarange if the list is empty or if missing head or tail.
func (l DataList) DataRange(ts *timeslice.TimeSlice, maxSteps uint) (dr datarange.DataRange) {
	var low, high float64
	item := l.Head
	for item != nil {
		if ts == nil || ((item.TimeStamp.Equal(ts.From) || item.TimeStamp.After(ts.From)) && (item.TimeStamp.Equal(ts.To) || item.TimeStamp.Before(ts.To))) {
			if low == 0 || item.Low < low {
				low = item.Low
			}
			if item.High > high {
				high = item.High
			}
		}
		item = item.Next
	}

	dr = datarange.Build(low, high, -float64(maxSteps), l.Unit)
	return dr
}
