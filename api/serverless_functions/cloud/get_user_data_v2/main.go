package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/riotgames/key-conjurer/pkg/httputil"
)

func ServeUserApplications(client *okta.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

func main() {
	var client okta.Client
	fn := httputil.Lambdaify(ServeUserApplications(&client))
	lambda.Start(fn)
}
