package model

import (
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/url"
	"regexp"
	"scl/db"
	h "scl/helper"
	"strings"
)

const (
	VALIDATION_ERROR_REQUIRED  = "Required field has empty value."
	VALIDATION_ERROR_EMAIL     = "Wrong email format."
	VALIDATION_ERROR_LINK      = "Wrong link format."
	VALIDATION_ERROR_PASSWORD  = "Wrong password format: The password must contains at least one lowercase, one uppercase letter, and one number."
	VALIDATION_ERROR_SAMEAS    = "The fields do not match."
	VALIDATION_ERROR_REGEXP    = "Wrong input format."
	VALIDATION_ERROR_LENGTH    = "The length of the fields value is eighter not enough or too much."
	VALIDATION_ERROR_EXTENSION = "The file extension is not allowed. Allowed extensions are: %s"
	VALIDATION_ERROR_UNIQUE    = "The database already contains an entry with the same value."

	VALIDATION_FORMAT_EMAIL          = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	VALIDATION_FORMAT_PASSWORD_LOWER = "[a-z]"
	VALIDATION_FORMAT_PASSWORD_UPPER = "[A-Z]"
	VALIDATION_FORMAT_PASSWORD_DIGIT = "[0-9]"
)

type Validator struct {
	ctx    *fasthttp.RequestCtx
	Fields map[string]map[string]interface{}
	Errors map[string]error
	cKey   string
	values map[string]interface{}
}

func (v *Validator) Init(ctx *fasthttp.RequestCtx) {
	v.ctx = ctx
	v.values = make(map[string]interface{})
	v.Fields = make(map[string]map[string]interface{})
	v.Errors = make(map[string]error)
}

func (v *Validator) SetFields(fields map[string]map[string]interface{}) {
	v.Fields = fields
}

func (v *Validator) AddField(key string, properties map[string]interface{}) (bool, error) {
	_, ok := v.Fields[key]
	if ok {
		return false, errors.New(fmt.Sprintf("The key `%v` has already been added to the validator."))
	} else {
		v.Fields[key] = properties
	}

	return true, nil
}

func (v *Validator) Validate() (bool, map[string]error) {
	h.PrintlnIf("Validation start", h.GetConfig().Mode.Debug)
	for key, properties := range v.Fields {
		v.cKey = key
		if strings.Index(key, `%`) != -1 {
			v.ctx.PostArgs().VisitAll(v.iterateRegexp)
			continue
		}
		succ, err := v.ValidateField(key, properties)
		if !succ {
			v.Errors[key] = err
		}
	}

	return len(v.Errors) == 0, v.Errors
}

func (v *Validator) iterateRegexp(postKey []byte, postValue []byte) {
	var key = string(postKey)
	matched, err := regexp.MatchString(strings.Replace(v.cKey, "%", ".*", -1), key)
	h.Error(err, "", h.ERROR_LVL_NOTICE)
	if matched {
		succ, err := v.ValidateField(key, v.Fields[v.cKey])
		if !succ {
			_, ok := v.Errors[key]
			if !ok {
				v.Errors[key] = err
			}
		}
	}
}

func (v *Validator) ValidateField(key string, properties map[string]interface{}) (bool, error) {
	h.PrintlnIf(fmt.Sprintf("Validating field %v", key), h.GetConfig().Mode.Debug)
	var err error
	var succ bool
	for rule, options := range properties["roles"].(map[string]interface{}) {
		multiProp, ok := properties["multi"].(bool)
		multi := ok && multiProp
		fieldRoleProps := properties["roles"].(map[string]interface{})
		required, ok := fieldRoleProps["required"].(bool)
		if !ok {
			required = false
		}
		switch rule {
		case "required":
			succ, err = v.ValidateRequired(key, options, multi)
			break
		case "format":
			succ, err = v.ValidateFormat(key, options, required)
			break
		case "sameas":
			succ, err = v.ValidateSameAs(key, options, required)
			break
		case "length":
			succ, err = v.ValidateLength(key, options, required)
			break
		case "unique":
			succ, err = v.validateUnique(key, options, required)
			break
		case "extension":
			succ, err = v.ValidateExtension(key, options, required)
			break
		}

		if !succ {
			h.PrintlnIf("Not valid", h.GetConfig().Mode.Debug)
			return false, err
		} else {
			h.PrintlnIf("Valid", h.GetConfig().Mode.Debug)
		}
	}

	return true, nil
}

