package wm

import (
	"log"

	"github.com/francescorocca/go-xlib"
)

const (
	MouseBtnLeft   = 1
	MouseBtnMiddle = 2
	MouseBtnRight  = 3
	KeyWin         = xlib.Mod4Mask
	NumLock        = xlib.Mod2Mask
	CapsLock       = xlib.LockMask
)

func grabKeyWithMasks(display *xlib.Display, keycode uint, mods uint) {
	masks := []uint{
		0,
		NumLock,
		CapsLock,
		NumLock | CapsLock,
	}

	for _, m := range masks {
		if err := display.GrabKey(
			keycode,
			mods|m,
			display.DefaultRootWindow(),
			xlib.True,
			xlib.GrabModeAsync,
			xlib.GrabModeAsync,
		); err != nil {
			log.Printf("GrabKey failed (key=%d, mask=0x%x): %v\n", keycode, mods|m, err)
		}
	}
}

func (wm *WindowManager) setupGrabs() error {
	grabKeyWithMasks(wm.display, wm.retkeycode, KeyWin)

	if err := wm.display.GrabButton(
		MouseBtnLeft,
		KeyWin,
		wm.root,
		xlib.True,
		xlib.ButtonPressMask|xlib.ButtonReleaseMask|xlib.PointerMotionMask,
		xlib.GrabModeAsync,
		xlib.GrabModeAsync,
		xlib.NoneWindow(),
		xlib.NoneCursor(),
	); err != nil {
		log.Println("GrabButton(1) error:", err)
		return err
	}

	if err := wm.display.GrabButton(
		MouseBtnRight,
		KeyWin,
		wm.root,
		xlib.True,
		xlib.ButtonPressMask|xlib.ButtonReleaseMask|xlib.PointerMotionMask,
		xlib.GrabModeAsync,
		xlib.GrabModeAsync,
		xlib.NoneWindow(),
		xlib.NoneCursor(),
	); err != nil {
		log.Println("GrabButton(3) error:", err)
		return err
	}

	return nil
}

func (wm *WindowManager) handleKeyPress(ev xlib.Event) {
	ke := ev.AsKeyEvent()
	log.Printf("KeyPress keycode=%d state=0x%x\n", ke.Keycode, ke.State)
	log.Printf("\t(ret=%d, KeyWin=0x%x)\n", wm.retkeycode, xlib.Mod4Mask)

	if ke.Keycode == wm.retkeycode && (ke.State&KeyWin) != 0 {
		log.Println("WIN+Enter detected -> launcher")
		wm.runLauncher()
	}

	// Move to front the focused window
	if ke.Subwindow != xlib.NoneWindow() {
		if err := wm.display.RaiseWindow(ke.Subwindow); err != nil {
			log.Println("RaiseWindow error:", err)
		}
	}
}
