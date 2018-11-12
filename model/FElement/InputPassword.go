package FElement

import (
	h "base/helper"
	"strings"
	"html"
)

const (
	INPUT_PASSWORD_TEMPLATE = `%label%
	%pre%<input class="form-control %class%" %attrs% />%post%
%note%`
)

type InputPassword struct {
	Label       string
	Name        string
	Id          string
	Class       string
	Disabled    bool
	Readonly    bool
	Value		string
	Note        string
	PreStr      string
	PostStr     string
	PreFAIcon   string
	PostFAIcon  string
}

func (p InputPassword) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering text",h.GetConfig().Mode.Debug);
	var replaces map[string]string = make(map[string]string);
	output := INPUT_PASSWORD_TEMPLATE;

	var inpErrors []error;
	inpError, contains := errs[p.Name];
	if(contains){
		inpErrors = append(inpErrors, inpError);
	}

	replaces["%label%"] = "";
	if (p.Label != "") {
		replaces["%forattr%"] = "";
		if (p.Id != "") {
			replaces["%forattr%"] = h.Replace(LABEL_FOR_TEMPLATE, []string{"%for%"}, []string{p.Id})
		}
		replaces["%label%"] = h.Replace(LABEL_TEMPLATE, []string{"%forattr%","%label%"}, []string{replaces["%forattr%"],p.Label});
	}

	replaces["%pre%"] = "";
	if(p.PreFAIcon != ""){
		replaces["%pre%"] = h.Replace(PRE_FA_TEMPLATE,[]string{"%icon%"},[]string{p.PreFAIcon});
	} else if(p.PreStr != ""){
		replaces["%pre%"] = h.Replace(PRE_STR_TEMPLATE,[]string{"%text%"},[]string{p.PreStr});
	}

	replaces["%post%"] = "";
	if(p.PostFAIcon != ""){
		replaces["%post%"] = h.Replace(POST_FA_TEMPLATE,[]string{"%icon%"},[]string{p.PostFAIcon});
	} else if(p.PostStr != ""){
		replaces["%post%"] = h.Replace(POST_STR_TEMPLATE,[]string{"%text%"},[]string{p.PostStr});
	}

	replaces["%note%"] = "";
	if(p.Note != ""){
		replaces["%note%"] = h.Replace(NOTE_TEMPLATE,[]string{"%note%"},[]string{p.Note});
	}

	replaces["%class%"] = p.Class;

	replaces["%attrs%"] = "";
	var attr []string;

	attr = append(attr, h.HtmlAttribute("type", "password"));

	if (p.Name != "") {
		attr = append(attr, h.HtmlAttribute("name", p.Name));
	}
	if (p.Id != "") {
		attr = append(attr, h.HtmlAttribute("id", p.Id));
	}
	if (p.Value != "") {
		attr = append(attr, h.HtmlAttribute("value", html.EscapeString(p.Value)));
	}
	if (p.Disabled == true) {
		attr = append(attr, h.HtmlAttribute("disabled", "disabled"));
	}
	if (p.Readonly == true) {
		attr = append(attr, h.HtmlAttribute("readonly", "readonly"));
	}

	replaces["%attrs%"] = strings.Join(attr," ");

	for i,v := range replaces{
		output = h.Replace(output,[]string{i},[]string{v});
	}

	return GroupRender(output, p.HasPreOrPost(), false, inpErrors,"");
}

func (p InputPassword) HasPreOrPost() bool {
	return p.PreStr != "" || p.PostStr!="" || p.PreFAIcon != "" || p.PostFAIcon != "";
}
