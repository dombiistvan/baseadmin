package FElement

import (
	h "baseadmin/helper"
	"html"
	"strings"
)

const (
	INPUT_HIDDEN_TEMPLATE = `<input class="form-control %class%" %attrs% />`
)

type InputHidden struct {
	Name     string
	Id       string
	Class    string
	Disabled bool
	Readonly bool
	Value    string
}

func (i InputHidden) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering hidden", h.GetConfig().Mode.Debug)
	var replaces map[string]string = make(map[string]string)
	output := INPUT_HIDDEN_TEMPLATE

	replaces["%class%"] = i.Class

	replaces["%attrs%"] = ""
	var attr []string

	attr = append(attr, h.HTMLAttribute("type", "hidden"))

	if i.Name != "" {
		attr = append(attr, h.HTMLAttribute("name", i.Name))
	}
	if i.Id != "" {
		attr = append(attr, h.HTMLAttribute("id", i.Id))
	}
	if i.Value != "" {
		attr = append(attr, h.HTMLAttribute("value", html.EscapeString(i.Value)))
	}
	if i.Disabled == true {
		attr = append(attr, h.HTMLAttribute("disabled", "disabled"))
	}
	if i.Readonly == true {
		attr = append(attr, h.HTMLAttribute("readonly", "readonly"))
	}

	replaces["%attrs%"] = strings.Join(attr, " ")

	for i, v := range replaces {
		output = h.Replace(output, []string{i}, []string{v})
	}

	return output
}

func (i InputHidden) HasPreOrPost() bool {
	return false
}
