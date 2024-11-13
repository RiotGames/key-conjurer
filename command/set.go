package command

import (
	"fmt"
	"strconv"
)

type SetCommand struct {
	TTL           TTLCommand           `name:"ttl" help:"Sets ttl value in number of hours."`
	TimeRemaining TimeRemainingCommand `name:"time-remaining" help:"Sets time remaining value in number of minutes."`
}

type TTLCommand struct {
	TTL string `arg:"" help:"The ttl value in number of hours." placeholder:"hours"`
}

func (t TTLCommand) Run(globals *Globals, config *Config) error {
	ttl, err := strconv.ParseUint(t.TTL, 10, 32)
	if err != nil {
		return fmt.Errorf("unable to parse value %s", t.TTL)
	}

	config.TTL = uint(ttl)
	return nil
}

type TimeRemainingCommand struct {
	TimeRemaining string `arg:"" help:"The time remaining value in number of minutes." placeholder:"minutes"`
}

func (t TimeRemainingCommand) Run(globals *Globals, config *Config) error {
	timeRemaining, err := strconv.ParseUint(t.TimeRemaining, 10, 32)
	if err != nil {
		return fmt.Errorf("unable to parse value %s", t.TimeRemaining)
	}

	config.TimeRemaining = uint(timeRemaining)
	return nil
}
