package unit

import (
	"errors"
	"fmt"
	"sync"

	"github.com/forklift/geppetto/event"
)

var BasePath = "/etc/geppetto/services"

type Unit struct {
	//Unit is the "unit" files for Geppeto.
	Explicit bool

	// Reserved names `gGod`, `gCurrentTransaction`
	Name string

	// A space-separated list of required prerequiestp. if the units listed here are not started already, they will not be started and the transaction will fail immediately.
	Prerequests []string

	// Similar to Prerequests, but the opposite. if units listed here are already started, this transaction will fail immediately.
	Conflicts []string

	// A space-separated list of required units to start with this unit, in any order. if any of these units fails, this unit will be cancneld.
	Requires []string

	// A space-separated list of required units to start with this unit, in any order. if any of these units fails, this unit will be NOT cancled if any of these units
	// fail after a successful startup.
	Wants []string

	//TODO: Validate if Before and After values exists in Requires or Wantp.

	// A space-separated list of required units to start before this unit is started. The units must exist in Requires or Wantp.
	Before []string
	// A space-separated list of required units to start before this unit is started. The units must exist in Requires or Wantp.
	After []string

	//Similar in to Preequirep. However in addition to this behavior, if any of the units listed suddenly disappears or fails, this unit stopp.
	BindsTo []string

	status event.Type

	Meta    Meta
	Service Service

	//internals
	prepared bool

	//Transaction
	Events    chan event.Event
	Listeners *event.Pipe

	Deps *UnitList

	//Keeping it safe.
	lock sync.Mutex
}

func (u *Unit) Prepare() error {

	if u.prepared {
		return nil
	}

	u.Listeners = event.NewPipe()
	u.Events = make(chan event.Event)

	var err error

	err = u.Meta.Prepare()
	if err != nil {
		return fmt.Errorf("Meta: %s", err)
	}

	err = u.Service.Prepare()
	if err != nil {
		return err
	}

	u.prepared = true
	return nil
}

func (u *Unit) Clean() []error {
	return u.Service.CloseIO() //TODO: Check if we have other deps.
}

func (u *Unit) Type() event.Type {
	return u.status

}

func (u *Unit) Start() chan event.Event {

	u.lock.Lock()
	events := make(chan event.Event)

	go func() {
		defer u.lock.Unlock()
		defer close(events)

		/*pipeline, errs, cancel, units := NewPipeline()

		started := pipeline.Do(errs, cancel, units, pipeline.StartUnit)

		go func() {
			u.Deps.ForEach(func(u *Unit) {
				units <- u
			})
			close(units)
		}()
		//pipeline.
		err := pipeline.Wait(errs, cancel, started)

		if err != nil {
			return err
		}*/

		var err error
		err = u.Service.ConnectIO()
		if err != nil {
			events <- event.New(u.Name, event.ServiceConnectFailed, err.Error())
			return
		}

		//err = u.process.Start()
		//go u.processWatch()

		if err != nil {
			events <- event.New(u.Name, event.ServiceStartFailed, err.Error())
			return
		}

	}()

	return events
}

func (u *Unit) Stop(sender string) chan event.Event {
	out := make(chan event.Event)

	go func() {
		u.stop(out, sender)
		close(out)
	}()

	return out
}

func (u *Unit) stop(out chan event.Event, sender string) error {

	u.Listeners.Add("transaction-"+sender, out)
	defer u.Listeners.Drop("transaction-" + sender)

	out <- event.New(u.Name, event.UnitStopping, sender)

	//TODO: Check Deps.
	//If multiple. Drop the sender from Listners and return nil.

	pipeline, errs, cancel, units := NewPipeline()

	stopped := pipeline.Do(errs, cancel, units, func(u *Unit) error {
		return u.stop(out, u.Name)
	})

	go u.Deps.ForEach(func(u *Unit) {
		units <- u
	}).Then(func() { close(units) })

	err := pipeline.Wait(errs, cancel, stopped)

	if err != nil {
		return err
	}

	//u.process.Service.Signal(syscall.SIGKILL)

	//Wait for process to exit.
	//TODO: Timeout!
	exitStatus := errors.New("") // <-u.processExitStatus

	return exitStatus

}

//func (u *Unit) processWatch() {
//	u.processExitStatus <- u.process.Wait()
//	close(u.processExitStatus)
//}

func (u *Unit) unitWatch() {

	for {
		select {
		case e := <-u.Events:
			switch e.Type {
			}
			//case exitStatus := <-u.processExitStatus:
			//	_ = exitStatus
			//		return
		}
	}
}
