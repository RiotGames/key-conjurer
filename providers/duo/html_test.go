package duo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func Test_findFirstForm(t *testing.T) {
	doc := `<html>
<body>
	<div />
	<form method="post" id="form">
		<input type="hidden" name="tx" value="a jwt object" />
		<input type="hidden" name="parent" value="None" />
		<input type="hidden" name="_xsrf" value="xsrf token" />
		<input type="hidden" name="java_version" value="" />
		<input type="hidden" name="flash_version" value="" />
		<input type="hidden" name="screen_resolution_width" value="" />
		<input type="hidden" name="screen_resolution_height" value="" />
		<input type="hidden" name="color_depth" value="" />
		<input type="hidden" name="ch_ua_error" value="" />
		<input type="hidden" name="client_hints" value="" />

		<input type="hidden" name="is_cef_browser" value=""/>
		<input type="hidden" name="is_ipad_os" value="" />
		<input type="hidden" name="is_ie_compatibility_mode" value="" />
		<input type="hidden" name="is_user_verifying_platform_authenticator_available" value="" />
		<input type="hidden" name="user_verifying_platform_authenticator_available_error" value="" />
		<input type="hidden" name="acting_ie_version" value="" />
		<div id='react-test-container'></div>
		<input type="hidden" name="react_support" value="" />
		<input type="hidden" name="react_support_error_message" value="" />
	</form>
</body>
</html>`
	node, err := html.Parse(strings.NewReader(doc))
	require.NoError(t, err)

	form, ok := findFirstForm(node)
	require.True(t, ok)

	require.Equal(t, "a jwt object", form.Inputs["tx"])
	require.Equal(t, "xsrf token", form.Inputs["_xsrf"])
	require.Equal(t, "post", form.Method)
}
