package okta

import (
	"github.com/okta/okta-sdk-golang/v2/okta"
)

func getHrefLink(app okta.Application) (string, bool) {
	links := app.Links.(map[string]interface{})
	appLinks := links["appLinks"].([]interface{})
	for _, interf := range appLinks {
		entry := interf.(map[string]interface{})
		if entry["type"] == "text/html" {
			return entry["href"].(string), true
		}
	}

	return "", false
}
