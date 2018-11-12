package FElement

import (
	h "base/helper"
	"strings"
	"fmt"
)

const (
	LABEL_FOR_TEMPLATE = `for="%for%"`
	LABEL_TEMPLATE = `<label %forattr%>%label%</label>`
	NOTE_TEMPLATE = `<p class="help-block">%note%</p>`
	FORM_GROUP_TEMPLATE = `%grpstart%%formelement%%grpend%`;
	FORM_GROUPSTART_TEMPLATE = `<div class="form-group %class%">`;
	FORM_GROUPEND_TEMPLATE = `%error%</div>`;
	FORM_ELEMENT_ERROR_TEMPLATE = `<div class="alert alert-danger" style="line-height:20px;padding:0 5px;margin-top:5px">%error%</div>`
);

func GroupRender(elementsoutput string, InputHasPreOrPost bool, InlineDisplay bool, errs []error, pull string) string {
	h.PrintlnIf("Rendering input group", h.GetConfig().Mode.Debug);
	goutput := FORM_GROUP_TEMPLATE;
	var gClassArr []string;
	if(InputHasPreOrPost){
		gClassArr = append(gClassArr, "input-group");
	}
	if(len(errs)>0){
		gClassArr = append(gClassArr, "has-error");
	}

	if(pull != ""){
		gClassArr = append(gClassArr, fmt.Sprintf("pull-%v",pull));
	}

	gclass := strings.Join(gClassArr," ");

	gstart := "";
	gend := "%error%";
	if(!InlineDisplay){
		gstart = FORM_GROUPSTART_TEMPLATE;
		gend = FORM_GROUPEND_TEMPLATE;
	}

	var stErrs []string;
	for _,err := range errs{
		stErrs = append(stErrs, h.Replace(FORM_ELEMENT_ERROR_TEMPLATE,[]string{"%error%"},[]string{err.Error()}))
	}

	gend = h.Replace(gend,[]string{"%error%"},[]string{strings.Join(stErrs,"\n")})

	goutput = h.Replace(goutput, []string{"%grpstart%","%grpend%","%class%","%formelement%"}, []string{gstart,gend,gclass,elementsoutput});

	return goutput;
}
