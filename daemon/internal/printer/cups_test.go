package printer

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// Captured from "lpstat -W not-completed -o" on a CUPS host. The trailing
// columns are the owner, the byte size and the queue time.
const lpstatQueuedOutput = `Brother_DCP_1600_series-31 thiago          485376   Wed Sep  9 15:43:15 2026
Brother_DCP_1600_series_2-32 thiago           11264   Wed Sep  9 15:44:02 2026
`

// Captured from "lpq -a", whose columns are the rank, the owner, the job
// number, the document title and the size.
const lpqOutput = `Brother_DCP_1600_series is ready and printing
Rank    Owner   Job     File(s)                         Total Size
active  thiago  31      quarterly report.pdf            485376 bytes
1st     thiago  32      shopping-list.txt               11264 bytes
`

type recordedCommand struct {
	name string
	args []string
}

type commandStub struct {
	calls   []recordedCommand
	outputs map[string]string
	errs    map[string]error
}

func newCommandStub() *commandStub {
	return &commandStub{
		outputs: map[string]string{
			"lpstat -W not-completed -o": lpstatQueuedOutput,
			"lpq -a":                     lpqOutput,
		},
		errs: map[string]error{},
	}
}

func (s *commandStub) run(
	name string,
	args ...string,
) (string, error) {
	s.calls = append(s.calls, recordedCommand{
		name: name,
		args: args,
	})

	key := strings.Join(
		append([]string{name}, args...),
		" ",
	)

	return s.outputs[key], s.errs[key]
}

func newStubbedCUPSSource() (*CUPSSource, *commandStub) {
	stub := newCommandStub()

	source := NewCUPSSource()
	source.command = stub.run

	return source, stub
}

func TestCUPSSourceActiveJobs(t *testing.T) {
	source, stub := newStubbedCUPSSource()

	jobs, err := source.ActiveJobs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	if jobs[0].ID != "Brother_DCP_1600_series-31" {
		t.Errorf(
			"expected the lpstat identifier, got %q",
			jobs[0].ID,
		)
	}

	// The title carries a space, so it survives only if the name is taken as
	// everything between the job number and the size.
	if jobs[0].Name != "quarterly report.pdf" {
		t.Errorf(
			"expected the document title, got %q",
			jobs[0].Name,
		)
	}

	if jobs[1].Name != "shopping-list.txt" {
		t.Errorf(
			"expected the document title, got %q",
			jobs[1].Name,
		)
	}

	if stub.calls[0].name != "lpstat" {
		t.Errorf(
			"expected lpstat to be the first call, got %q",
			stub.calls[0].name,
		)
	}
}

func TestCUPSSourceReportsAnEmptyQueue(t *testing.T) {
	source, stub := newStubbedCUPSSource()

	stub.outputs["lpstat -W not-completed -o"] = ""

	jobs, err := source.ActiveJobs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 0 {
		t.Fatalf("expected no jobs, got %d", len(jobs))
	}

	// An empty queue needs no titles, so lpq is not worth running.
	for _, call := range stub.calls {
		if call.name == "lpq" {
			t.Error("expected lpq to be skipped for an empty queue")
		}
	}
}

// A job with no recoverable title keeps its identifier as a name, which is the
// only thing the frontend can show for it.
func TestCUPSSourceFallsBackToTheJobIdentifier(t *testing.T) {
	source, stub := newStubbedCUPSSource()

	stub.errs["lpq -a"] = errors.New("lpq: Unable to connect")

	jobs, err := source.ActiveJobs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	if jobs[0].Name != "Brother_DCP_1600_series-31" {
		t.Errorf(
			"expected the identifier as a name, got %q",
			jobs[0].Name,
		)
	}
}

func TestCUPSSourceFailsWhenTheQueueCannotBeRead(t *testing.T) {
	source, stub := newStubbedCUPSSource()

	stub.errs["lpstat -W not-completed -o"] = errors.New(
		"lpstat: Bad file descriptor",
	)

	if _, err := source.ActiveJobs(); err == nil {
		t.Fatal("expected an error when lpstat fails")
	}
}

func TestCUPSSourceNamesAMissingInstallation(t *testing.T) {
	source, stub := newStubbedCUPSSource()

	stub.errs["lpstat -W not-completed -o"] = exec.ErrNotFound

	_, err := source.ActiveJobs()
	if err == nil {
		t.Fatal("expected an error when lpstat is missing")
	}

	if !strings.Contains(err.Error(), "cups-client") {
		t.Errorf(
			"expected the error to name the missing package, got %q",
			err,
		)
	}
}

// lpstat prints warnings and lpq prints headers and a "no entries" line. None
// of them describe a job.
func TestParseIgnoresNonJobLines(t *testing.T) {
	jobs := parseQueuedJobs(
		"lpstat: Transport endpoint is not connected\n" +
			"Brother_DCP_1600_series-31 thiago 485376 Wed Sep  9 15:43:15 2026\n",
	)

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	titles := parseJobTitles(
		"no entries\n" +
			"Rank    Owner   Job     File(s)                         Total Size\n",
	)

	if len(titles) != 0 {
		t.Fatalf("expected no titles, got %d", len(titles))
	}
}
