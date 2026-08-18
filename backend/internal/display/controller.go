package display

type Controller interface {
	Wake() error
	Sleep() error
	Dim() error
}
