// Use prysm version
package timeutil

import (
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"time"
)

// SlotStartTime returns the start time in terms of its unix epoch
// value.
func SlotStartTime(genesis, slot uint64) time.Time {
	duration := time.Second * time.Duration(slot*config.GenYangNetConfig().SecondsPerSlot)
	startTime := time.Unix(int64(genesis), 0).Add(duration)
	return startTime
}

// SlotsSinceGenesis returns the number of slots since
// the provided genesis time.
func SlotsSinceGenesis(genesis time.Time) uint64 {
	if genesis.After(Now()) { // Genesis has not occurred yet.
		return 0
	}
	return uint64(Since(genesis).Seconds()) / config.GenYangNetConfig().SecondsPerSlot
}

// EpochsSinceGenesis returns the number of slots since
// the provided genesis time.
func EpochsSinceGenesis(genesis time.Time) uint64 {
	return SlotsSinceGenesis(genesis) / config.GenYangNetConfig().SecondsPerSlot
}

// DivideSlotBy divides the SECONDS_PER_SLOT configuration
// parameter by a specified number. It returns a value of time.Duration
// in milliseconds, useful for dividing values such as 1 second into
// millisecond-based durations.
func DivideSlotBy(timesPerSlot int64) time.Duration {
	return time.Duration(int64(config.GenYangNetConfig().SecondsPerSlot*1000)/timesPerSlot) * time.Millisecond
}

// return the start time of the next epoch
func NextEpochTime(genesis time.Time, slot uint64) time.Time {
	cfg := config.GenYangNetConfig()
	epoch := slot / cfg.SlotsPerEpoch
	duration := time.Duration((epoch + 1) * cfg.SlotsPerEpoch * cfg.SecondsPerSlot)
	return genesis.Add(duration)
}

// return the start time of the next slot
func NextSlotTime(genesis time.Time, slot uint64) time.Time {
	cfg := config.GenYangNetConfig()
	duration := time.Duration((slot + 1) * cfg.SecondsPerSlot)
	return genesis.Add(duration)
}