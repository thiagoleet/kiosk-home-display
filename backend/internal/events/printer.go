package events

type PrinterEvent struct {
	JobID string `json:"jobId"`
	Name  string `json:"name"`
}
