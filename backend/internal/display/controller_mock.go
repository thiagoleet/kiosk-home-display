package display

import "log"

type MockController struct{}

func (c *MockController) Wake() error {
	log.Println("Display: Wake called")
	return nil
}

func (c *MockController) Sleep() error {
	log.Println("Display: Sleep called")
	return nil
}

// func (c *MockController) Dim() error {
// 	log.Println("Display: Dim called")
// 	return nil
// }

func (c *MockController) SetBrightness(level int) error {
	log.Printf("Display: SetBrightness called with level %d", level)
	return nil
}
