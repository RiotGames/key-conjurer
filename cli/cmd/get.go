package cmd

import (
	"log"

	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var (
	ttl            uint
	timeRemaining  uint
	promptForCreds bool
	outputType     string
	awsCliPath     string
)

func init() {
	getCmd.Flags().UintVar(&ttl, "ttl", 0, "The key timeout in hours from 1 to 8.")
	getCmd.Flags().UintVarP(&timeRemaining, "time-remaining", "t", keyconjurer.DefaultTimeRemaining, "Request new keys if there are no keys in the environment or the current keys expire within <time-remaining> minutes. Defaults to 60.")
	getCmd.Flags().BoolVarP(&promptForCreds, "creds-prompt", "c", false, "Prompt for username and password through stdin. Can be piped in using the following format \"<username>\\n<pasword>\\n\".")
	getCmd.Flags().StringVarP(&outputType, "out", "o", "env", "Format to save new credentials in. Supported outputs: env, awscli")
	getCmd.Flags().StringVarP(&awsCliPath, "awscli", "", "~/.aws/", "Path for directory used by the aws-cli tool. Default is \"~/.aws\".")
}

var getCmd = &cobra.Command{
	Use:     "get <accountName/alias>",
	Short:   "Retrieves temporary AWS API credentials.",
	Long:    "Retrieves temporary AWS API credentials for the specified account.  It sends a push request to the first Duo device it finds associated with your account.",
	Example: "keyconjurer get <accountName/alias>",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userData := keyconjurer.Login(keyConjurerRcPath, false)
		accountName := args[0]

		// make sure we enforce limit
		if ttl > 8 {
			ttl = 8
		}

		credentials, err := keyconjurer.GetCredentials(userData, accountName, ttl)
		if err != nil {
			if saveError := userData.Save(); saveError != nil {
				log.Println(saveError)
			}
			log.Fatal(err)
		}
		account, err := userData.FindAccount(accountName)
		if err != nil {
			if saveError := userData.Save(); err != nil {
				log.Println(saveError)
			}
			log.Fatal(err)
		}

		switch outputType {
		case "env":
			credentials.PrintCredsForEnv()
		case "awscli":
			newCliEntry := keyconjurer.NewAWSCliEntry(credentials, account)
			keyconjurer.SaveAWSCredentialInCLI(awsCliPath, newCliEntry)
		default:
			log.Fatalf("%s is an invalid output type.\n", outputType)
		}

	}}
