package models

import (
	"fmt"
	"time"

	"github.com/tidwall/gjson"
	null "gopkg.in/guregu/null.v3"
)

// JobRun tracks the status of a job by holding its TaskRuns and the
// Result of each Run.
type JobRun struct {
	ID        string    `json:"id" storm:"id,index,unique"`
	JobID     string    `json:"jobId" storm:"index"`
	Status    string    `json:"status" storm:"index"`
	CreatedAt time.Time `json:"createdAt" storm:"index"`
	Result    RunResult `json:"result" storm:"inline"`
	TaskRuns  []TaskRun `json:"taskRuns" storm:"inline"`
}

// ForLogger formats the JobRun for a common formatting in the log.
func (jr JobRun) ForLogger(kvs ...interface{}) []interface{} {
	output := []interface{}{
		"job", jr.JobID,
		"run", jr.ID,
		"status", jr.Status,
	}

	if jr.Result.HasError() {
		output = append(output, "error", jr.Result.Error())
	}

	return append(kvs, output...)
}

// UnfinishedTaskRuns returns a list of TaskRuns for a JobRun
// which are not Completed or Errored.
func (jr JobRun) UnfinishedTaskRuns() []TaskRun {
	unfinished := jr.TaskRuns
	for _, tr := range jr.TaskRuns {
		if tr.Completed() {
			unfinished = unfinished[1:]
		} else if tr.Errored() {
			return []TaskRun{}
		} else {
			return unfinished
		}
	}
	return unfinished
}

// NextTaskRun returns the next immediate TaskRun in the list
// of unfinished TaskRuns.
func (jr JobRun) NextTaskRun() TaskRun {
	return jr.UnfinishedTaskRuns()[0]
}

// TaskRun stores the Task and represents the status of the
// Task to be ran.
type TaskRun struct {
	Task   Task      `json:"task"`
	ID     string    `json:"id" storm:"id,index,unique"`
	Status string    `json:"status"`
	Result RunResult `json:"result"`
}

// Completed returns true if the TaskRun status is StatusCompleted.
func (tr TaskRun) Completed() bool {
	return tr.Status == StatusCompleted
}

// Errored returns true if the TaskRun status is StatusErrored.
func (tr TaskRun) Errored() bool {
	return tr.Status == StatusErrored
}

// String returns info on the TaskRun as "ID,Type,Status,Result".
func (tr TaskRun) String() string {
	return fmt.Sprintf("TaskRun(%v,%v,%v,%v)", tr.ID, tr.Task.Type, tr.Status, tr.Result)
}

// ForLogger formats the TaskRun info for a common formatting in the log.
func (tr TaskRun) ForLogger(kvs ...interface{}) []interface{} {
	output := []interface{}{
		"type", tr.Task.Type,
		"params", tr.Task.Params,
		"taskrun", tr.ID,
		"status", tr.Status,
	}

	if tr.Result.HasError() {
		output = append(output, "error", tr.Result.Error())
	}

	return append(kvs, output...)
}

// MergeTaskParams merges the existing parameters on a TaskRun with the given JSON.
func (tr TaskRun) MergeTaskParams(j JSON) (TaskRun, error) {
	merged, err := tr.Task.Params.Merge(j)
	if err != nil {
		return tr, fmt.Errorf("TaskRun#Merge merging outputs: %v", err.Error())
	}

	tr.Task.Params = merged
	return tr, nil
}

// RunResult keeps track of the outcome of a TaskRun. It stores
// the Output and ErrorMessage, if any of either, and contains
// a Pending field to track the status.
type RunResult struct {
	Output       JSON        `json:"output"`
	ErrorMessage null.String `json:"error"`
	Pending      bool        `json:"pending"`
}

// RunResultWithValue returns a new RunResult with the given string
// value as a JSON object.
func RunResultWithValue(val string) RunResult {
	output := JSON{}
	output, err := output.Add("value", val)
	if err != nil {
		return RunResultWithError(err)
	}

	return RunResult{Output: output}
}

// RunResultWithError returns a new RunResult with the given
// error message.
func RunResultWithError(err error) RunResult {
	return RunResult{
		ErrorMessage: null.StringFrom(err.Error()),
	}
}

// RunResultPending returns a new RunResult keeping the same
// Output and ErrorMessage as given but with a Pending status
// set to true.
func RunResultPending(input RunResult) RunResult {
	return RunResult{
		Output:       input.Output,
		ErrorMessage: input.ErrorMessage,
		Pending:      true,
	}
}

// Get searches for and returns the JSON at the given path.
func (rr RunResult) Get(path string) (gjson.Result, error) {
	return rr.Output.Get(path), nil
}

func (rr RunResult) value() (gjson.Result, error) {
	return rr.Get("value")
}

// Value returns the string value of the Output JSON field.
func (rr RunResult) Value() (string, error) {
	val, err := rr.value()
	if err != nil {
		return "", err
	}
	if val.Type != gjson.String {
		return "", fmt.Errorf("non string value")
	}
	return val.String(), nil
}

// HasError returns true if the ErrorMessage is present.
func (rr RunResult) HasError() bool {
	return rr.ErrorMessage.Valid
}

// Error returns the string value of the ErrorMessage field.
func (rr RunResult) Error() string {
	return rr.ErrorMessage.String
}

// SetError stores the given error in the ErrorMessage field.
func (rr RunResult) SetError(err error) {
	rr.ErrorMessage = null.StringFrom(err.Error())
}

// GetError returns the error of a RunResult if it is present.
func (rr RunResult) GetError() error {
	if rr.HasError() {
		return fmt.Errorf("Run Result: ", rr.Error())
	} else {
		return nil
	}
}

// MergeOutput merges the existing Output on a RunResult with the given JSON.
func (rr RunResult) MergeOutput(j JSON) (RunResult, error) {
	merged, err := rr.Output.Merge(j)
	if err != nil {
		return rr, fmt.Errorf("TaskRun#Merge merging outputs: %v", err.Error())
	}

	rr.Output = merged
	return rr, nil
}
