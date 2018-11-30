package FElement

import (
	h "baseadmin/helper"
	"strings"
)

const (
	INPUT_BUTTON_TEMPLATE = `<button class="btn btn-%subtype% %class%" %attrs%>%label%</button>
%note%`
)

type InputButton struct {
	Label      string
	Name       string
	Id         string
	Class      string
	Disabled   bool
	Note       string
	Submit     bool
	PullLeft   bool
	PullRight  bool
	Attributes map[string]string
}

func (b *InputButton) AddAttribute(attribute string, value string) {
	if b.Attributes == nil {
		b.Attributes = map[string]string{}
	}
	attributes, ok := b.Attributes[attribute]
	if ok {
		b.Attributes[attribute] = attributes + " " + value
	} else {
		b.Attributes[attribute] = value
	}
}

func (b InputButton) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering button", h.GetConfig().Mode.Debug)
	var replaces map[string]string = make(map[string]string)
	output := INPUT_BUTTON_TEMPLATE

	replaces["%label%"] = b.Label

	replaces["%note%"] = ""
	if b.Note != "" {
		replaces["%note%"] = h.Replace(NOTE_TEMPLATE, []string{"%note%"}, []string{b.Note})
	}

	replaces["%class%"] = b.Class

	replaces["%subtype%"] = "success"
	if b.Submit == true {
		replaces["%subtype%"] = "primary"
	}

	replaces["%attrs%"] = ""
	var attr []string

	for attrKey, attrValue := range b.Attributes {
		attr = append(attr, h.HtmlAttribute(attrKey, attrValue))
	}

	if b.Name != "" {
		attr = append(attr, h.HtmlAttribute("name", b.Name))
	}
	if b.Id != "" {
		attr = append(attr, h.HtmlAttribute("id", b.Id))
	}
	if b.Disabled == true {
		attr = append(attr, h.HtmlAttribute("disabled", "disabled"))
	}

	var btnType = "button"
	if b.Submit == true {
		btnType = "submit"
	}

	attr = append(attr, h.HtmlAttribute("type", btnType))

	replaces["%attrs%"] = strings.Join(attr, " ")

	for i, v := range replaces {
		output = h.Replace(output, []string{i}, []string{v})
	}

	var pull string = ""
	if b.PullLeft {
		pull = "left"
	} else if b.PullRight {
		pull = "right"
	}
	return GroupRender(output, b.HasPreOrPost(), false, nil, pull)
}

func (b InputButton) HasPreOrPost() bool {
	return false
}
