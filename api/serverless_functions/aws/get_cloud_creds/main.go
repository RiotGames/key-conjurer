package main

import (
	"fmt"

	"github.com/riotgames/key-conjurer/api/consts"
	"github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/api/settings"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	fmt.Printf(`Starting GetAWSCreds Lambda
	Version: %v
	`, consts.Version)
	cfg, err := settings.NewSettings()
	if err != nil {
		panic(err)
	}

	h := keyconjurer.NewHandler(cfg)
	lambda.Start(h.GetTemporaryCredentialEventHandler)
}
