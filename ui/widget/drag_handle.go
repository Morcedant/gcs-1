package widget

import (
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/unison"
)

// DragHandle provides a simple draggable handle.
type DragHandle struct {
	unison.Panel
	svg      *unison.DrawableSVG
	data     map[string]any
	rollover bool
}

// NewDragHandle creates a new draggable handle widget.
func NewDragHandle(data map[string]any) *DragHandle {
	h := &DragHandle{data: data}
	h.Self = h
	h.DrawCallback = h.draw
	h.MouseEnterCallback = h.mouseEnter
	h.MouseExitCallback = h.mouseExit
	h.MouseDownCallback = h.mouseDown
	h.MouseDragCallback = h.mouseDrag
	baseline := unison.DefaultButtonTheme.Font.Baseline()
	size := unison.NewSize(baseline, baseline)
	h.svg = &unison.DrawableSVG{
		SVG:  res.GripSVG,
		Size: *size.GrowToInteger(),
	}
	h.SetSizer(h.size)
	h.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.MiddleAlignment,
		VAlign: unison.StartAlignment,
	})
	h.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: 3}))
	return h
}

func (h *DragHandle) size(_ unison.Size) (min, pref, max unison.Size) {
	s := h.svg.LogicalSize()
	s.AddInsets(h.Border().Insets())
	s.GrowToInteger()
	return s, s, s
}

func (h *DragHandle) draw(gc *unison.Canvas, rect unison.Rect) {
	var ink unison.Ink
	if h.rollover {
		ink = unison.IconButtonRolloverColor
	} else {
		ink = unison.IconButtonColor
	}
	h.svg.DrawInRect(gc, h.ContentRect(false), nil, ink.Paint(gc, rect, unison.Fill))
}

func (h *DragHandle) mouseEnter(_ unison.Point, _ unison.Modifiers) bool {
	h.rollover = true
	h.MarkForRedraw()
	return true
}

func (h *DragHandle) mouseExit() bool {
	h.rollover = false
	h.MarkForRedraw()
	return true
}

func (h *DragHandle) mouseDown(_ unison.Point, _, _ int, _ unison.Modifiers) bool {
	return true
}

func (h *DragHandle) mouseDrag(where unison.Point, _ int, _ unison.Modifiers) bool {
	if h.IsDragGesture(where) {
		size := h.svg.LogicalSize()
		h.StartDataDrag(&unison.DragData{
			Data:     h.data,
			Drawable: h.svg,
			Ink:      unison.IconButtonColor,
			Offset:   unison.Point{X: -size.Width / 2, Y: -size.Height / 2},
		})
	}
	return true
}
