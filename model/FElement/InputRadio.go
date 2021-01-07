package FElement

import (
	h "baseadmin/helper"
	"html"
	"strings"
)

const (
	INPUT_RADIO_TEMPLATE = `<div class="radio">
	<label>
    	 <input %attrs%>%label%
	</label>
</div>`
	INPUT_RADIO_TEMPLATE_INLINE = `<label class="radio-inline">
    	 <input %attrs%>%label%
	</label>`
)

type InputRadio struct {
	Label         string
	Name          string
	Id            string
	Class         string
	Disabled      bool
	Readonly      bool
	Value         string
	SelectedValue string
	DisplayInline bool
}

func (r InputRadio) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering radio", h.GetConfig().Mode.Debug)
	var replaces map[string]string = make(map[string]string)
	output := INPUT_RADIO_TEMPLATE
	if r.DisplayInline {
		output = INPUT_RADIO_TEMPLATE_INLINE
	}

	var inpErrors []error
	inpError, contains := errs[r.Name]
	if contains {
		inpErrors = append(inpErrors, inpError)
	}

	replaces["%label%"] = r.Label

	replaces["%class%"] = r.Class
	replaces["%attrs%"] = ""
	var attr []string
	attr = append(attr, h.HTMLAttribute("type", "radio"))
	if r.Name != "" {
		attr = append(attr, h.HTMLAttribute("name", r.Name))
	}
	if r.Id != "" {
		attr = append(attr, h.HTMLAttribute("id", r.Id))
	}
	if r.Value != "" {
		attr = append(attr, h.HTMLAttribute("value", html.EscapeString(r.Value)))
	}
	if r.Value != "" && r.SelectedValue != "" && r.Value == r.SelectedValue {
		attr = append(attr, h.HTMLAttribute("checked", "checked"))
	}
	if r.Disabled == true {
		attr = append(attr, h.HTMLAttribute("disabled", "disabled"))
	}
	if r.Readonly == true {
		attr = append(attr, h.HTMLAttribute("readonly", "readonly"))
	}

	replaces["%attrs%"] = strings.Join(attr, " ")

	for i, v := range replaces {
		output = h.Replace(output, []string{i}, []string{v})
	}

	return GroupRender(output, r.HasPreOrPost(), false, inpErrors, "")
}

func (r InputRadio) HasPreOrPost() bool {
	return false
}
