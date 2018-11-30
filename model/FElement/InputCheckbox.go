package FElement

import (
	h "baseadmin/helper"
	"fmt"
	"html"
	"strings"
)

const (
	INPUT_CHECKBOX_TEMPLATE = `<div class="checkbox">
	<label>
    	 <input %attrs% %class%>%label%
	</label>
</div>`
	INPUT_CHECKBOX_TEMPLATE_INLINE = `<label class="checkbox-inline">
    	 <input %attrs% %class%>%label%
	</label>`
)

type InputCheckbox struct {
	Label          string
	Name           string
	Id             string
	Class          string
	Disabled       bool
	Readonly       bool
	Value          string
	SelectedValues []string
	DisplayInline  bool
}

func (c InputCheckbox) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering checkbox", h.GetConfig().Mode.Debug)
	var replaces map[string]string = make(map[string]string)
	output := INPUT_CHECKBOX_TEMPLATE
	if c.DisplayInline {
		output = INPUT_CHECKBOX_TEMPLATE_INLINE
	}

	var inpErrors []error
	inpError, contains := errs[c.Name]
	if contains {
		inpErrors = append(inpErrors, inpError)
	}

	replaces["%label%"] = c.Label

	replaces["%class%"] = ""
	if c.Class != "" {
		replaces["%class%"] = fmt.Sprintf(`class="%v"`, c.Class)
	}

	replaces["%attrs%"] = ""
	var attr []string

	attr = append(attr, h.HtmlAttribute("type", "checkbox"))

	if c.Name != "" {
		attr = append(attr, h.HtmlAttribute("name", c.Name))
	}
	if c.Id != "" {
		attr = append(attr, h.HtmlAttribute("id", c.Id))
	}
	if c.Value != "" {
		attr = append(attr, h.HtmlAttribute("value", html.EscapeString(c.Value)))
	}

	if c.Value != "" && len(c.SelectedValues) > 0 {
		for _, sv := range c.SelectedValues {
			if sv == c.Value {
				attr = append(attr, h.HtmlAttribute("checked", "checked"))
				break
			}
		}
	}
	if c.Disabled == true {
		attr = append(attr, h.HtmlAttribute("disabled", "disabled"))
	}
	if c.Readonly == true {
		attr = append(attr, h.HtmlAttribute("readonly", "readonly"))
	}

	replaces["%attrs%"] = strings.Join(attr, " ")
	for i, v := range replaces {
		output = h.Replace(output, []string{i}, []string{v})
	}

	return GroupRender(output, c.HasPreOrPost(), c.DisplayInline, inpErrors, "")
}

func (c InputCheckbox) HasPreOrPost() bool {
	return false
}
