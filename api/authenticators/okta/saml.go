package okta

import (
	"errors"
	"fmt"
	"io"

	"github.com/riotgames/key-conjurer/api/core"

	"golang.org/x/net/html"
)

func walkHTMLTree(node *html.Node, condition func(*html.Node) bool) (*html.Node, bool) {
	if condition(node) {
		return node, true
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if n, present := walkHTMLTree(child, condition); present {
			return n, true
		}
	}

	return nil, false
}

var errMalformedResponse = errors.New("malformed response")

func getAttribute(node *html.Node, name string) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val, true
		}
	}

	return "", false
}

func hasAttributeEquals(node *html.Node, name string, value string) bool {
	attr, present := getAttribute(node, name)
	return present && attr == value
}

func extractEncodedSAMLResponseFromBody(node *html.Node) (str string, err error) {
	node, ok := walkHTMLTree(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "form"
	})

	if !ok {
		err = fmt.Errorf("could not find form: %w", errMalformedResponse)
		return
	}

	input, ok := walkHTMLTree(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "input" && hasAttributeEquals(n, "name", "SAMLResponse")
	})

	if !ok {
		err = fmt.Errorf("could not find input with name SAMLResponse: %w", errMalformedResponse)
		return
	}

	value, ok := getAttribute(input, "value")
	if !ok {
		err = fmt.Errorf("could not find value attribute on input: %w", errMalformedResponse)
		return
	}

	return value, nil
}

func extractSAMLResponse(reader io.Reader) (*core.SAMLResponse, error) {
	document, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	encoded, err := extractEncodedSAMLResponseFromBody(document)
	if err != nil {
		return nil, nil
	}

	return core.ParseEncodedResponse(encoded)
}
