package main

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	setCmd.AddCommand(setTTLCmd)
	setCmd.AddCommand(setTimeRemainingCmd)
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets config values.",
	Long:  "Sets config values.",
}

var setTTLCmd = &cobra.Command{
	Use:   "ttl <ttl>",
	Short: "Sets ttl value in number of hours.",
	Long:  "Sets ttl value in number of hours.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ttl, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("unable to parse value %s", args[0])
		}

		userData.SetTTL(uint(ttl))
		return nil
	},
}

var setTimeRemainingCmd = &cobra.Command{
	Use:   "time-remaining <timeRemaining>",
	Short: "Sets time remaining value in number of minutes.",
	Long:  "Sets time remaining value in number of minutes. Using minutes is an artifact from when keys could only live for 1 hour.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timeRemaining, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("unable to parse value %s", args[0])
		}

		userData.SetTimeRemaining(uint(timeRemaining))
		return nil
	},
}
