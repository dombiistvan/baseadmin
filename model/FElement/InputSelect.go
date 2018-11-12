package FElement

import (
	h "base/helper"
	"strings"
	"fmt"
	"html"
)

const (
	INPUT_SELECT_TEMPLATE = `%label%
	<select class="form-control %class%" %attrs%>%options%</select>
%note%`
	INPUT_SELECT_OPTION_TEMPLATE = `<option %value% %selected%>%label%</label>`
)

type InputSelect struct {
	Label    string
	Name     string
	Id       string
	Class    string
	Disabled bool
	Readonly bool
	Value    []string
	Multiple bool
	Option   []map[string]string
	Note     string
}

func (s InputSelect) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering text", h.GetConfig().Mode.Debug);
	var replaces map[string]string = make(map[string]string);
	output := INPUT_SELECT_TEMPLATE;

	var inpErrors []error;
	inpError, contains := errs[s.Name];
	if(contains){
		inpErrors = append(inpErrors, inpError);
	}

	replaces["%label%"] = "";
	if (s.Label != "") {
		replaces["%forattr%"] = "";
		if (s.Id != "") {
			replaces["%forattr%"] = h.Replace(LABEL_FOR_TEMPLATE, []string{"%for%"}, []string{s.Id})
		}
		replaces["%label%"] = h.Replace(LABEL_TEMPLATE, []string{"%forattr%","%label%"}, []string{replaces["%forattr%"],s.Label});
	}

	replaces["%note%"] = "";
	if (s.Note != "") {
		replaces["%note%"] = h.Replace(NOTE_TEMPLATE,[]string{"%note%"},[]string{s.Note});
	}

	replaces["%class%"] = s.Class;
	replaces["%attrs%"] = "";
	var attr []string;
	if (s.Name != "") {
		attr = append(attr, h.HtmlAttribute("name", s.Name));
	}
	if (s.Id != "") {
		attr = append(attr, h.HtmlAttribute("id", s.Id));
	}
	if (s.Multiple == true) {
		attr = append(attr, h.HtmlAttribute("multiple", "multiple"));
	}
	if (s.Disabled == true) {
		attr = append(attr, h.HtmlAttribute("disabled", "disabled"));
	}
	if (s.Readonly == true) {
		attr = append(attr, h.HtmlAttribute("readonly", "readonly"));
	}

	replaces["%attrs%"] = strings.Join(attr, " ");
	replaces["%options%"] = "";
	for _, o := range s.Option {
		optionTemplate := INPUT_SELECT_OPTION_TEMPLATE;
		optionTemplate = h.Replace(optionTemplate,[]string{"%value%","%label%"},[]string{fmt.Sprintf(`value="%s"`,html.EscapeString(o["value"])),html.EscapeString(o["label"])});
		selected := "";
		for _,sv := range s.Value{
			if(o["value"] == sv){
				selected = `SELECTED="SELECTED"`;
			}
		}
		replaces["%options%"] += "\n"+h.Replace(optionTemplate,[]string{"%selected%"},[]string{selected});

	}

	for i, v := range replaces {
		output = h.Replace(output, []string{i}, []string{v});
	}

	return GroupRender(output, s.HasPreOrPost(),false, inpErrors,"");
}

func (s InputSelect) HasPreOrPost() bool {
	return false;
}
