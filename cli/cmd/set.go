package cmd

import (
	"log"
	"os"
	"strconv"

	"keyconjurer-cli/keyconjurer"

	"github.com/spf13/cobra"
)

func init() {
	setCmd.AddCommand(setTTLCmd)
	setCmd.AddCommand(setTimeRemainingCmd)
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets config values.",
	Long:  "Sets config values."}

var setTTLCmd = &cobra.Command{
	Use:   "ttl <ttl>",
	Short: "Sets ttl value in number of hours.",
	Long:  "Sets ttl value in number of hours.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userData := keyconjurer.Login(keyConjurerRcPath, false)
		ttl, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			log.Printf("Unable to parse value %v\n", args[0])
			os.Exit(1)
		}
		userData.SetTTL(uint(ttl))
		userData.Save()
	}}

var setTimeRemainingCmd = &cobra.Command{
	Use:   "time-remaining <timeRemaining>",
	Short: "Sets time remaining value in number of minutes.",
	Long:  "Sets time remaining value in number of minutes. Using minutes is an artifact from when keys could only live for 1 hour.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userData := keyconjurer.Login(keyConjurerRcPath, false)
		timeRemaining, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			log.Printf("Unable to parse value %v\n", args[0])
			os.Exit(1)
		}
		userData.SetTimeRemaining(uint(timeRemaining))
		userData.Save()
	}}
