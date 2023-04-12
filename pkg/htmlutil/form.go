package htmlutil

import (
	"errors"
	"net/url"

	"golang.org/x/net/html"
)

type Form struct {
	Method string
	Inputs map[string]string
}

func (f *Form) Set(key, value string) {
	if f.Inputs == nil {
		f.Inputs = make(map[string]string)
	}

	f.Inputs[key] = value
}

func (f Form) Values() url.Values {
	v := url.Values{}
	for key, val := range f.Inputs {
		v.Set(key, val)
	}
	return v
}

func getAttribute(attrs []html.Attribute, key string) (string, bool) {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Val, true
		}
	}

	return "", false
}

// Walk traverses the given HTML node and calls the function walker on each one encountered.
//
// Walk will continue executing until walker returns false or Walk reaches the end of the tree.
func Walk(node *html.Node, walker func(node *html.Node) bool) {
	if node == nil {
		return
	}

	for node := node.FirstChild; node != nil; node = node.NextSibling {
		if node.Type != html.ElementNode {
			continue
		}

		if stop := walker(node); stop {
			return
		}

		Walk(node, walker)
	}
}

func collectFormValues(node *html.Node) (Form, error) {
	var f Form
	if node == nil || node.Type != html.ElementNode || node.Data != "form" {
		return Form{}, errors.New("invalid element given to parseForm")
	}

	f.Method, _ = getAttribute(node.Attr, "method")

	Walk(node, func(node *html.Node) bool {
		if node.Data != "input" {
			return false
		}

		if f.Inputs == nil {
			f.Inputs = make(map[string]string)
		}

		name, _ := getAttribute(node.Attr, "name")
		val, _ := getAttribute(node.Attr, "value")
		f.Inputs[name] = val
		return false
	})

	return f, nil
}

// FindFirstForm finds the first form within the given HTML document and returns it, or false if it doesn't exist.
func FindFirstForm(node *html.Node) (Form, bool) {
	if node == nil {
		return Form{}, false
	}

	if node.Data == "form" {
		form, err := collectFormValues(node)
		return form, err == nil
	}

	for node := node.FirstChild; node != nil; node = node.NextSibling {
		form, ok := FindFirstForm(node)
		if ok {
			return form, ok
		}
	}

	return Form{}, false
}

// FindFormByID returns the first form present in the given document with the given ID, or false if it doesn't exist.
func FindFormByID(tree *html.Node, id string) (Form, bool) {
	var formNode *html.Node
	Walk(tree, func(n *html.Node) bool {
		if n.Data == "form" {
			attrID, _ := getAttribute(n.Attr, "id")
			if attrID == id {
				formNode = n
				return true
			}
		}

		return false
	})

	form, err := collectFormValues(formNode)
	return form, err == nil
}
