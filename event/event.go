package event

type Status string

func (s Status) Error() string {
	return string(s)
}

const (
/*
	not-found
	active
	loaded
	inactive
	waiting
	running
	exited
	dead
*/
//TODO: Socket activeation??
//listening

//FUTURE:
//elapsed
//mounted
//plugged
//stub
)

const (
	//General Errors.
	ForbiddenOperation Status = "Forbidden Operation."

	//Process Status
	ProcessConnectFailed Status = "Failed to connect IO."
	ProcessStartFailed   Status = "Failed to start process."
	ProcessRunning       Status = "Process Running."

	//Unit Status
	UnitStarting        Status = "Starting Unit."
	UnitLoading         Status = "Loading Unit."
	UnitActive          Status = "Unit Active."
	UnitDead            Status = "Unit Dead."
	UnitLoadingFailed   Status = "Unit Load failed."
	UnitNotLoaded       Status = "Unit Not Loaded."
	UnitAlreadyLoaded   Status = "Unit Already loaded."
	UnitPreparingFailed Status = "Unit Failed to prepare."
	UnitRegistering     Status = "Registering Unit."
	UnitDeregistered    Status = "Dergistered Unit."

	StatusExitSuccess Status = "Exit."
	StatusExitCrash   Status = "Exit."
	StatusLoaded      Status = "Loaded."
)

func New(from string, status Status, details string) Event {
	return Event{from, status, details}
}

type Event struct {
	From    string
	Status  Status
	Details string
}

func (e *Event) String() string {
	return string(e.Status)
}
