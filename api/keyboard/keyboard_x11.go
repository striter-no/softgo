//go:build linux

package keyboard

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

var keySymTable = map[Key]string{
	KeyW: "w", KeyA: "a", KeyS: "s", KeyD: "d", KeyE: "e", KeyQ: "q",

	KeySpace:     "space",
	KeyEsc:       "Escape",
	KeyShift:     "Shift_L",
	KeyCtrl:      "Control_L",
	KeyAlt:       "Alt_L",
	KeyTab:       "Tab",
	KeyEnter:     "Return",
	KeyBackspace: "BackSpace",
	KeyInsert:    "Insert",
	KeyDelete:    "Delete",
	KeyHome:      "Home",
	KeyEnd:       "End",
	KeyPageUp:    "Prior",
	KeyPageDown:  "Next",

	KeyF1: "F1", KeyF2: "F2", KeyF3: "F3", KeyF4: "F4",
	KeyF5: "F5", KeyF6: "F6", KeyF7: "F7", KeyF8: "F8",
	KeyF9: "F9", KeyF10: "F10", KeyF11: "F11", KeyF12: "F12",

	KeyUp: "Up", KeyDown: "Down", KeyLeft: "Left", KeyRight: "Right",

	KeyB: "b", KeyC: "c", KeyF: "f", KeyG: "g", KeyH: "h", KeyI: "i",
	KeyJ: "j", KeyK: "k", KeyL: "l", KeyM: "m", KeyN: "n", KeyO: "o",
	KeyP: "p", KeyR: "r", KeyT: "t", KeyU: "u", KeyV: "v", KeyX: "x",
	KeyY: "y", KeyZ: "z",

	Key0: "0", Key1: "1", Key2: "2", Key3: "3", Key4: "4",
	Key5: "5", Key6: "6", Key7: "7", Key8: "8", Key9: "9",
}

type linuxKeyboard struct {
	display *C.Display
	keymap  [32]byte
	keysMap map[Key]C.KeyCode

	names  map[string]Key
	nextID Key
}

func NewWindowKeyboard() (WindowKeyboard, error) {
	dpy := C.XOpenDisplay(nil)
	if dpy == nil {
		return nil, errors.New("cannot open X display for keyboard")
	}

	kbd := &linuxKeyboard{
		display: dpy,
		keysMap: make(map[Key]C.KeyCode),
		names:   make(map[string]Key),
		nextID:  keyDynamicStart,
	}

	// Авто-регистрация всех стандартных клавиш.
	for k, name := range keySymTable {
		kbd.registerByName(k, name)
	}

	return kbd, nil
}

func (k *linuxKeyboard) registerByName(id Key, name string) {
	cname := C.CString(name)
	sym := C.XStringToKeysym(cname)
	C.free(unsafe.Pointer(cname))

	if sym == 0 {
		return
	}
	k.keysMap[id] = C.XKeysymToKeycode(k.display, sym)
}

func (k *linuxKeyboard) RegisterKey(name string) Key {
	if id, ok := k.names[name]; ok {
		return id
	}
	id := k.nextID
	k.nextID++
	k.names[name] = id
	k.registerByName(id, name)
	return id
}

func (k *linuxKeyboard) IsKeyPressedByName(name string) bool {
	return k.IsKeyPressed(k.RegisterKey(name))
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
	keycode, ok := k.keysMap[key]
	if !ok || keycode == 0 {
		return false
	}

	byteIdx := keycode / 8
	bitIdx := keycode % 8

	return (k.keymap[byteIdx] & (1 << bitIdx)) != 0
}
