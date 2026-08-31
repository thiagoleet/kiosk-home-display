package display

type Controller interface {
	Wake() error
	Sleep() error
	SetBrightness(level int) error
}
