package model

import (
	h "base/helper"
	"fmt"
	"strings"
)

const FORM_TEMPLATE = `<form role="form" action="%action%" method="%method%" %class% %enctype%>
	%formgroup%
</form>`

type Form struct {
	Action     string
	Method     string
	UploadFile bool
	Fieldset   []Fieldset
	Inline     bool
	Errors     map[string]error
	Class      []string
}

func (f Form) Render() string {
	h.PrintlnIf("Rendering form", h.GetConfig().Mode.Debug)
	foutput := FORM_TEMPLATE
	var fsoutput []string
	for _, fs := range f.Fieldset {
		fsoutput = append(fsoutput, fs.Render(f.Errors))
	}

	foutput = h.Replace(foutput, []string{"%action%", "%method%"}, []string{f.Action, f.Method})
	encReplace := ""
	if f.UploadFile {
		encReplace = `enctype="multipart/form-data"`
	}
	foutput = h.Replace(foutput, []string{"%enctype%", "%formgroup%"}, []string{encReplace, strings.Join(fsoutput, "\n")})

	class := ""
	if f.Inline {
		f.AddClass("form-inline")
	}
	if len(f.Class) > 0 {
		class = fmt.Sprintf(`class="%v"`, strings.Join(f.Class, " "))
	}
	foutput = h.Replace(foutput, []string{"%class%"}, []string{class})
	return foutput
}

func (f *Form) SetErrors(errs map[string]error) {
	f.Errors = errs
}

func NewForm(method string, action string, uploadFiles bool, inline bool) Form {
	var form = Form{
		action,
		method,
		uploadFiles,
		nil,
		inline,
		nil,
		nil,
	}
	return form
}

func (f *Form) AddClass(class string) {
	f.Class = append(f.Class, class)
}
