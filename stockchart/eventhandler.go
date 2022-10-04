package stockchart

type evtHandler int

const (
	evt_None       evtHandler = 0b0000000000000000
	evt_All        evtHandler = 0b1111111111111111
	evt_MouseUp    evtHandler = 0b0000000000000001
	evt_MouseDown  evtHandler = 0b0000000000000010
	evt_MouseMove  evtHandler = 0b0000000000000100
	evt_MouseEnter evtHandler = 0b0000000000001000
	evt_MouseLeave evtHandler = 0b0000000000010000
	evt_Wheel      evtHandler = 0b0000000000100000
	evt_Click      evtHandler = 0b0000000001000000
)

func (layer *Layer) HandledEvents() evtHandler {
	var e evtHandler
	for _, drawing := range layer.drawings {
		if drawing.OnMouseUp != nil {
			e |= evt_MouseUp
		}
		if drawing.OnMouseDown != nil {
			e |= evt_MouseDown
		}
		if drawing.OnMouseMove != nil {
			e |= evt_MouseMove
		}
		if drawing.OnMouseEnter != nil {
			e |= evt_MouseEnter
		}
		if drawing.OnMouseLeave != nil {
			e |= evt_MouseLeave
		}
		if drawing.OnWheel != nil {
			e |= evt_Wheel
		}
		if drawing.OnClick != nil {
			e |= evt_Click
		}
	}
	return e
}
