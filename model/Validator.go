package model

import (
	"baseadmin/db"
	h "baseadmin/helper"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/url"
	"regexp"
	"strings"
)

const (
	ValidationErrorRequired  = "Required field has empty value."
	ValidationErrorEmail     = "Wrong email format."
	ValidationErrorLink      = "Wrong link format."
	ValidationErrorPassword  = "Wrong password format: The password must contains at least one lowercase, one uppercase letter, and one number."
	ValidationErrorSameAs    = "The fields do not match."
	ValidationErrorRegexp    = "Wrong input format."
	ValidationErrorLength    = "The length of the fields value is eighter not enough or too much."
	ValidationErrorCount     = "The amount of values of field is eighter not enough or too much."
	ValidationErrorExtension = "The file extension is not allowed. Allowed extensions are: %s"
	ValidationErrorUnique    = "The database already contains an entry with the same value."

	ValidationFormatEmail         = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	ValidationFormatpasswordLower = "[a-z]"
	ValidationFormatpasswordUpper = "[A-Z]"
	ValidationFormatpasswordDigit = "[0-9]"
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

func (v *Validator) initField(formKey string) {
	_, ok := v.Fields[formKey]
	if !ok {
		v.Fields[formKey] = map[string]interface{}{}
	}
	_, ok = v.Fields[formKey]["roles"]
	if !ok {
		v.Fields[formKey]["roles"] = map[string]interface{}{}
	}
	_, ok = v.Fields[formKey]["error"]
	if !ok {
		v.Fields[formKey]["error"] = map[string]string{}
	}
}

func (v *Validator) addCustomError(formKey string, errType string, err string) {
	v.Fields[formKey]["error"].(map[string]string)[errType] = err
}

func (v *Validator) AddEmailValidator(formKey string, customError string) {
	v.initField(formKey)
	v.Fields[formKey]["roles"].(map[string]interface{})["format"] = map[string]interface{}{"type": "email"}
	if customError != "" {
		v.addCustomError(formKey, "format", customError)
	}
}

func (v *Validator) AddUrlValidator(formKey string, multi bool, customError string) {
	v.initField(formKey)
	v.Fields[formKey]["roles"].(map[string]interface{})["format"] = map[string]interface{}{"type": "url"}
	if multi {
		v.Fields[formKey]["multi"] = true
	}
	if customError != "" {
		v.addCustomError(formKey, "format", customError)
	}
}

func (v *Validator) AddRegexpValidator(formKey string, pattern string, customError string) {
	v.initField(formKey)
	v.Fields[formKey]["roles"].(map[string]interface{})["format"] = map[string]interface{}{"type": "regexp", "pattern": pattern}
	if customError != "" {
		v.addCustomError(formKey, "format", customError)
	}
}

func (v *Validator) AddPasswordValidator(formKey string, customError string) {
	v.initField(formKey)
	v.Fields[formKey]["roles"].(map[string]interface{})["format"] = map[string]interface{}{"type": "password"}
	if customError != "" {
		v.addCustomError(formKey, "format", customError)
	}
}

func (v *Validator) AddSameasValidator(formKey string, sameFormKey string, customError string) {
	v.initField(formKey)
	v.Fields[formKey]["roles"].(map[string]interface{})["sameas"] = sameFormKey
	if customError != "" {
		v.addCustomError(formKey, "sameas", customError)
	}
}

func (v *Validator) AddRequiredValidator(formKey string, customError string, multi bool) {
	v.initField(formKey)
	v.Fields[formKey]["roles"].(map[string]interface{})["required"] = true
	v.Fields[formKey]["multi"] = multi
	if customError != "" {
		v.addCustomError(formKey, "required", customError)
	}
}

func (v *Validator) AddLengthValidator(formKey string, minLength int, maxLength int, customError string) {
	v.initField(formKey)
	v.Fields[formKey]["roles"].(map[string]interface{})["length"] = map[string]interface{}{}
	if minLength != 0 {
		v.Fields[formKey]["roles"].(map[string]interface{})["length"].(map[string]interface{})["min"] = minLength
	}
	if maxLength != 0 {
		v.Fields[formKey]["roles"].(map[string]interface{})["length"].(map[string]interface{})["max"] = maxLength
	}
	if customError != "" {
		v.addCustomError(formKey, "length", customError)
	}
}

func (v *Validator) AddCountValidator(formKey string, minCount int, maxCount int, customError string) {
	v.initField(formKey)
	v.Fields[formKey]["roles"].(map[string]interface{})["count"] = map[string]interface{}{}
	if minCount != 0 {
		v.Fields[formKey]["roles"].(map[string]interface{})["count"].(map[string]interface{})["min"] = minCount
	}
	if maxCount != 0 {
		v.Fields[formKey]["roles"].(map[string]interface{})["count"].(map[string]interface{})["max"] = maxCount
	}
	if customError != "" {
		v.addCustomError(formKey, "count", customError)
	}
}

func (v *Validator) AddUniqueValidator(formKey string, table string, field string, current interface{}, customError string) {
	/*
		"roles":map[string]interface{}{
			"unique": map[string]interface{}{
				"table":   "page",
				"field":   "url_key",
				"current": Page.UrlKey,
			},
		},
	*/
	v.initField(formKey)
	v.Fields[formKey]["roles"].(map[string]interface{})["unique"] = map[string]interface{}{}
	v.Fields[formKey]["roles"].(map[string]interface{})["unique"] = map[string]interface{}{
		"table":   table,
		"field":   field,
		"current": current,
	}
	if customError != "" {
		v.addCustomError(formKey, "unique", customError)
	}
}

func (v *Validator) AddField(key string, properties map[string]interface{}) (bool, error) {
	_, ok := v.Fields[key]
	if ok {
		return false, errors.New(fmt.Sprintf("The key `%v` has already been added to the validator.", key))
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
	h.Error(err, "", h.ErrLvlNotice)
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
			succ, err = v.ValidateFormat(key, options, required, multi)
			break
		case "sameas":
			succ, err = v.ValidateSameAs(key, options, required, multi)
			break
		case "length":
			succ, err = v.ValidateLength(key, options, required, multi)
			break
		case "count":
			succ, err = v.ValidateCount(key, options, required, multi)
			break
		case "unique":
			succ, err = v.ValidateUnique(key, options, required, multi)
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
		return false, v.GetErrorToType(key, "required", ValidationErrorRequired)
	}

	return true, nil
}

func (v Validator) GetErrorToType(key string, typ string, defaultError string) error {
	properties, ok := v.Fields[key]
	if !ok {
		panic(errors.New("could not find field %s in validator fields"))
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

func (v *Validator) ValidateUnique(key string, option interface{}, required bool, multi bool) (bool, error) {
	h.PrintlnIf(fmt.Sprintf("Validating unique"), h.GetConfig().Mode.Debug)
	postVals := v.getValue(key, multi)
	var empty bool
	if multi {
		empty = v.isEmptyMulti(key)
	} else {
		empty = v.isEmpty(key)
	}

	if empty && required {
		return true, nil
	}

	uniqOpt := option.(map[string]interface{})
	table := uniqOpt["table"].(string)
	field := uniqOpt["field"].(string)
	current, cok := uniqOpt["current"] // before save

	for _, val := range postVals.([]string) {
		var countMax int = 0
		if val != "" && cok && current.(string) == val {
			countMax = 1
		}
		if !v.ValidateUniqueDb(table, field, countMax, val) {
			return false, v.GetErrorToType(key, "unique", ValidationErrorUnique)
		}
	}

	return true, nil
}

func (v *Validator) ValidateUniqueDb(table string, field string, maxCount int, value interface{}) bool {
	var strCount = fmt.Sprintf(`SELECT COUNT(id) FROM %s WHERE %s = ?`, table, field)
	h.PrintlnIf(strCount, h.GetConfig().Mode.Debug)
	count, err := db.DbMap.SelectInt(strCount, value)
	h.Error(err, "", h.ErrorLvlError)
	if maxCount < int(count) {
		return false
	}

	return true
}

func (v *Validator) ValidateExtension(key string, option interface{}, required bool) (bool, error) {
	h.PrintlnIf(fmt.Sprintf("Validating extensions"), h.GetConfig().Mode.Debug)

	file, err := v.ctx.FormFile(key)
	h.Error(err, "", h.ErrLvlWarning)
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

	return false, v.GetErrorToType(key, "extension", fmt.Sprintf(ValidationErrorExtension, strings.Join(allowedExts, ",")))
}

func (v *Validator) ValidateRequiredMulti(key string, option interface{}) (bool, error) {
	h.PrintlnIf(fmt.Sprintf("Validating required multi"), h.GetConfig().Mode.Debug)
	if option == false {
		return true, nil
	}

	if v.isEmptyMulti(key) {
		return false, v.GetErrorToType(key, "required", ValidationErrorRequired)
	}

	return true, nil
}

func (v *Validator) ValidateFormat(key string, option interface{}, required bool, multi bool) (bool, error) {
	var postVals []string
	var valid bool
	var err error
	var optionMap = option.(map[string]interface{})

	postVals = v.getValue(key, multi).([]string)

	if !required {
		return true, nil
	} else if len(postVals) == 0 {
		return false, v.GetErrorToType(key, "required", ValidationErrorRequired)
	}
	for _, postVal := range postVals {
		switch optionMap["type"].(string) {
		case "email":
			valid = v.ValidateEmail(postVal)
			err = v.GetErrorToType(key, "format", ValidationErrorEmail)
			break
		case "link":
		case "url":
			valid = v.ValidateLink(postVal)
			err = v.GetErrorToType(key, "format", ValidationErrorLink)
			break
		case "regexp":
			valid = v.ValidateRegexp(postVal, optionMap["pattern"].(string))
			err = v.GetErrorToType(key, "format", ValidationErrorRegexp)
			break
		case "password":
			valid = v.ValidatePassword(postVal)
			err = v.GetErrorToType(key, "format", ValidationErrorPassword)
			break
		}
		if !valid {
			break
		}
	}

	return valid, err
}

func (v *Validator) ValidateLink(value string) bool {
	h.PrintlnIf(fmt.Sprintf("Validating link %s", value), h.GetConfig().Mode.Debug)
	_, err := url.ParseRequestURI(value)
	return err == nil
}

func (v *Validator) ValidateEmail(value string) bool {
	h.PrintlnIf(fmt.Sprintf("Validating email %s", value), h.GetConfig().Mode.Debug)
	return v.ValidateRegexp(value, ValidationFormatEmail)
}

func (v *Validator) ValidateRegexp(value string, pattern string) bool {
	h.PrintlnIf(fmt.Sprintf("Validating regexp %v in %v", pattern, value), h.GetConfig().Mode.Debug)
	re := regexp.MustCompile(pattern)
	return re.MatchString(value)
}

func (v *Validator) ValidatePassword(value string) bool {
	h.PrintlnIf("Validating password", h.GetConfig().Mode.Debug)
	for _, exp := range []string{ValidationFormatpasswordUpper, ValidationFormatpasswordLower, ValidationFormatpasswordDigit} {
		if !v.ValidateRegexp(value, exp) {
			return false
		}
	}

	return true
}

func (v *Validator) ValidateLength(key string, option interface{}, required bool, multi bool) (bool, error) {
	minLength, okMin := option.(map[string]interface{})["min"].(int)
	maxLength, okMax := option.(map[string]interface{})["max"].(int)
	h.PrintlnIf(fmt.Sprintf("Validating length %v %v", minLength, maxLength), h.GetConfig().Mode.Debug)
	postVals := v.getValue(key, multi).([]string)
	var empty bool
	if multi {
		empty = v.isEmptyMulti(key)
	} else {
		empty = v.isEmpty(key)
	}

	if !required && empty {
		return true, nil
	}

	for _, postVal := range postVals {
		if (okMin && len(postVal) < minLength) || (okMax && len(postVal) > maxLength) {
			return false, v.GetErrorToType(key, "length", ValidationErrorLength)
		}
	}
	return true, nil
}

func (v *Validator) ValidateCount(key string, option interface{}, required bool, multi bool) (bool, error) {
	minCount, okMin := option.(map[string]interface{})["min"].(int)
	maxCount, okMax := option.(map[string]interface{})["max"].(int)
	h.PrintlnIf(fmt.Sprintf("Validating count %v %v", minCount, maxCount), h.GetConfig().Mode.Debug)
	postVals := v.getValue(key, multi).([]string)
	var empty bool
	if multi {
		empty = v.isEmptyMulti(key)
	} else {
		empty = v.isEmpty(key)
	}

	if !required && empty {
		return true, nil
	}

	if (okMin && len(postVals) < minCount) || (okMax && len(postVals) > maxCount) {
		return false, v.GetErrorToType(key, "count", ValidationErrorCount)
	}

	return true, nil
}

func (v *Validator) ValidateSameAs(key string, option interface{}, required bool, multi bool) (bool, error) {
	h.PrintlnIf(fmt.Sprintf("Validating same as"), h.GetConfig().Mode.Debug)
	var empty bool
	var postVals []string = v.getValue(key, multi).([]string)
	if multi {
		empty = v.isEmptyMulti(key)
	} else {
		empty = v.isEmpty(key)
	}

	if !required && empty {
		return true, nil
	}

	for i, postVal := range postVals {
		sameValue := v.getValue(option.(string), multi).([]string)[i]
		if postVal != sameValue {
			return false, v.GetErrorToType(key, "sameas", ValidationErrorSameAs)
		}
	}

	return true, nil
}

func (v *Validator) isEmpty(key string) bool {
	value := v.getValue(key, false).([]string)
	if len(value) == 0 {
		return true
	}
	h.PrintlnIf(fmt.Sprintf("Empty check -> Value of %v is %v", key, value), h.GetConfig().Mode.Debug)
	for _, ev := range v.getEmptyValues() {
		if ev == value[0] {
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
		return value
	}

	if v.Fields[key]["type"] == "file" {
		file, err := v.ctx.FormFile(key)
		h.Error(err, "", h.ErrLvlWarning)
		v.values[key] = []string{file.Filename}
	} else {
		h.PrintlnIf(fmt.Sprintf("value of %v has not set yet, getting from request", key), h.GetConfig().Mode.Debug)
		var values []string
		for _, mv := range v.ctx.PostArgs().PeekMulti(key) {
			mvs := string(mv)
			if mvs != "" {
				values = append(values, mvs)
			}
		}
		v.values[key] = values
	}

	return v.values[key]
}
