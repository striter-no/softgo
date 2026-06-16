package keyboard

type Key int

const (
	Unknown Key = iota

	KeyW
	KeyA
	KeyS
	KeyD
	KeyE
	KeyQ

	KeySpace
	KeyEsc
	KeyShift
	KeyCtrl

	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	KeyUp
	KeyDown
	KeyLeft
	KeyRight

	KeyB
	KeyC
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyR
	KeyT
	KeyU
	KeyV
	KeyX
	KeyY
	KeyZ

	Key0
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9

	KeyAlt
	KeyTab
	KeyEnter
	KeyBackspace
	KeyInsert
	KeyDelete
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown

	keyDynamicStart Key = 10000
)

type WindowKeyboard interface {
	PollEvents()
	IsKeyPressed(key Key) bool

	RegisterKey(name string) Key

	IsKeyPressedByName(name string) bool

	Close()
}
