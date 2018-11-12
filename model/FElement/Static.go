package FElement

import (
	h "base/helper"
	"strings"
)

const (
	STATIC_TEMPLATE = `%label%
<p class="form-control-static %class%" %attrs%>%value%</p>`
)

type Static struct {
	Label string
	Name  string
	Id    string
	Class string
	Value string
}

func (s Static) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering input", h.GetConfig().Mode.Debug)
	var replaces map[string]string = make(map[string]string)
	output := STATIC_TEMPLATE
	replaces["%label%"] = ""
	if s.Label != "" {
		replaces["%forattr%"] = ""
		if s.Id != "" {
			replaces["%forattr%"] = h.Replace(LABEL_FOR_TEMPLATE, []string{"%for%"}, []string{s.Id})
		}
		replaces["%label%"] = h.Replace(LABEL_TEMPLATE, []string{"%forattr%", "%label%"}, []string{replaces["%forattr%"], s.Label})
	}

	replaces["%value%"] = ""
	if s.Value != "" {
		replaces["%value%"] = s.Value
	}

	replaces["%class%"] = s.Class
	replaces["%attrs%"] = ""
	var attr []string
	if s.Name != "" {
		attr = append(attr, h.HtmlAttribute("name", s.Name))
	}
	if s.Id != "" {
		attr = append(attr, h.HtmlAttribute("id", s.Id))
	}

	replaces["%attrs%"] = strings.Join(attr, " ")
	for i, v := range replaces {
		output = h.Replace(output, []string{i}, []string{v})
	}

	return GroupRender(output, s.HasPreOrPost(), false, nil, "")
}

func (s Static) HasPreOrPost() bool {
	return false
}
