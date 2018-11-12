package FElement

import (
	h "base/helper"
)

const (
	CHECKBOXGROUP_TEMPLATE = `%label%%inputs%`
)

type CheckboxGroup struct {
	Label    string
	Checkbox []InputCheckbox
	Static   Static
}

func (cg CheckboxGroup) Render(errs map[string]error) string {
	h.PrintlnIf("Rendering checkboxgroup", h.GetConfig().Mode.Debug);
	var replaces map[string]string = make(map[string]string);
	cgoutput := CHECKBOXGROUP_TEMPLATE;
	replaces["%label%"] = "";
	if (cg.Label != "") {
		replaces["%label%"] = h.Replace(LABEL_TEMPLATE, []string{"%forattr%", "%label%"}, []string{"", cg.Label});
	}

	var inpErrors []error;
	replaces["%inputs%"] = "";
	var inputName string;
	for _, v := range cg.Checkbox {

		inpError, contains := errs[v.Name];
		if (contains && inputName != v.Name) {
			inpErrors = append(inpErrors, inpError);
			inputName = v.Name;
		}

		replaces["%inputs%"] += "\n" + v.Render(nil);
	}

	replaces["%inputs%"] += "\n" + cg.Static.Render(nil);

	for i, v := range replaces {
		cgoutput = h.Replace(cgoutput, []string{i}, []string{v});
	}

	return GroupRender(cgoutput, cg.HasPreOrPost(), false, inpErrors, "");
}

func (cg CheckboxGroup) HasPreOrPost() bool {
	return false;
}
