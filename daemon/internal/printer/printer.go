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

	// Number is the print system's own job counter, which rises with every
	// job the server accepts. The manager keeps the highest one it has seen
	// so it can recognise a finished job that was never in the queue when it
	// looked.
	//
	// It stays out of the API: a simulated job has no number and leaves this
	// zero, which the manager reads as "no counter to compare".
	Number int `json:"-"`
}

type Snapshot struct {
	State State     `json:"state"`
	Job   *PrintJob `json:"job,omitempty"`
}