func (v *Validator) ValidateRequired(key string, option interface{}, isMultiValue bool) (bool, error) {

	if isMultiValue {
		return v.ValidateRequiredMulti(key, option)
	}

	h.PrintlnIf(fmt.Sprintf("Validating required"), h.GetConfig().Mode.Debug)
	if option == false {
		return true, nil
	}

	if v.isEmpty(key) {
		return false, v.GetErrorToType(key, "required", VALIDATION_ERROR_REQUIRED)
	}

	return true, nil
}

func (v Validator) GetErrorToType(key string, typ string, defaultError string) error {
	properties, ok := v.Fields[key]
	if !ok {
		panic(errors.New("Could not find field %s in validator fields."))
	}

	customError, ok := properties["error"]
	if !ok {
		return errors.New(defaultError)
	}

	typeError, ok := customError.(map[string]string)[typ]

	if ok {
		return errors.New(typeError)
	}

	globalError, ok := customError.(map[string]string)["global"]

	if ok {
		return errors.New(globalError)
	}

	return errors.New(defaultError)

}

func (v *Validator) validateUnique(key string, option interface{}, required bool) (bool, error) {
	h.PrintlnIf(fmt.Sprintf("Validating unique"), h.GetConfig().Mode.Debug)
	postVal := v.getValue(key, false)
	empty := v.isEmpty(key)
	if empty && required {
		return true, nil
	}

	uniqOpt := option.(map[string]interface{})
	table := uniqOpt["table"].(string)
	field := uniqOpt["field"].(string)
	current := uniqOpt["current"].(string) //before save

	var strCount = fmt.Sprintf(`SELECT COUNT(id) FROM %v WHERE %v = "%v"`, table, field, postVal)
	h.PrintlnIf(strCount, h.GetConfig().Mode.Debug)
	count, err := db.DbMap.SelectInt(strCount)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	var countMax int64 = 0
	if postVal != "" && current == postVal {
		countMax = 1
	}

	if countMax < count {
		return false, v.GetErrorToType(key, "unique", VALIDATION_ERROR_UNIQUE)
	}

	return true, nil
}

func (v *Validator) ValidateExtension(key string, option interface{}, required bool) (bool, error) {
	h.PrintlnIf(fmt.Sprintf("Validating extensions"), h.GetConfig().Mode.Debug)

	file, err := v.ctx.FormFile(key)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	postVal := file.Filename
	if !required {
		for _, ev := range v.getEmptyValues() {
			if ev == postVal {
				return true, nil
			}
		}
	}

	exts := strings.Split(postVal, ".")
	allowedExts := option.([]string)
	for _, ext := range allowedExts {
		if ext == exts[len(exts)-1] {
			return true, nil
		}
	}

	return false, v.GetErrorToType(key, "unique", fmt.Sprintf(VALIDATION_ERROR_EXTENSION, strings.Join(allowedExts, ",")))
}

func (v *Validator) ValidateRequiredMulti(key string, option interface{}) (bool, error) {
	h.PrintlnIf(fmt.Sprintf("Validating required multi"), h.GetConfig().Mode.Debug)
	if option == false {
		return true, nil
	}

	if v.isEmptyMulti(key) {
		return false, v.GetErrorToType(key, "required", VALIDATION_ERROR_REQUIRED)
	}

	return true, nil
}

func (v *Validator) ValidateFormat(key string, option interface{}, required bool) (bool, error) {
	postVal := v.getValue(key, false).(string)
	empty := v.isEmpty(key)
	if !required && empty {
		return true, nil
	}
	var valid bool
	var err error
	var optionMap = option.(map[string]interface{})
	switch optionMap["type"].(string) {
	case "email":
		valid = v.validateEmail(postVal)
		err = v.GetErrorToType(key, "format", VALIDATION_ERROR_EMAIL)
		break
	case "link":
		valid = v.validateLink(postVal)
		err = v.GetErrorToType(key, "format", VALIDATION_ERROR_LINK)
		break
	case "regexp":
		valid = v.validateRegexp(postVal, optionMap["pattern"].(string))
		err = v.GetErrorToType(key, "format", VALIDATION_ERROR_REGEXP)
		break
	case "password":
		valid = v.validatePassword(postVal)
		err = v.GetErrorToType(key, "format", VALIDATION_ERROR_PASSWORD)
		break
	}

	if valid == true {
		err = nil
	}

	return valid, err
}

