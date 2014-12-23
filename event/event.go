package event

import "strconv"

type Type string

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
	UnitActive          Type = "Unit Active."
	UnitDead            Type = "Unit Dead."
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
