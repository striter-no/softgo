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
)

type WindowKeyboard interface {
	PollEvents()
	IsKeyPressed(key Key) bool
	Close()
}
