package sdk

import (
	"github.com/vapor-ware/synse-server-grpc/go"
)

// HealthManager is a global manager for health checks.
var HealthManager = healthManager{}

// HealthStatus describes the state of a health check. This type is a wrapper around
// the Synse GRPC PluginHealth_Status to make it easier for authors of health checks
// to consume.
type HealthStatus synse.PluginHealth_Status

const (
	HealthUnknown  = HealthStatus(synse.PluginHealth_UNKNOWN)
	HealthOk       = HealthStatus(synse.PluginHealth_OK)
	HealthDegraded = HealthStatus(synse.PluginHealth_PARTIALLY_DEGRADED)
	HealthFailing  = HealthStatus(synse.PluginHealth_FAILING)
)

type HealthCheckType string

const (
	PeriodicCheck HealthCheckType = "periodic"
)

type HealthChecker func() (HealthStatus, string)

type healthManager struct {
	checks []*HealthCheck
}

func (manager *healthManager) AddCheck(check *HealthCheck) {
	manager.checks = append(manager.checks, check)
}

func (manager *healthManager) RunChecks() {
	// FIXME: for now, this will only run checks once. We need to switch on the HealthCheckType
	// of each check and run it appropriately. e.g. "OneShot" would just run once. "Periodic" would
	// run on some period (tbd how to define the period), ...
	for _, check := range manager.checks {
		check.Check()
	}
}

/*
Would be cool to have different types of health checks.. not sure what that necessarily
means right now, but something to think about.

e.g.:
- event-driven health checks
- time-based health checks
- error-based health checks
- others?

For now in this first implementation, we will only support periodic time-based health
checks, though it shouldn't be too hard to extend in the future.
*/

type HealthCheck struct {
	Name      string
	Type      HealthCheckType
	Timestamp string
	Message   string
	Status    synse.PluginHealth_Status

	checkFn HealthChecker
}

func (healthCheck *HealthCheck) encode() *synse.HealthCheck {
	return &synse.HealthCheck{
		Name:      healthCheck.Name,
		Status:    healthCheck.Status,
		Message:   healthCheck.Message,
		Timestamp: healthCheck.Timestamp,
		Type:      string(healthCheck.Type),
	}
}

func (healthCheck *HealthCheck) Check() {
	status, message := healthCheck.checkFn()
	healthCheck.Timestamp = GetCurrentTime()
	healthCheck.Status = synse.PluginHealth_Status(status)
	healthCheck.Message = message
}

// NewHealthCheck creates a new instance of a HealthCheck.
func NewHealthCheck(name string, checkType HealthCheckType, checker HealthChecker) *HealthCheck {
	return &HealthCheck{
		Name:    name,
		Type:    checkType,
		checkFn: checker,
	}
}

// checkDataManagerReadBufferHealth is a plugin health check that periodically
// looks at the data manager's read buffer and determines whether the buffer is
// becoming full or not. Below is a summary of the conditions that will result
// in each state.
//
// UNKNOWN
//
// This health check should never result in an UNKNOWN state.
//
// OK
//
// This health check will result in an OK state when less than 90% of the
// buffer's capacity is being used.
//
// DEGRADED
//
// This health check will result in a DEGRADED state when less than 100% of the
// buffer's capacity is being used, but >= 90% of the capacity is used.
//
// FAILING
//
// This health check will result in a FAILING state when 100% of the buffer's
// capacity is being used.
func checkDataManagerReadBufferHealth() HealthStatus {
	// current number of elements queued in the channel
	channelSize := len(DataManager.readChannel)

	// total capacity of the channel
	channelCap := cap(DataManager.readChannel)

	pctUsage := float64(channelSize) / float64(channelCap)
	if pctUsage < 90 {
		return HealthOk
	} else if pctUsage < 100 {
		return HealthDegraded
	} else {
		return HealthFailing
	}
}

// checkDataManagerWriteBufferHealth is a plugin health check that periodically
// looks at the data manager's write buffer and determines whether the buffer is
// becoming full or not. Below is a summary of the conditions that will result
// in each state.
//
// UNKNOWN
//
// This health check should never result in an UNKNOWN state.
//
// OK
//
// This health check will result in an OK state when less than 90% of the
// buffer's capacity is being used.
//
// DEGRADED
//
// This health check will result in a DEGRADED state when less than 100% of the
// buffer's capacity is being used, but >= 90% of the capacity is used.
//
// FAILING
//
// This health check will result in a FAILING state when 100% of the buffer's
// capacity is being used.
func checkDataManagerWriteBufferHealth() HealthStatus {
	// current number of elements queued in the channel
	channelSize := len(DataManager.writeChannel)

	// total capacity of the channel
	channelCap := cap(DataManager.writeChannel)

	pctUsage := float64(channelSize) / float64(channelCap)
	if pctUsage < 90 {
		return HealthOk
	} else if pctUsage < 100 {
		return HealthDegraded
	} else {
		return HealthFailing
	}
}
