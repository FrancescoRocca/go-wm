package wm

import (
	"log"

	"github.com/francescorocca/go-xlib"
)

func (wm *WindowManager) handleEvent(ev xlib.Event) {
	switch ev.Type() {

	case xlib.KeyPress:
		wm.handleKeyPress(ev)

	case xlib.ButtonPress:
		wm.handleButtonPress(ev)

	case xlib.MotionNotify:
		wm.handleMotionNotify(ev)

	case xlib.ButtonRelease:
		wm.handleButtonRelease(ev)

	case xlib.MapRequest:
		wm.handleMapRequest(ev)

	case xlib.ConfigureRequest:
		wm.handleConfigureRequest(ev)
	}
}

func (wm *WindowManager) handleMapRequest(ev xlib.Event) {
	mr := ev.AsMapRequestEvent()
	if err := wm.display.MapWindow(mr.Window); err != nil {
		log.Fatal("Unable to MapWindow:", err)
	}
}

func (wm *WindowManager) handleConfigureRequest(ev xlib.Event) {
	cr := ev.AsConfigureRequestEvent()

	var mask uint
	var changes xlib.WindowChanges

	if (cr.ValueMask & xlib.CWX) != 0 {
		mask |= xlib.CWX
		changes.X = cr.X
	}
	if (cr.ValueMask & xlib.CWY) != 0 {
		mask |= xlib.CWY
		changes.Y = cr.Y
	}
	if (cr.ValueMask & xlib.CWWidth) != 0 {
		mask |= xlib.CWWidth
		changes.Width = cr.Width
	}
	if (cr.ValueMask & xlib.CWHeight) != 0 {
		mask |= xlib.CWHeight
		changes.Height = cr.Height
	}
	if (cr.ValueMask & xlib.CWBorderWidth) != 0 {
		mask |= xlib.CWBorderWidth
		changes.BorderWidth = 2
	}
	if (cr.ValueMask & xlib.CWSibling) != 0 {
		mask |= xlib.CWSibling
		changes.Sibling = cr.Above
	}
	if (cr.ValueMask & xlib.CWStackMode) != 0 {
		mask |= xlib.CWStackMode
		changes.StackMode = cr.Detail
	}

	if err := wm.display.ConfigureWindow(cr.Window, mask, changes); err != nil {
		log.Fatal("Unable to ConfigureWindow", err)
	}
}
