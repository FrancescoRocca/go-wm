package wm

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/francescorocca/go-xlib/xlib"
)

type WindowManager struct {
	root       xlib.Window
	display    *xlib.Display
	retkeycode uint
}

func spawnLauncher() error {
	commands := [][]string{
		{"dmenu_run"},
		{"rofi", "-show", "run"},
		{"xterm"},
	}

	var lastErr error

	for _, cmd := range commands {
		if len(cmd) == 0 {
			continue
		}

		if _, err := exec.LookPath(cmd[0]); err != nil {
			lastErr = err
			log.Printf("spawnLauncher: %s not found: %v\n", cmd[0], err)
			continue
		}

		if err := exec.Command(cmd[0], cmd[1:]...).Start(); err == nil {
			return nil
		} else {
			lastErr = err
			log.Printf("spawnLauncher: error launching %v: %v\n", cmd, err)
		}
	}

	return lastErr
}

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

func InitWM() (*WindowManager, error) {
	var display *xlib.Display
	var err error

	/* Initialize display */
	if display, err = xlib.OpenDisplay(""); err != nil {
		fmt.Fprintln(os.Stderr, "OpenDisplay:", err)
		return nil, err
	}

	defer func() {
		if err != nil && display != nil {
			display.Close()
		}
	}()

	/* Initialize error handler */
	display.InitErrorHandler()

	root := display.DefaultRootWindow()

	/* Listen for SubstructureNotifyMask and SubstructureRedirectMask */
	if err = display.SelectInput(root, xlib.SubstructureNotifyMask|xlib.SubstructureRedirectMask); err != nil {
		fmt.Fprintln(os.Stderr, "SelectInput:", err)
		return nil, err
	}

	/* Get the Return keycode */
	var ret uint
	if ret, err = display.KeysymToKeycode(xlib.StringToKeysym("Return")); err != nil {
		fmt.Fprintln(os.Stderr, "KeysymToKeycode(Return):", err)
	}

	/* Grab WIN + Enter */
	grabKeyWithMasks(display, ret, xlib.Mod4Mask)

	/* Grab mouse buttons (1 = move, 3 = resize) */
	if err = display.GrabButton(
		1,
		xlib.Mod4Mask,
		root,
		xlib.True,
		xlib.ButtonPressMask|xlib.ButtonReleaseMask|xlib.PointerMotionMask,
		xlib.GrabModeAsync,
		xlib.GrabModeAsync,
		xlib.NoneWindow(),
		xlib.NoneCursor(),
	); err != nil {
		log.Println("GrabButton(1) error:", err)
	}

	if err = display.GrabButton(
		3,
		xlib.Mod4Mask,
		root,
		xlib.True,
		xlib.ButtonPressMask|xlib.ButtonReleaseMask|xlib.PointerMotionMask,
		xlib.GrabModeAsync,
		xlib.GrabModeAsync,
		xlib.NoneWindow(),
		xlib.NoneCursor(),
	); err != nil {
		log.Println("GrabButton(3) error:", err)
	}

	return &WindowManager{
		display:    display,
		root:       root,
		retkeycode: ret,
	}, nil
}

func (wm *WindowManager) Close() {
	if wm.display != nil {
		wm.display.Close()
	}
}

func (wm *WindowManager) Run() {
	var start xlib.ButtonEvent
	var attr xlib.WindowAttributes
	start.Subwindow = xlib.NoneWindow()
	var err error

	for {
		ev := wm.display.NextEvent()

		switch ev.Type() {

		case xlib.KeyPress:
			ke := ev.AsKeyEvent()
			log.Printf("KeyPress keycode=%d state=0x%x\n", ke.Keycode, ke.State)
			log.Printf("\t(ret=%d, Mod4Mask=0x%x)\n", wm.retkeycode, xlib.Mod4Mask)

			/* WIN + Enter -> launcher */
			if ke.Keycode == wm.retkeycode && (ke.State&xlib.Mod4Mask) != 0 {
				log.Println("WIN+Enter detected -> launcher")
				if err := spawnLauncher(); err != nil {
					fmt.Fprintln(os.Stderr, "Spawn error:", err)
				}
			}

			/* Move to front the focused window */
			if ke.Subwindow != xlib.NoneWindow() {
				if err := wm.display.RaiseWindow(ke.Subwindow); err != nil {
					log.Println("RaiseWindow error:", err)
				}
			}

		case xlib.ButtonPress:
			be := ev.AsButtonEvent()
			log.Printf("ButtonPress button=%d subwindow=%d xroot=%d yroot=%d state=0x%x\n",
				be.Button, be.Subwindow, be.XRoot, be.YRoot, be.State)

			if be.Subwindow != xlib.NoneWindow() {
				attr, err = wm.display.GetWindowAttributes(be.Subwindow)
				if err != nil {
					log.Println("GetWindowAttributes error:", err)
				}
				start = *be

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

		case xlib.MotionNotify:
			if start.Subwindow != xlib.NoneWindow() {
				me := ev.AsButtonEvent()
				xdiff := me.XRoot - start.XRoot
				ydiff := me.YRoot - start.YRoot

				x := attr.X
				y := attr.Y
				w := uint(attr.Width)
				h := uint(attr.Height)

				switch start.Button {
				case 1:
					/* WIN + left click = MOVE */
					x += xdiff
					y += ydiff

				case 3:
					/* WIN + right click = RESIZE */
					if int(w)+int(xdiff) > 1 {
						w = uint(int(attr.Width) + int(xdiff))
					}
					if int(h)+int(ydiff) > 1 {
						h = uint(int(attr.Height) + int(ydiff))
					}
				}

				wm.display.MoveResizeWindow(
					start.Subwindow,
					x,
					y,
					w,
					h,
				)
			}

		case xlib.ButtonRelease:
			br := ev.AsButtonEvent()
			log.Printf("ButtonRelease button=%d subwindow=%d xroot=%d yroot=%d\n",
				br.Button, br.Subwindow, br.XRoot, br.YRoot)

			wm.display.UngrabPointer(xlib.CurrentTime)
			start.Subwindow = xlib.NoneWindow()

		case xlib.MapRequest:
			mr := ev.AsMapRequestEvent()
			if err := wm.display.MapWindow(mr.Window); err != nil {
				log.Println("MapWindow error:", err)
			}

		case xlib.ConfigureRequest:
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
				changes.BorderWidth = cr.BorderWidth
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
				log.Println("ConfigureWindow error:", err)
			}
		}
	}
}
