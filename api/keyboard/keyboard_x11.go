//go:build linux

package keyboard

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
*/
import "C"
import (
	"errors"
)

type linuxKeyboard struct {
	display *C.Display
	keymap  [32]byte
	keysMap map[Key]C.KeyCode
}

func NewWindowKeyboard() (WindowKeyboard, error) {
	dpy := C.XOpenDisplay(nil)
	if dpy == nil {
		return nil, errors.New("cannot open X display for keyboard")
	}

	kbd := &linuxKeyboard{
		display: dpy,
		keysMap: make(map[Key]C.KeyCode),
	}

	kbd.keysMap[KeyW] = C.XKeysymToKeycode(dpy, C.XK_w)
	kbd.keysMap[KeyA] = C.XKeysymToKeycode(dpy, C.XK_a)
	kbd.keysMap[KeyS] = C.XKeysymToKeycode(dpy, C.XK_s)
	kbd.keysMap[KeyD] = C.XKeysymToKeycode(dpy, C.XK_d)
	kbd.keysMap[KeyE] = C.XKeysymToKeycode(dpy, C.XK_e)
	kbd.keysMap[KeyQ] = C.XKeysymToKeycode(dpy, C.XK_q)
	kbd.keysMap[KeySpace] = C.XKeysymToKeycode(dpy, C.XK_space)
	kbd.keysMap[KeyEsc] = C.XKeysymToKeycode(dpy, C.XK_Escape)
	kbd.keysMap[KeyShift] = C.XKeysymToKeycode(dpy, C.XK_Shift_L)
	kbd.keysMap[KeyCtrl] = C.XKeysymToKeycode(dpy, C.XK_Control_L)

	return kbd, nil
}

func (k *linuxKeyboard) Close() {
	if k.display != nil {
		C.XCloseDisplay(k.display)
		k.display = nil
	}
}

func (k *linuxKeyboard) PollEvents() {
	var keys [32]C.char
	C.XQueryKeymap(k.display, &keys[0])

	for i := range 32 {
		k.keymap[i] = byte(keys[i])
	}
}

func (k *linuxKeyboard) IsKeyPressed(key Key) bool {
	keycode := k.keysMap[key]
	if keycode == 0 {
		return false
	}

	byteIdx := keycode / 8
	bitIdx := keycode % 8

	return (k.keymap[byteIdx] & (1 << bitIdx)) != 0
}
