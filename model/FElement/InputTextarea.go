package FElement

import (
	h "baseadmin/helper"
	"html"
	"strconv"
	"strings"
)

const (
	INPUT_TEXTAREA_ROWS     = 4
	INPUT_TEXTAREA_COLS     = 60
	INPUT_TEXTAREA_TEMPLATE = `%label%
	<textarea class="form-control %class%" %attrs%>%value%</textarea>
%note%`
)

type InputTextarea struct {
	Label       string
	Name        string
	Id          string
	Class       string
	Placeholder string
	Disabled    bool
	Readonly    bool
	Value       string
	Note        string
	Cols        int
	Rows        int
}

func (t InputTextarea) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering textarea", h.GetConfig().Mode.Debug)
	var replaces map[string]string = make(map[string]string)
	output := INPUT_TEXTAREA_TEMPLATE

	var inpErrors []error
	inpError, contains := errs[t.Name]
	if contains {
		inpErrors = append(inpErrors, inpError)
	}

	replaces["%label%"] = ""
	if t.Label != "" {
		replaces["%forattr%"] = ""
		if t.Id != "" {
			replaces["%forattr%"] = h.Replace(LABEL_FOR_TEMPLATE, []string{"%for%"}, []string{t.Id})
		}
		replaces["%label%"] = h.Replace(LABEL_TEMPLATE, []string{"%forattr%", "%label%"}, []string{replaces["%forattr%"], t.Label})
	}

	replaces["%note%"] = ""
	if t.Note != "" {
		replaces["%note%"] = h.Replace(NOTE_TEMPLATE, []string{"%note%"}, []string{t.Note})
	}

	replaces["%value%"] = ""
	if t.Value != "" {
		replaces["%value%"] = html.EscapeString(t.Value)
	}

	replaces["%class%"] = t.Class

	replaces["%attrs%"] = ""
	var attr []string

	var rows int = INPUT_TEXTAREA_ROWS
	var cols int = INPUT_TEXTAREA_COLS

	if t.Rows > 0 {
		rows = t.Rows
	}

	if t.Cols > 0 {
		cols = t.Cols
	}

	attr = append(attr, h.HTMLAttribute("rows", strconv.Itoa(rows)))
	attr = append(attr, h.HTMLAttribute("cols", strconv.Itoa(cols)))

	if t.Name != "" {
		attr = append(attr, h.HTMLAttribute("name", t.Name))
	}
	if t.Id != "" {
		attr = append(attr, h.HTMLAttribute("id", t.Id))
	}
	if t.Placeholder != "" {
		attr = append(attr, h.HTMLAttribute("placeholder", t.Placeholder))
	}
	if t.Disabled == true {
		attr = append(attr, h.HTMLAttribute("disabled", "disabled"))
	}
	if t.Readonly == true {
		attr = append(attr, h.HTMLAttribute("readonly", "readonly"))
	}

	replaces["%attrs%"] = strings.Join(attr, " ")

	for i, v := range replaces {
		output = h.Replace(output, []string{i}, []string{v})
	}

	return GroupRender(output, t.HasPreOrPost(), false, inpErrors, "")
}

func (t InputTextarea) HasPreOrPost() bool {
	return false
}