func (v *Validator) validateLink(value string) bool {
	h.PrintlnIf(fmt.Sprintf("Validating link"), h.GetConfig().Mode.Debug)
	_, err := url.ParseRequestURI(value)
	return err == nil
}

func (v *Validator) validateEmail(value string) bool {
	h.PrintlnIf(fmt.Sprintf("Validating email"), h.GetConfig().Mode.Debug)
	return v.validateRegexp(value, VALIDATION_FORMAT_EMAIL)
}

func (v *Validator) validateRegexp(value string, pattern string) bool {
	h.PrintlnIf(fmt.Sprintf("Validating regexp %v in %v", pattern, value), h.GetConfig().Mode.Debug)
	re := regexp.MustCompile(pattern)
	return re.MatchString(value)
}

func (v *Validator) validatePassword(value string) bool {
	h.PrintlnIf("Validating password", h.GetConfig().Mode.Debug)
	for _, exp := range []string{VALIDATION_FORMAT_PASSWORD_UPPER, VALIDATION_FORMAT_PASSWORD_LOWER, VALIDATION_FORMAT_PASSWORD_DIGIT} {
		if !v.validateRegexp(value, exp) {
			return false
		}
	}

	return true
}

func (v *Validator) ValidateLength(key string, option interface{}, required bool) (bool, error) {
	minLength, okMin := option.(map[string]interface{})["min"].(int)
	maxLength, okMax := option.(map[string]interface{})["max"].(int)
	h.PrintlnIf(fmt.Sprintf("Validating length %v %v", minLength, maxLength), h.GetConfig().Mode.Debug)
	postVal := v.getValue(key, false).(string)
	empty := v.isEmpty(key)
	if !required && empty {
		return true, nil
	}

	if (okMin && len(postVal) < minLength) || (okMax && len(postVal) > maxLength) {
		return false, v.GetErrorToType(key, "length", VALIDATION_ERROR_LENGTH)
	}
	return true, nil
}

func (v *Validator) ValidateSameAs(key string, option interface{}, required bool) (bool, error) {
	h.PrintlnIf(fmt.Sprintf("Validating same as"), h.GetConfig().Mode.Debug)
	postVal := v.getValue(key, false).(string)
	empty := v.isEmpty(key)
	if !required && empty {
		return true, nil
	}
	sameValue := v.getValue(option.(string), false).(string)
	if postVal != sameValue {
		return false, v.GetErrorToType(key, "sameas", VALIDATION_ERROR_SAMEAS)
	}

	return true, nil
}

func (v *Validator) isEmpty(key string) bool {
	value := v.getValue(key, false).(string)
	h.PrintlnIf(fmt.Sprintf("Empty check -> Value of %v is %v", key, value), h.GetConfig().Mode.Debug)
	for _, ev := range v.getEmptyValues() {
		if ev == value {
			return true
		}
	}

	return false
}

func (v *Validator) isEmptyMulti(key string) bool {
	postVal := v.getValue(key, true).([]string)
	nonEmptyValues := len(postVal)
	for _, mv := range postVal {
		for _, ev := range v.getEmptyValues() {
			if ev == mv {
				nonEmptyValues--
			}
		}
	}

	h.PrintlnIf(fmt.Sprintf("non empty values %v", nonEmptyValues), h.GetConfig().Mode.Debug)
	return nonEmptyValues < 1
}

func (v Validator) getEmptyValues() []interface{} {
	return []interface{}{0, "", nil}
}

func (v *Validator) getValue(key string, multi bool) interface{} {
	value, ok := v.values[key]
	if ok {
		return value.(string)
	}

	h.PrintlnIf(fmt.Sprintf("value of %v has not set yet, getting from request", key), h.GetConfig().Mode.Debug)
	if multi {
		var values []string
		for _, mv := range v.ctx.FormValue(key) {
			mvs := string(mv)
			values = append(values, mvs)
		}
		v.values[key] = values
		return v.values[key]
	}

	var postVal string
	if v.Fields[key]["type"] == "file" {
		file, err := v.ctx.FormFile(key)
		h.Error(err, "", h.ERROR_LVL_ERROR)
		postVal = file.Filename
	} else {
		postVal = string(v.ctx.FormValue(key))
	}

	v.values[key] = postVal
	return v.values[key].(string)
}
