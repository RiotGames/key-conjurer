package okta

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func Test_CorrectlyParsesJSONFromDuo(t *testing.T) {
	blob := []byte(`{"success": {"href": "http://example.com"}}`)
	require.Equal(t, "http://example.com", gjson.GetBytes(blob, "success.href").Str)
}
