package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/pkg/httputil"
)

func main() {
	var service keyconjurer.OktaService
	fn := httputil.Lambdaify(keyconjurer.ServeUserApplications(service))
	lambda.Start(fn)
}
