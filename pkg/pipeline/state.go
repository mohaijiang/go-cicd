package pipeline

type State string

const (
	StateNotRun  = State("NotRun")
	StateRunning = State("Running")
	StateDone    = State("Done")
	StateFail    = State("Fail")
)

var (
	NotRunInstance  = &NotRunStep{}
	RunningInstance = &RunningStep{}
	DoneInstance    = &DoneStep{}
	FailInstance    = &FailStep{}
)

type IStepState interface {
	getState() State
	execute(machine *StepMachine)
}

// NotRunStep 未开始状态
type NotRunStep struct {
}

func (step *NotRunStep) getState() State {
	return StateNotRun
}

func (step *NotRunStep) execute(sm *StepMachine) {

	sm.stepState = RunningInstance
}

type ActionHandler interface {
	Pre() error
	Hook() error
	Post() error
}

// RunningStep 运行中状态
type RunningStep struct {
	handler ActionHandler
}

func NewRunningStep(handler ActionHandler) *RunningStep {
	return &RunningStep{
		handler: handler,
	}
}

func (step *RunningStep) getState() State {
	return StateRunning
}

func (step *RunningStep) execute(sm *StepMachine) {
	// will not happen
	err := step.handler.Hook()

	if err != nil {
		sm.stepState = FailInstance
	} else {
		sm.stepState = DoneInstance
	}

}

// DoneStep 完成状态
type DoneStep struct {
}

func (step *DoneStep) getState() State {
	return StateDone
}

func (step *DoneStep) execute(sm *StepMachine) {
	// will not happen
}

// FailStep 失败状态
type FailStep struct {
}

func (step *FailStep) getState() State {
	return StateFail
}

func (step *FailStep) execute(sm *StepMachine) {
	// will not happen
}
