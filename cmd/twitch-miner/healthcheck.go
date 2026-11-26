package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/heartbeat"
)

// healthyWithin is how old the last watch round may be. A round takes about
// a minute, or twenty seconds when nobody is live, so five minutes without
// one means the miner is stuck.
const healthyWithin = 5 * time.Minute

// healthStatus says whether a miner whose last beat has this age is working.
func healthStatus(age time.Duration, err error) (message string, healthy bool) {
	switch {
	case os.IsNotExist(err):
		return "unhealthy: the miner has not started watching yet", false
	case err != nil:
		return "unhealthy: " + err.Error(), false
	case age > healthyWithin:
		return fmt.Sprintf("unhealthy: no watch round for %s", age.Truncate(time.Second)), false
	default:
		return fmt.Sprintf("healthy: last watch round %s ago", age.Truncate(time.Second)), true
	}
}

// runHealthcheck reports on the miner running in the current data folder and
// ends the program: status 0 when it is working, 1 when it is not.
func runHealthcheck() {
	age, err := heartbeat.Age(heartbeat.File, time.Now())
	message, healthy := healthStatus(age, err)
	fmt.Println(message)
	if !healthy {
		os.Exit(1)
	}
	os.Exit(0)
}
