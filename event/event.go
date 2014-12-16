package event

import "strconv"

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

	//Service Type
	ServiceConnectFailed Type = "Failed to connect IO."
	ServiceStartFailed   Type = "Failed to start process."
	ServiceRunning       Type = "Process Running."
	ServiceExited        Type = "Process Exited."

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

	switch details.(type) {
	case error:
		d.err = details.(error)
	case string:
		d.str = details.(string)
	case uint, int, int32, uint32, uint64, int64:
		d.str = strconv.Itoa(details.(int))
	}

	return Event{from, status, d}
}

type Details struct {
	err error
	str string
}

type Event struct {
	From    string
	Type    Type
	payload Details
}

func (e *Event) String() string {
	return string(e.payload.str)
}

func (e *Event) Error() error {
	return e.payload.err
}
