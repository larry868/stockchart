# Technical documentation


## Selecting a timeslice and a datastock

Users can select the time range and a single datastock on the chart.

The time range is a TimeSlice.

Both informations are saved in `selectedTimeSlice` and `selectedData` properties of the StockChart struct. These values reflects the selection in request for a drawing. That means these values are used by the drawing process.

They can be updated through an user interaction with the chart, like clicking on a data or changing the navebar selector, or they can be updated through an external request.

## xAxisRange 

xAxisRange is the timeslice used by drawings, it's pointing to 2 kind of ranges according to layers:

|layer          | xAxisRange | Comment |
|-|-|-|
|1-navbar       | &chart.timeRange |
|2-timeselector | &chart.timeRange |
|3-DrawingYGrid | &chart.selectedTimeSlice |
|4-chart        | &chart.selectedTimeSlice | SubChart used the same xAxisRange
|5-hover        | &chart.selectedTimeSlice |


## Is there any selection ?

selectedTimeSlice.IsZero() means there's no selected time range, which may occurs for an empty chart, otherwise the selected timeslice is always defined with valid boundaries.

For the selectedData it's different, this is a pointer. A nil value indicate ther's no selection otherwise it points directly to the DataStock within the series.

## Initial Update

``SetTimeRange()`` is called to setup the global time range of the chart. It's called at least during the chart creation and can be set later.

```go 
- SetTimeRange()
    - if selchange is required
        - DoChangeSelTimeSlice()
            - UPDATE `selectedTimeSlice` in the chart
            - RedrawOnlyNeeds() 
                - for each layer
                    - RedrawOnlyNeeds() 
                        - for each drawing
                            - if at least one drawing needs a redraw (call to NeedRedraw())
                                - Redraw() all the layer
                        - update selection on the layer with the selection at the chart level `selectedTimeSlice` & `selectedData`
            - if notify requested
                - NotifySelChangeTimeSlice() 

```

If drawings on layers are not created when calling ``SetTimeRange`` has no effect. Each drawing will get the chart selection data to redraw themselves.

## Update from outside of the chart

`DoChangeSelTimeSlice()` and ``DoChangeSelData()`` redraw all drawings if any change occured.

## Update with user interacting with the chart

Drawings uses `drawing.xAxisRange` and `chart.selectedData` for a redraw.

Some drawings can update `chart.selectedData` or `selectedTimeSlice` according to user interactions.

At the end of the drawing process, the event dispatcher will detect a change in `selectedTimeSlice` or `selectedData` and will propagate a `DoChangeSelTimeSlice()` and ``DoChangeSelData()`` at the chart level to redraw layers that needs to.

### Propagation of changes to other layers and other drawings

To enable user interaction with drawings ``SetEventDispatcher()`` must have been called initially on the layer. Then every drawing must implement one or more ``On{event}()`` function. For the time behing two layers are dispatching events: 
 - 2-timeselector layer: user can change the selected timeslice
 - 5-hover layer: user can select a single data point

For these layers all mouse events are activated on the canvas. When an event occurs on the canvas it's dispatchd to all drawings having implemented the ``On{event}()`` function. When leaving this function the dispatcher checks if any selection have changed during the drawings. If any change occurs, the dispatcher calls ``DoChangeSel{timesllice|data}()`` on the chart only one time.

|layer          | chart                         | xAxisRange             | Dispatch |
|-|-|-|-|
|1-navbar       | DrawingSeries                 | &chart.timeRange          | no    | 
|1-navbar       | DrawingXGrid(not time dep.)   | &chart.timeRange          | no    |
|2-timeselector | DrawingTimeSelector           | &chart.timeRange          | Yes (1) |
|3-DrawingYGrid | DrawingYGrid                  | &chart.selectedTimeSlice  | no    |
|4-chart        | DrawingBackground             | &chart.selectedTimeSlice  | no    |
|4-chart        | DrawingYGrid                  | &chart.selectedTimeSlice  | no    |
|4-chart        | DrawingXGrid(time dep.)       | &chart.selectedTimeSlice  | no    |
|4-chart        | DrawingBars                   | &chart.selectedTimeSlice  | no    |
|4-chart        | DrawingCandles (subchart)     | &chart.selectedTimeSlice  | no    |
|4-chart        | DrawingCandles (mainchar)     | &chart.selectedTimeSlice  | no    |
|5-hover        | DrawingHoverCandles           | &chart.selectedTimeSlice  | Yes (2) |

#### (1) Dispatch on the timeselctor

- OnRedraw() : uses drawing.chart.selectedTimeSlice for the dragtimeSelection
- OnMouseUp() : `chart.selectedTimeSlice` is **updated** with the user selection
- OnWheel() : `chart.selectedTimeSlice` is **updated** according to shift and zoom

#### (2) Dispatch on the hover

- OnClick() : `chart.selectedData` is **updated** with the user selection


## RedrawOnlyNeeds

The RedrawOnlyNeeds process is used to optimze redrawing.

```go
DoChange[SelTimeslice|SelData|Timezone]()
    Update chart data
    RedrawOnlyNeeds() 
        scan all layers
            scan all drawings
            if at least one drawing.NeedRedraw() on the layer
                then Redraw() the layer
                    scan all drawings
                        call OnRedraw() on each layer

    if func is defined, NotifyChange[SelTimeslice|SelData] func 

```

To answer to NeedRedraw(), each drawing need to compare their local data with the chart data.
If the function is not implemented then the layer consider that the drawing do NOT need a redraw, this is the case for layer not dependant on chart selection.

|layer          | chart            | NeedRedraw Implemented | NeedRedraw  |
|-|-|-|-|
|1-navbar       | DrawingSeries                 | Yes   | SelectedData changed |
|1-navbar       | DrawingXGrid(not time dep.)   | Yes   | localZone changed |
|2-timeselector | DrawingTimeSelector           | YES   | dragtimeSelection no more equal to chart.selectedTimeSlice |
|3-DrawingYGrid | DrawingYGrid                  | YES   | yrange(chart.selectedTimeSlice) changed |
|4-chart        | DrawingBackground             | No    | No |
|4-chart        | DrawingYGrid                  | YES   | yrange(chart.selectedTimeSlice) changed |
|4-chart        | DrawingXGrid(time dep.)       | Yes   | localZone changed OR SelectedTimeSlice changed |
|4-chart        | DrawingBars                   | Yes   | SelectedTimeSlice changed |
|4-chart        | DrawingCandles (subchart)     | Yes   | SelectedTimeSlice changed OR SelectedData changed |
|4-chart        | DrawingCandles (mainchar)     | Yes   | SelectedTimeSlice changed OR SelectedData changed |
|5-hover        | DrawingHoverCandles           | No    | No |

All Drawing's series are pointing to the main series referenced at chart level, unless for subchart. This main series are updated by `RsetMainSeries()`.

> NOTA: OnRedraw redraw all drawings without exception



