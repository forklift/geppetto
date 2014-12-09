package event

type Type string

func (s Type) Error() string {
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
	ForbiddenOperation Type = "Forbidden Operation."

	//Process Type
	ProcessConnectFailed Type = "Failed to connect IO."
	ProcessStartFailed   Type = "Failed to start process."
	ProcessRunning       Type = "Process Running."
	ProcessExited        Type = "Process Exited."

	//Unit Type
	UnitStarting        Type = "Starting Unit."
	UnitLoading         Type = "Loading Unit."
	UnitActive          Type = "Unit Active."
	UnitDead            Type = "Unit Dead."
	UnitLoadingFailed   Type = "Unit Load failed."
	UnitNotLoaded       Type = "Unit Not Loaded."
	UnitAlreadyLoaded   Type = "Unit Already loaded."
	UnitPreparingFailed Type = "Unit Failed to prepare."
	UnitRegistering     Type = "Registering Unit."
	UnitDeregistered    Type = "Dergistered Unit."
	UnitStopping        Type = "Unit Recieved requested to Stop."
	UnitStopWatch       Type = "Process Monitoring stopped."

	StatusExitSuccess Type = "Exit."
	StatusExitCrash   Type = "Exit."
	StatusType             = "Loaded."
)

func New(from string, status Type, details interface{}) Event {

	d := Details{}
	if e, ok := details.(error); ok {
		d.err = e
	} else if s, ok := details.(string); ok {
		d.str = s
	}
	//TODO: Convert int to string and et al.

	return Event{from, status, d}
}

type Details struct {
	err error
	str string
}

type Event struct {
	From    string
	Type    Type
	Payload Details
}

func (e *Event) String() string {
	return string(e.Payload.str)
}

func (e *Event) Error() error {
	return e.Payload.err
}

func (e *Event) Details() string {
	if e.Payload.err != nil {
		return e.Payload.err.Error()
	}
	return e.Payload.str
}
