package stockchart

import (
	"fmt"
	"log"
	"time"

	"github.com/sunraylab/datarange"
	"github.com/sunraylab/timeline/v2"
)

// DataStock is a value at a given timestamp.
// It's linked with previous and following data
type DataStock struct {
	timeline.TimeSlice `json:"timeslice"`

	Open   float64 `json:"open"`
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volule"`

	Next *DataStock `json:"-"` // going to the head
	Prev *DataStock `json:"-"` // going to the tail
}

func (dp *DataStock) String() string {
	if dp == nil {
		return "nil"
	}
	str := fmt.Sprintf("o=%v h=%v l=%v c=%v v=%v at:%s\n", dp.Open, dp.High, dp.Low, dp.Close, dp.Volume, dp.TimeSlice)
	return str
}

// DataList is a time ordered chained list of DataPoint.
// We assume that ordered points are linked in chronological order
type DataList struct {
	Name              string        // the name of the data list, usually the name of the pair and its precision
	Tail *DataStock // the tail !-----...
	Head *DataStock // ...-----! the head
}

func (dl DataList) String() string {
	str := fmt.Sprintf("list size:%v, timeslice:%s, datarange:%s", dl.Size(), dl.TimeSlice(), dl.DataRange(nil, 0))
	return str
}

func (dl DataList) IsEmpty() bool {
	return dl.Head == nil
}

// func (pdl *DataList) Reset() {
// 	pdl.Tail = nil
// 	pdl.Head = nil
// 	pdl.Aggregate = nil
// 	pdl.AggregateLevel = 0
// 	pdl.AggregateDuration = 0
// }

func (pdl DataList) Size() (size int) {
	scan := pdl.Head
	for scan != nil {
		size++
		if scan == pdl.Tail {
			break
		}
		scan = scan.Prev
	}
	return size
}

// Append a dataPoint to the head
func (dl *DataList) Append(newdata *DataStock) {
	// add the data point to the list
	newdata.Next = nil
	newdata.Prev = dl.Head
	// link previous data
	if newdata.Prev != nil {
		newdata.Prev.Next = newdata
	}
	// first data
	if dl.Tail == nil {
		dl.Tail = newdata
	}
	// update head
	dl.Head = newdata
}

// Insert a dataPoint at the right position according to dates
// backward lookup to find the insert position
func (dl *DataList) Insert(newdata *DataStock) {
	if dl.Head == nil || dl.Tail == nil {
		dl.Tail = newdata
		dl.Head = newdata
		return
	}

	// scan backward
	scan := dl.Head
	for scan != nil {
		if scan.To.Equal(newdata.To) {
			// houston on a un pb
			log.Printf("insert fails because newdata end at the same time of an existing one: %s", newdata.To)
			return
		} else if scan.To.Before(newdata.To) {
			newdata.Next = scan.Next
			newdata.Prev = scan
			if scan.Next == nil {
				dl.Head = newdata
			} else {
				scan.Next.Prev = newdata
			}
			scan.Next = newdata
			return
		}
		scan = scan.Prev
	}
	// we're at the tail
	dl.Tail.Prev = newdata
	newdata.Next = dl.Tail
	newdata.Prev = nil
	dl.Tail = newdata
}

// return the dataPoint at t time, nil if no points found
func (dl DataList) GetDataAt(t time.Time) (data *DataStock) {
	item := dl.Tail
	for item != nil {
		if (t.Equal(item.TimeSlice.From) || t.After(item.TimeSlice.From)) && (t.Equal(item.TimeSlice.To) || t.Before(item.TimeSlice.To)) {
			return item
		}
		item = item.Next
	}
	return item
}

// TimeSlice returns the time boundaries of the DataList, between the Head and the Tail.
//
// returns an empty timeslice if the list is empty or if missing head or tail
func (dl DataList) TimeSlice() timeline.TimeSlice {
	var ts timeline.TimeSlice
	if dl.Tail != nil && dl.Head != nil {
		ts.From = dl.Tail.TimeSlice.From
		ts.To = dl.Head.TimeSlice.To
	}
	return ts
}

// DataRange returns the data boundaries of the DataList, scanning all datapoint between the timeslice boundaries
//
//	ts == nil scan all data points between the Head and the Tail:
//	if maxSteps == 0 the returned datarange doesn't have any stepzise.
//	if maxSteps > 0 the returned datarange gets a stepzise and boudaries are rounded.
//
// returns an empty datarange if the list is empty or if missing head or tail.
func (dl DataList) DataRange(ts *timeline.TimeSlice, maxSteps uint) (dr datarange.DataRange) {
	var low, high float64
	item := dl.Tail
	for item != nil {
		// if ts == nil || ((item.TimeSlice.From.Equal(ts.From) || item.TimeSlice.From.After(ts.From)) && (item.TimeSlice.To.Equal(ts.To) || item.TimeSlice.To.Before(ts.To))) {
		if ts == nil || (ts.WhereIs(item.From)|ts.WhereIs(item.To)&timeline.TS_IN > 0) {
			if low == 0 || item.Low < low {
				low = item.Low
			}
			if item.High > high {
				high = item.High
			}
		}
		item = item.Next
	}

	dr = datarange.Make(low, high, -float64(maxSteps), dl.Name)
	return dr
}

// VolumeDataRange returns the data boundaries of the DataList, scanning all datapoint between the timeslice boundaries
//
//	ts == nil scan all data points between the Head and the Tail:
//	if maxSteps == 0 the returned datarange doesn t have any stepzise.
//	if maxSteps > 0 the returned datarange gets a stepzise and boudaries are rounded.
//
// returns an empty datarange if the list is empty or if missing head or tail.
func (dl DataList) VolumeDataRange(ts *timeline.TimeSlice, maxSteps uint) (dr datarange.DataRange) {
	var low, high float64
	item := dl.Tail
	for item != nil {
		if ts == nil || (ts.WhereIs(item.From)|ts.WhereIs(item.To)&timeline.TS_IN > 0) {
			if low == 0 || item.Volume < low {
				low = item.Volume
			}
			if item.Volume > high {
				high = item.Volume
			}
		}
		item = item.Next
	}

	dr = datarange.Make(low, high, -float64(maxSteps), dl.Name)
	return dr
}
