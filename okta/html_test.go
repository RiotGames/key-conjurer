package okta

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

	form, ok := FindFirstForm(node)
	require.True(t, ok)

	require.Equal(t, "a jwt object", form.Inputs["tx"])
	require.Equal(t, "xsrf token", form.Inputs["_xsrf"])
	require.Equal(t, "post", form.Method)
}

func Test_FindFirstFormFindsFormsNestedWithinChildren(t *testing.T) {
	doc := `<html>
<body>
	<div />
	<div>
		<form method="post" id="form">
			<input type="string" name="value1" value="a non-empty value" />
		</form>
	</div>
	<form id="another form"></form>
</body>
</html>`

	node, err := html.Parse(strings.NewReader(doc))
	require.NoError(t, err)

	form, ok := FindFirstForm(node)
	require.True(t, ok)
	require.Equal(t, "a non-empty value", form.Inputs["value1"])
}

func Test_FindFirstFormFindsNestedInputs(t *testing.T) {
	doc := `<html>
<body>
	<div />
	<div>
		<form method="post" id="form">
			<div>
				<input type="string" name="value1" value="a non-empty value" />
			</div>
		</form>
	</div>
	<form id="another form"></form>
</body>
</html>`

	node, err := html.Parse(strings.NewReader(doc))
	require.NoError(t, err)

	form, ok := FindFirstForm(node)
	require.True(t, ok)

	require.Equal(t, "a non-empty value", form.Inputs["value1"])
}

func TestWalkWalksElementsCorrectly(t *testing.T) {
	doc := `<html>
<head />
<body>
	<div />
	<div>
		<form method="post" id="form">
			<div>
				<input type="string" name="value1" value="a non-empty value" />
			</div>
		</form>
	</div>
	<form id="another form"></form>
</body>
</html>`
	expected := []string{"html", "head", "body", "div", "div", "form", "div", "input", "form"}
	elements := []string{}
	node, err := html.Parse(strings.NewReader(doc))
	require.NoError(t, err)

	Walk(node, func(node *html.Node) bool {
		elements = append(elements, node.Data)
		return false
	})

	assert.Equal(t, expected, elements)
}

func TestWalkStopsWhenFunctionReturnsTrue(t *testing.T) {
	doc := `<html>
<head />
<body>
	<div id=1 />
	<div id=2 />
	<div id=3>
		<div id=4>
	</div>
	<div id=5 />
</body>
</html>`

	var lastElement *html.Node
	node, err := html.Parse(strings.NewReader(doc))
	require.NoError(t, err)

	Walk(node, func(node *html.Node) bool {
		lastElement = node
		id, _ := getAttribute(lastElement.Attr, "id")
		return id == "4"
	})

	require.NotNil(t, lastElement)
	id, _ := getAttribute(lastElement.Attr, "id")
	assert.Equal(t, id, "4")
}
