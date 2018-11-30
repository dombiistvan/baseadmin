package model

import (
	h "baseadmin/helper"
	"fmt"
	"strings"
)

const FIELDSET_TEMPLATE = `<div class="%class%" style="margin-bottom:10px;">
	%formgroup%
</div>`

type Fieldset struct {
	Identifier string
	Elements   []FormElement
	Colmap     map[string]string //fe.: Colmap = map[string]string{"lg":"6","md":"6","sm":"12","xs":"12"}
}

func (fs Fieldset) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering fieldset", h.GetConfig().Mode.Debug)
	var foutput string = FIELDSET_TEMPLATE
	var classes []string

	classes = append(classes, "fieldset")
	for size, col := range fs.Colmap {
		classes = append(classes, fmt.Sprintf("col-%v-%v", size, col))
	}

	var eoutput []string

	for _, e := range fs.Elements {
		eoutput = append(eoutput, e.Render(errs))
	}

	foutput = h.Replace(foutput, []string{"%class%"}, []string{strings.Join(classes, " ")})
	foutput = h.Replace(foutput, []string{"%formgroup%"}, []string{strings.Join(eoutput, "\n")})

	return foutput
}

func (fs *Fieldset) AddElement(element FormElement) {
	fs.Elements = append(fs.Elements, element)
}
