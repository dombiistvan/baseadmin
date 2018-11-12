package FElement

import (
	h "base/helper"
	"fmt"
	"strings"
)

const (
	INPUT_FILE_TEMPLATE = `%label%
	%display%
	<input %attrs% />
%note%`
)

type InputFile struct {
	Label    string
	Name     string
	Id       string
	Value    string
	Disabled bool
	Note     string
	Dir      string
	Type     string
}

func (f InputFile) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering file", h.GetConfig().Mode.Debug)
	var replaces map[string]string = make(map[string]string)
	output := INPUT_FILE_TEMPLATE
	var inpErrors []error
	inpError, contains := errs[f.Name]
	if contains {
		inpErrors = append(inpErrors, inpError)
	}

	replaces["%label%"] = ""
	if f.Label != "" {
		replaces["%forattr%"] = ""
		if f.Id != "" {
			replaces["%forattr%"] = h.Replace(LABEL_FOR_TEMPLATE, []string{"%for%"}, []string{f.Id})
		}
		replaces["%label%"] = h.Replace(LABEL_TEMPLATE, []string{"%forattr%", "%label%"}, []string{replaces["%forattr%"], f.Label})
	}

	replaces["%note%"] = ""
	if f.Note != "" {
		replaces["%note%"] = h.Replace(NOTE_TEMPLATE, []string{"%note%"}, []string{f.Note})
	}

	replaces["%attrs%"] = ""
	var attr []string
	attr = append(attr, h.HtmlAttribute("type", "file"))
	if f.Name != "" {
		attr = append(attr, h.HtmlAttribute("name", f.Name))
	}
	if f.Id != "" {
		attr = append(attr, h.HtmlAttribute("id", f.Id))
	}
	if f.Disabled == true {
		attr = append(attr, h.HtmlAttribute("disabled", "disabled"))
	}

	replaces["%attrs%"] = strings.Join(attr, " ")

	var path []string

	if f.Dir != "" {
		path = append(path, strings.Trim(f.Dir, "/"))
	}

	if f.Value != "" {
		path = append(path, strings.TrimLeft(f.Value, "/"))
	}

	filePath := "/" + strings.Join(path, "/")
	if f.Type == "image" && f.Value != "" {
		replaces["%display%"] = fmt.Sprintf(`<a style="display: block;float: left;margin: 0 10px 0 0;" href="%v" target="_blank"><img src="%v" width="100" /></a>`, filePath, filePath)
	} else if f.Type == "file" && f.Value != "" {
		replaces["%display%"] = fmt.Sprintf(`<a href="%v" style="display:block;">%v</a>`, filePath, f.Value)
	} else {
		replaces["%display%"] = ""
	}

	for i, v := range replaces {
		output = h.Replace(output, []string{i}, []string{v})
	}

	return GroupRender(output, f.HasPreOrPost(), false, inpErrors, "")
}

func (f InputFile) HasPreOrPost() bool {
	return false
}
