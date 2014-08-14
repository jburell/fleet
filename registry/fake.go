package registry

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/coreos/fleet/Godeps/_workspace/src/github.com/coreos/go-semver/semver"

	"github.com/coreos/fleet/job"
	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/unit"
)

func NewFakeRegistry() *FakeRegistry {
	return &FakeRegistry{
		machines:  []machine.MachineState{},
		jobStates: map[string]*unit.UnitState{},
		jobs:      map[string]job.Job{},
		units:     []unit.UnitFile{},
		version:   nil,
	}
}

type FakeRegistry struct {
	// Not all methods of required by the Registry interface are implemented
	// by the TestRegistry. Any calls to these unimplemented methods will
	// result in a panic.
	Registry
	sync.RWMutex

	machines  []machine.MachineState
	jobStates map[string]*unit.UnitState
	jobs      map[string]job.Job
	units     []unit.UnitFile
	version   *semver.Version
}

func (f *FakeRegistry) SetMachines(machines []machine.MachineState) {
	f.Lock()
	defer f.Unlock()

	f.machines = machines
}

func (f *FakeRegistry) SetJobs(jobs []job.Job) {
	f.Lock()
	defer f.Unlock()

	f.jobs = make(map[string]job.Job, len(jobs))
	for _, j := range jobs {
		f.jobs[j.Name] = j
	}
}

func (f *FakeRegistry) SetUnitStates(jobStates map[string]*unit.UnitState) {
	f.Lock()
	defer f.Unlock()

	f.jobStates = jobStates
}

func (f *FakeRegistry) SetUnits(units []unit.UnitFile) {
	f.Lock()
	defer f.Unlock()

	f.units = units
}

func (f *FakeRegistry) SetLatestVersion(v semver.Version) {
	f.Lock()
	defer f.Unlock()

	f.version = &v
}

func (f *FakeRegistry) Machines() ([]machine.MachineState, error) {
	f.RLock()
	defer f.RUnlock()

	return f.machines, nil
}

func (f *FakeRegistry) Units() ([]job.Unit, error) {
	f.RLock()
	defer f.RUnlock()

	var sorted sort.StringSlice
	for _, j := range f.jobs {
		sorted = append(sorted, j.Name)
	}
	sorted.Sort()

	units := make([]job.Unit, 0, len(f.jobs))
	for _, jName := range sorted {
		j := f.jobs[jName]
		u := job.Unit{
			Name:        j.Name,
			Unit:        j.Unit,
			TargetState: j.TargetState,
		}
		units = append(units, u)
	}

	return units, nil
}

func (f *FakeRegistry) Schedule() ([]job.ScheduledUnit, error) {
	f.RLock()
	defer f.RUnlock()

	var sorted sort.StringSlice
	for _, j := range f.jobs {
		sorted = append(sorted, j.Name)
	}
	sorted.Sort()

	sUnits := make([]job.ScheduledUnit, 0, len(f.jobs))
	for _, jName := range sorted {
		j := f.jobs[jName]
		su := job.ScheduledUnit{
			Name:            j.Name,
			State:           j.State,
			TargetMachineID: j.TargetMachineID,
		}
		sUnits = append(sUnits, su)
	}

	return sUnits, nil
}

func (f *FakeRegistry) Unit(name string) (*job.Unit, error) {
	f.RLock()
	defer f.RUnlock()

	j, ok := f.jobs[name]
	if !ok {
		return nil, nil
	}

	u := job.Unit{
		Name:        j.Name,
		Unit:        j.Unit,
		TargetState: j.TargetState,
	}
	return &u, nil
}

func (f *FakeRegistry) ScheduledUnit(name string) (*job.ScheduledUnit, error) {
	f.RLock()
	defer f.RUnlock()

	j, ok := f.jobs[name]
	if !ok {
		return nil, nil
	}

	j.UnitState = f.jobStates[name]
	su := job.ScheduledUnit{
		Name:            j.Name,
		State:           j.State,
		TargetMachineID: j.TargetMachineID,
	}
	return &su, nil
}

func (f *FakeRegistry) CreateUnit(u *job.Unit) error {
	f.Lock()
	defer f.Unlock()

	_, ok := f.jobs[u.Name]
	if ok {
		return errors.New("unit already exists")
	}

	j := job.Job{
		Name: u.Name,
		Unit: u.Unit,
	}

	f.jobs[u.Name] = j
	return nil
}

func (f *FakeRegistry) DestroyJob(name string) error {
	f.Lock()
	defer f.Unlock()

	delete(f.jobs, name)
	return nil
}

func (f *FakeRegistry) SetJobTargetState(name string, target job.JobState) error {
	f.Lock()
	defer f.Unlock()

	j, ok := f.jobs[name]

	if !ok {
		return errors.New("job does not exist")
	}

	j.TargetState = target
	f.jobs[name] = j

	return nil
}

func (f *FakeRegistry) ScheduleJob(name string, machID string) error {
	f.Lock()
	defer f.Unlock()

	j, ok := f.jobs[name]

	if !ok {
		return errors.New("job does not exist")
	}

	j.TargetMachineID = machID
	f.jobs[name] = j

	return nil
}

func (f *FakeRegistry) SaveUnitState(jobName string, unitState *unit.UnitState) {
	f.Lock()
	defer f.Unlock()

	f.jobStates[jobName] = unitState
}

func (f *FakeRegistry) RemoveUnitState(jobName string) error {
	delete(f.jobStates, jobName)
	return nil
}

func (f *FakeRegistry) LatestVersion() (*semver.Version, error) {
	f.RLock()
	defer f.RUnlock()

	return f.version, nil
}

func (f *FakeRegistry) JobHeartbeat(jobName, agentMachID string, ttl time.Duration) error {
	return nil
}

func (f *FakeRegistry) ClearJobHeartbeat(string) {}
