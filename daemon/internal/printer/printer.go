package printer

type State string

const (
	StateIdle      State = "idle"
	StatePrinting  State = "printing"
	StateCompleted State = "completed"
	ErrorState     State = "error"
)

type PrintJob struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Snapshot struct {
	State State     `json:"state"`
	Job   *PrintJob `json:"job,omitempty"`
}
