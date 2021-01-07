package FElement

import (
	h "baseadmin/helper"
	"html"
	"strings"
)

const (
	PRE_STR_TEMPLATE    = `<span class="input-group-addon">%text%</span>`
	POST_STR_TEMPLATE   = `<span class="input-group-addon">%text%</span>`
	PRE_FA_TEMPLATE     = `<span class="input-group-addon"><i class="fa fa-%icon%"></i></span>`
	POST_FA_TEMPLATE    = `<span class="input-group-addon"><i class="fa fa-%icon%"></i></span>`
	INPUT_TEXT_TEMPLATE = `%label%
	%pre%<input class="form-control %class%" %attrs% />%post%
%note%`
)

type InputText struct {
	Label       string
	Name        string
	Id          string
	Class       string
	Placeholder string
	Disabled    bool
	Readonly    bool
	Value       string
	Note        string
	PreStr      string
	PostStr     string
	PreFAIcon   string
	PostFAIcon  string
}

func (i InputText) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering text", h.GetConfig().Mode.Debug)
	var replaces map[string]string = make(map[string]string)
	output := INPUT_TEXT_TEMPLATE

	var inpErrors []error
	inpError, contains := errs[i.Name]
	if contains {
		inpErrors = append(inpErrors, inpError)
	}

	replaces["%label%"] = ""
	if i.Label != "" {
		replaces["%forattr%"] = ""
		if i.Id != "" {
			replaces["%forattr%"] = h.Replace(LABEL_FOR_TEMPLATE, []string{"%for%"}, []string{i.Id})
		}
		replaces["%label%"] = h.Replace(LABEL_TEMPLATE, []string{"%forattr%", "%label%"}, []string{replaces["%forattr%"], i.Label})
	}

	replaces["%pre%"] = ""
	if i.PreFAIcon != "" {
		replaces["%pre%"] = h.Replace(PRE_FA_TEMPLATE, []string{"%icon%"}, []string{i.PreFAIcon})
	} else if i.PreStr != "" {
		replaces["%pre%"] = h.Replace(PRE_STR_TEMPLATE, []string{"%text%"}, []string{i.PreStr})
	}

	replaces["%post%"] = ""
	if i.PostFAIcon != "" {
		replaces["%post%"] = h.Replace(POST_FA_TEMPLATE, []string{"%icon%"}, []string{i.PostFAIcon})
	} else if i.PostStr != "" {
		replaces["%post%"] = h.Replace(POST_STR_TEMPLATE, []string{"%text%"}, []string{i.PostStr})
	}

	replaces["%note%"] = ""
	if i.Note != "" {
		replaces["%note%"] = h.Replace(NOTE_TEMPLATE, []string{"%note%"}, []string{i.Note})
	}

	replaces["%class%"] = i.Class

	replaces["%attrs%"] = ""
	var attr []string

	attr = append(attr, h.HTMLAttribute("type", "text"))

	if i.Name != "" {
		attr = append(attr, h.HTMLAttribute("name", i.Name))
	}
	if i.Id != "" {
		attr = append(attr, h.HTMLAttribute("id", i.Id))
	}
	if i.Value != "" {
		attr = append(attr, h.HTMLAttribute("value", html.EscapeString(i.Value)))
	}
	if i.Placeholder != "" {
		attr = append(attr, h.HTMLAttribute("placeholder", i.Placeholder))
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

	return GroupRender(output, i.HasPreOrPost(), false, inpErrors, "")
}

func (i InputText) HasPreOrPost() bool {
	return i.PreStr != "" || i.PostStr != "" || i.PreFAIcon != "" || i.PostFAIcon != ""
}
