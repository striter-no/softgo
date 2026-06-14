package mouse

type MouseButton int

const (
	Unknown MouseButton = iota
	Left
	Middle
	Right
	ScrollUp
	ScrollDown
)

type WindowMouse interface {
	PollEvents()
	IsButtonPressed(button MouseButton) bool
	MoveMouse(x, y int)
	GetPosition() (x, y int)
	GetButtonPressed() MouseButton
	LockCursor()
	UnlockCursor()
	HideMouse()
	ShowMouse()
	GetSize() (width, height int)
	Close()
}
