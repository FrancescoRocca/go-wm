package wm

import (
	"fmt"
	"log"
	"os"

	"github.com/francescorocca/go-xlib"
)

type WindowManager struct {
	root          xlib.Window
	width, height int
	display       *xlib.Display
	retkeycode    uint
}

func InitWM() (*WindowManager, error) {
	var display *xlib.Display
	var err error

	// Initialize display
	if display, err = xlib.OpenDisplay(""); err != nil {
		fmt.Fprintln(os.Stderr, "OpenDisplay:", err)
		return nil, err
	}

	defer func() {
		if err != nil && display != nil {
			display.Close()
		}
	}()

	// Initialize error handler
	display.InitErrorHandler()

	// Root window
	root := display.DefaultRootWindow()
	rwa, err := display.GetWindowAttributes(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to get root window attributes:", err)
		return nil, err
	}

	// Listen for SubstructureNotifyMask and SubstructureRedirectMask
	if err = display.SelectInput(root, xlib.SubstructureNotifyMask|xlib.SubstructureRedirectMask); err != nil {
		fmt.Fprintln(os.Stderr, "SelectInput:", err)
		return nil, err
	}

	// Get the Return keycode
	ret, err := display.KeysymToKeycode(xlib.StringToKeysym("Return"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "KeysymToKeycode(Return):", err)
	}

	wm := &WindowManager{
		display:    display,
		root:       root,
		retkeycode: ret,
		width:      rwa.Width,
		height:     rwa.Height,
	}

	if err = wm.setupGrabs(); err != nil {
		log.Println("setupGrabs error:", err)
		return nil, err
	}

	return wm, nil
}

func (wm *WindowManager) Close() {
	if wm.display != nil {
		wm.display.Close()
	}
}

func (wm *WindowManager) Run() {
	for {
		ev := wm.display.NextEvent()
		wm.handleEvent(ev)
	}
}
