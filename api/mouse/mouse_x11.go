//go:build linux

package mouse

/*
#cgo LDFLAGS: -lX11 -lXfixes
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/extensions/Xfixes.h>
#include <stdlib.h>

static int X11ErrorHandler(Display *d, XErrorEvent *e) {
    return 0;
}

void setup_x11_error_handler() {
    XSetErrorHandler(X11ErrorHandler);
}

Window get_active_window(Display *display) {
    Atom net_active = XInternAtom(display, "_NET_ACTIVE_WINDOW", False);
    Atom type;
    int format;
    unsigned long nitems, bytes_after;
    unsigned char *data = NULL;
    Window win = 0;

    if (XGetWindowProperty(display, DefaultRootWindow(display), net_active, 0, ~0, False,
                           XA_WINDOW, &type, &format, &nitems, &bytes_after, &data) == Success) {
        if (type == XA_WINDOW && nitems > 0 && data != NULL) {
            win = *((Window*)data);
            XFree(data);
        }
    }

    if (win == 0) {
        int revert_to;
        XGetInputFocus(display, &win, &revert_to);
    }
    return win;
}
*/
import "C"
import (
	"errors"
)

type linuxMouse struct {
	display *C.Display
	window  C.Window

	posX, posY    int
	currentButton MouseButton
}

func NewWindowMouse() (WindowMouse, error) {
	dpy := C.XOpenDisplay(nil)
	if dpy == nil {
		return nil, errors.New("cannot open X display")
	}

	win := C.get_active_window(dpy)

	C.setup_x11_error_handler()

	return &linuxMouse{
		display: dpy,
		window:  win,
	}, nil
}

func (m *linuxMouse) Close() {
	if m.display != nil {
		C.XCloseDisplay(m.display)
		m.display = nil
	}
}

func (m *linuxMouse) PollEvents() {
	var root, child C.Window
	var rootX, rootY, winX, winY C.int
	var mask C.uint

	C.XQueryPointer(m.display, m.window, &root, &child, &rootX, &rootY, &winX, &winY, &mask)

	m.posX = int(winX)
	m.posY = int(winY)

	switch {
	case mask&C.Button1Mask != 0:
		m.currentButton = Left
	case mask&C.Button3Mask != 0:
		m.currentButton = Right
	case mask&C.Button2Mask != 0:
		m.currentButton = Middle
	case mask&C.Button4Mask != 0:
		m.currentButton = ScrollUp
	case mask&C.Button5Mask != 0:
		m.currentButton = ScrollDown
	default:
		m.currentButton = Unknown
	}
}

func (m *linuxMouse) IsButtonPressed(button MouseButton) bool {
	return m.currentButton == button
}

func (m *linuxMouse) MoveMouse(x, y int) {
	C.XWarpPointer(m.display, C.None, m.window, 0, 0, 0, 0, C.int(x), C.int(y))
	C.XFlush(m.display)
	m.posX = x
	m.posY = y
}

func (m *linuxMouse) GetPosition() (int, int) {
	return m.posX, m.posY
}

func (m *linuxMouse) GetButtonPressed() MouseButton {
	return m.currentButton
}

func (m *linuxMouse) LockCursor() {
	C.XRaiseWindow(m.display, m.window)
	C.XSetInputFocus(m.display, m.window, C.RevertToParent, C.CurrentTime)
	C.XFlush(m.display)

	mask := C.long(C.ButtonPressMask | C.ButtonReleaseMask | C.PointerMotionMask)
	C.XGrabPointer(m.display, m.window, C.True, C.uint(mask),
		C.GrabModeAsync, C.GrabModeAsync, m.window, C.None, C.CurrentTime)
	C.XFlush(m.display)
}

func (m *linuxMouse) UnlockCursor() {
	C.XUngrabPointer(m.display, C.CurrentTime)
	C.XFlush(m.display)
}

func (m *linuxMouse) HideMouse() {
	C.XFixesHideCursor(m.display, m.window)
	C.XFlush(m.display)
}

func (m *linuxMouse) ShowMouse() {
	C.XFixesShowCursor(m.display, m.window)
	C.XFlush(m.display)
}

func (m *linuxMouse) GetSize() (int, int) {
	if m.display == nil {
		return 0, 0
	}
	var attrs C.XWindowAttributes
	if C.XGetWindowAttributes(m.display, m.window, &attrs) == 0 {
		return 0, 0
	}
	return int(attrs.width), int(attrs.height)
}
