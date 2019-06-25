package main

import (
	"fmt"
	"keyconjurer-lambda/consts"
	"keyconjurer-lambda/keyconjurer"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	fmt.Printf(`Starting GetUserData Lambda
	Version: %v
	`, consts.Version)
	lambda.Start(keyconjurer.GetUserDataEventHandler)
}
