# Technical documentation


## Selecting a timeslice and a datastock

On a drawing user can select the timeslice and a single datastock.

Theses informations are saved in `selectedTimeSlice` and `SelectedData` properties of the StockChart struct. These values reflects the selection in request for a drawing. That means that these value are updated before any rdraw and every leyers and drawing will use this value to redraw.

They can be updated through an user interaction with the chart, like clicking on a data or changing the navebar selector, or they can be updated through an external request.

## Is there any selection ?

selectedTimeSlice.IsZero() means there's no selectedTimeslice, which may occurs for an empty chart, otherwise the selected timeslice is always defined with valid boundaries.

for the SelectedData it's different, this is a pointer. A nil value indicate ther's no selection otherwise it points directly to the DataStock with the series.

## Initial Update

``SetTimeRange()`` is called to setup the overall timeslice to display on the chart. It'ts called at least during the chart creation but can be reset later.

``SetTimeRange()`` must updates the `selectedTimeSlice` to ensure a consistent selection. 

```go 
- SetTimeRange()
    - if selchange is required
        - DoSelChangeTimeSlice()
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

If drawings layers are not created, then layers inits the selection data themself with the layer factory.

## Update from outside of the chart

A call to ``DoSelChangeTimeSlice()`` and ``DoSelChangeData()`` allows a redraw of all drawings if the change impact them.

## Updating with user interacting with the chart

Drawings interact only with `selectedTimeSlice` and `selectedData` at the layer level. These values can be updated by the drawing in response to en event. In this case the event dispatcher will ensure to propagate the change to all layers and also to notify ourside of the chart.

### Propagation of the changed selection to other layers and other drawings

To enable user interaction with drawings, ``SetEventDispatcher()`` must have been called initially on the layer. Then every drawing must implement one or more ``On{event}()`` function. There's two layers which propagate events: 
 - 2-timeselector layer 
 - 5-hover layer

For these layers all mouse events are activated on the canvas of the layer. When an event occurs on the canvas, it's dispatchd to all drawings having implemented the ``On{event}()`` function. When leaving this function the dispatcher checks if any  selection have changed during the drawings. If any change occurs, the dispatcher calls ``DoSelChange{timesllice|data}()`` on the chart only one time.

### selectedTimeSlice

 - OnRedraw()   : if this is the first drawing, ``DrawingTimeSelector`` memorizes the ``selectedTimeSlice``
 - OnMouseUp() :  `selectedTimeSlice` is **updated** with the user selection
 - OnMouseMove() : in draging mode only, `selectedTimeSlice` is used to check the user selection stay within valid boundaries, and user selection is updated with an immediate redraw of this drawing only (not the layer nor the full chart)
 - OnWheel() : `selectedTimeSlice` is **updated** according to shift and zoom

