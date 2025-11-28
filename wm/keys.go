package wm

import (
	"log"

	"github.com/francescorocca/go-xlib"
)

func grabKeyWithMasks(display *xlib.Display, keycode uint, mods uint) {
	masks := []uint{
		0,
		xlib.Mod2Mask,                 // NumLock
		xlib.LockMask,                 // CapsLock
		xlib.Mod2Mask | xlib.LockMask, // NumLock + CapsLock
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
	// Grab WIN + Enter
	grabKeyWithMasks(wm.display, wm.retkeycode, xlib.Mod4Mask)

	// Grab mouse buttons (1 = move, 3 = resize)
	if err := wm.display.GrabButton(
		1,
		xlib.Mod4Mask,
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
		3,
		xlib.Mod4Mask,
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
	log.Printf("\t(ret=%d, Mod4Mask=0x%x)\n", wm.retkeycode, xlib.Mod4Mask)

	// WIN + Enter -> launcher
	if ke.Keycode == wm.retkeycode && (ke.State&xlib.Mod4Mask) != 0 {
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
