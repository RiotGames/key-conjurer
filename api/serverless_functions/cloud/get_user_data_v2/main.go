package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/okta/okta-sdk-golang/v2/okta"
)

func ServeUserApplications(client *okta.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func lambdaify(next http.Handler) lambda.Handler {
	return nil
}

func main() {
	var client okta.Client
	fn := lambdaify(ServeUserApplications(&client))
	lambda.Start(fn)
}
