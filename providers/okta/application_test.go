package okta

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_findAuthenticatorByMethodType(t *testing.T) {
	blob := `
{
	"rel": [
		"create-form"
	],
	"name": "select-authenticator-authenticate",
	"href": "https://sso.example.com/idp/idx/challenge",
	"method": "POST",
	"produces": "application/ion+json; okta-version=1.0.0",
	"value": [
		{
			"name": "authenticator",
			"type": "object",
			"options": [
				{
					"label": "the label you are looking for",
					"value": {
						"form": {
							"value": [
								{
									"name": "id",
									"required": true,
									"value": "the id you are looking for",
									"mutable": false
								},
								{
									"name": "methodType",
									"required": false,
									"value": "idp",
									"mutable": false
								}
							]
						}
					},
					"relatesTo": "$.authenticatorEnrollments.value[0]"
				},
				{
					"label": "not the label you are looking for",
					"value": {
						"form": {
							"value": [
								{
									"name": "id",
									"required": true,
									"value": "not the id you are looking for",
									"mutable": false
								},
								{
									"name": "methodType",
									"required": false,
									"value": "duo",
									"mutable": false
								}
							]
						}
					},
					"relatesTo": "$.authenticatorEnrollments.value[1]"
				}
			]
		},
		{
			"name": "stateHandle",
			"required": true,
			"value": "a state handle",
			"visible": false,
			"mutable": false
		}
	],
	"accepts": "application/json; okta-version=1.0.0"
}
`

	var rem Remediation
	require.NoError(t, json.Unmarshal([]byte(blob), &rem))

	authenticatorId, ok := findAuthenticatorByMethodType(rem, "idp")
	assert.True(t, ok)
	assert.Equal(t, "the id you are looking for", authenticatorId)
}
