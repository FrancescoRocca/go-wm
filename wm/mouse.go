package wm

import (
	"log"

	"github.com/francescorocca/go-xlib"
)

type mouseState struct {
	start xlib.ButtonEvent
	attr  xlib.WindowAttributes
}

func (wm *WindowManager) handleButtonPress(ev xlib.Event) {
	be := ev.AsButtonEvent()
	log.Printf("ButtonPress button=%d subwindow=%d xroot=%d yroot=%d state=0x%x\n",
		be.Button, be.Subwindow, be.XRoot, be.YRoot, be.State)

	if be.Subwindow != xlib.NoneWindow() {
		attr, err := wm.display.GetWindowAttributes(be.Subwindow)
		if err != nil {
			log.Println("GetWindowAttributes error:", err)
		}
		wm.ms.attr = attr
		wm.ms.start = *be

		wm.display.GrabPointer(
			be.Subwindow,
			xlib.True,
			xlib.ButtonPressMask|xlib.ButtonReleaseMask|xlib.PointerMotionMask,
			xlib.GrabModeAsync,
			xlib.GrabModeAsync,
			xlib.NoneWindow(),
			xlib.NoneCursor(),
			xlib.CurrentTime,
		)
	}
}

func (wm *WindowManager) handleMotionNotify(ev xlib.Event) {
	if wm.ms.start.Subwindow == xlib.NoneWindow() {
		return
	}

	me := ev.AsButtonEvent()
	xdiff := me.XRoot - wm.ms.start.XRoot
	ydiff := me.YRoot - wm.ms.start.YRoot

	x := wm.ms.attr.X
	y := wm.ms.attr.Y
	w := uint(wm.ms.attr.Width)
	h := uint(wm.ms.attr.Height)

	switch wm.ms.start.Button {
	case 1:
		// WIN + left click = MOVE
		x += xdiff
		y += ydiff

	case 3:
		// WIN + right click = RESIZE
		if int(w)+int(xdiff) > 1 {
			w = uint(int(wm.ms.attr.Width) + int(xdiff))
		}
		if int(h)+int(ydiff) > 1 {
			h = uint(int(wm.ms.attr.Height) + int(ydiff))
		}
	}

	wm.display.MoveResizeWindow(
		wm.ms.start.Subwindow,
		x,
		y,
		w,
		h,
	)
}

func (wm *WindowManager) handleButtonRelease(ev xlib.Event) {
	br := ev.AsButtonEvent()
	log.Printf("ButtonRelease button=%d subwindow=%d xroot=%d yroot=%d\n",
		br.Button, br.Subwindow, br.XRoot, br.YRoot)

	wm.display.UngrabPointer(xlib.CurrentTime)
	wm.ms.start.Subwindow = xlib.NoneWindow()
}
