package okta

import (
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

func collectFormValues(node *html.Node) (Form, error) {
	var f Form
	if node.Type != html.ElementNode || node.Data != "form" {
		panic("invalid element given to parseForm")
	}

	f.Method, _ = getAttribute(node.Attr, "method")
	for node := node.FirstChild; node != nil; node = node.NextSibling {
		if node.Data == "input" {
			if f.Inputs == nil {
				f.Inputs = make(map[string]string)
			}

			name, _ := getAttribute(node.Attr, "name")
			val, _ := getAttribute(node.Attr, "value")
			f.Inputs[name] = val
		}
	}

	return f, nil
}

func findFormByID(node *html.Node, id string) (Form, bool) {
	if node == nil {
		return Form{}, false
	}

	if node.Data == "form" {
		attrID, _ := getAttribute(node.Attr, "id")
		if attrID == id {
			form, err := collectFormValues(node)
			return form, err == nil
		}
	}

	for node := node.FirstChild; node != nil; node = node.NextSibling {
		form, ok := findFormByID(node, id)
		if ok {
			return form, ok
		}
	}

	return Form{}, false
}
