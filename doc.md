# Technical documentation


## Selecting a timeslice and a datastock

On a drawing user can select the time range, a time slice, and a single datastock.

Theses informations are saved in `selectedTimeSlice` and `selectedData` properties of the StockChart struct. These values reflects the selection in request for a drawing. That means that these values are used by the drawing process.

They can be updated through an user interaction with the chart, like clicking on a data or changing the navebar selector, or they can be updated through an external request.

## Is there any selection ?

selectedTimeSlice.IsZero() means there's no selected time range, which may occurs for an empty chart, otherwise the selected timeslice is always defined with valid boundaries.

For the selectedData it's different, this is a pointer. A nil value indicate ther's no selection otherwise it points directly to the DataStock within the series.

## Initial Update

``SetTimeRange()`` is called to setup the global time range of the chart. It's called at least during the chart creation and can be set later.

``SetTimeRange()`` calls DoChangeSelTimeSlice to update the `selectedTimeSlice` with the overall time range of the chart. 

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

If drawings layers are not created yet when calling SetTimeRange that's not important because each drawing will get the chart selection info to redraw themselfs.

## Update from outside of the chart

`DoChangeSelTimeSlice()` and ``DoChangeSelData()`` are the most important functions, they allows a redraw of all drawings if the change impact them.

## Updating with user interacting with the chart

Drawings interact only with `selectedTimeSlice` and `selectedData` at the chart level. Some drawings stores a copy of the last selection  they use for a redraw, allowing them to know if they'are concerned by a future change. 

At the end of the drawing process, the event dispatcher will detect a change in `selectedTimeSlice` or `selectedData` and will propagate a `DoChangeSelTimeSlice()` and ``DoChangeSelData()`` at the chart level to redraw layers that needs it.

### Propagation of the changed selection to other layers and other drawings

To enable user interaction with drawings, ``SetEventDispatcher()`` must have been called initially on the layer. Then every drawing must implement one or more ``On{event}()`` function. There's two layers which propagate events: 
 - 2-timeselector layer 
 - 5-hover layer

For these layers all mouse events are activated on the canvas of the layer. When an event occurs on the canvas, it's dispatchd to all drawings having implemented the ``On{event}()`` function. When leaving this function the dispatcher checks if any  selection have changed during the drawings. If any change occurs, the dispatcher calls ``DoChangeSel{timesllice|data}()`` on the chart only one time.

### selectedTimeSlice

 - OnRedraw()   : if this is the first drawing, ``DrawingTimeSelector`` memorizes the ``selectedTimeSlice``
 - OnMouseUp() :  `selectedTimeSlice` is **updated** with the user selection
 - OnMouseMove() : in draging mode only, `selectedTimeSlice` is used to check the user selection stay within valid boundaries, and user selection is updated with an immediate redraw of this drawing only (not the layer nor the full chart)
 - OnWheel() : `selectedTimeSlice` is **updated** according to shift and zoom

