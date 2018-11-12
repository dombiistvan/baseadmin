package model

import (
	h "base/helper"
	"base/model/FElement"
	"fmt"
	"github.com/valyala/fasthttp"
	"html"
	"strings"
)

type SearchParam struct {
	Label   string
	Name    string
	Type    string
	DbField string
	Option  map[string]interface{}
}

func (sp SearchParam) GetInputsByType(ctx *fasthttp.RequestCtx) []FormElement {
	switch sp.Type {
	case "number":
		return sp.GetNumberInputs(ctx)
		break
	case "number_range":
		return sp.GetNumberRangeInputs(ctx)
		break
	case "text":
		return sp.GetTextInput(ctx)
		break
	case "select":
		return sp.GetSelectInput(ctx)
		break
	case "bool":
		tempSp := sp.getBoolTempSp(ctx)
		return tempSp.GetInputsByType(ctx)
		break
	}

	return []FormElement{}
}

func (sp SearchParam) getBoolTempSp(ctx *fasthttp.RequestCtx) SearchParam {
	tempSp := SearchParam{
		sp.Label,
		sp.Name,
		"select",
		sp.DbField,
		map[string]interface{}{
			"options": []map[string]string{
				0: {
					"label": "Anything",
					"value": "",
				},
				1: {
					"label": "No",
					"value": "0",
				},
				2: {
					"label": "Yes",
					"value": "1",
				},
			},
		},
	}

	return tempSp
}

func (sp SearchParam) GetSelectInput(ctx *fasthttp.RequestCtx) []FormElement {
	var elements []FormElement
	inpSelect := FElement.InputSelect{
		sp.Label,
		sp.Name,
		sp.GetSpId(""),
		sp.GetString("class"),
		sp.GetBool("disabled"),
		sp.GetBool("readonly"),
		[]string{},
		false,
		sp.Option["options"].([]map[string]string),
		"",
	}
	var value []string
	if len(ctx.QueryArgs().Peek(inpSelect.Name)) > 0 {
		value = append(value, string(ctx.QueryArgs().Peek(inpSelect.Name)))
	}
	inpSelect.Value = value
	elements = append(elements, inpSelect)
	return elements
}

func (sp SearchParam) GetNumberRangeInputs(ctx *fasthttp.RequestCtx) []FormElement {
	var elements []FormElement
	inpFrom := FElement.InputText{
		fmt.Sprintf("%v from", sp.Label),
		fmt.Sprintf("%v", sp.Name),
		sp.GetSpId("from"),
		sp.GetString("class"),
		sp.GetString("placeholder"),
		sp.GetBool("disabled"),
		sp.GetBool("readonly"),
		"",
		sp.GetString("note"),
		sp.GetString("preStr"),
		sp.GetString("postStr"),
		sp.GetString("preFAIcon"),
		sp.GetString("postFAIcon"),
	}
	if len(ctx.QueryArgs().PeekMulti(inpFrom.Name)) > 0 {
		inpFrom.Value = string(ctx.QueryArgs().PeekMulti(inpFrom.Name)[0])
	}
	elements = append(elements, inpFrom)
	inpTo := FElement.InputText{
		fmt.Sprintf("%v to", sp.Label),
		fmt.Sprintf("%v", sp.Name),
		sp.GetSpId("to"),
		sp.GetString("class"),
		sp.GetString("placeholder"),
		sp.GetBool("disabled"),
		sp.GetBool("readonly"),
		"",
		sp.GetString("note"),
		sp.GetString("preStr"),
		sp.GetString("postStr"),
		sp.GetString("preFAIcon"),
		sp.GetString("postFAIcon"),
	}
	if len(ctx.QueryArgs().PeekMulti(inpTo.Name)) > 0 {
		inpTo.Value = string(ctx.QueryArgs().PeekMulti(inpTo.Name)[1])
	}
	elements = append(elements, inpTo)
	return elements
}

func (sp SearchParam) GetNumberInputs(ctx *fasthttp.RequestCtx) []FormElement {
	var elements []FormElement
	inp := FElement.InputText{
		sp.Label,
		sp.Name,
		sp.GetSpId(""),
		sp.GetString("class"),
		sp.GetString("placeholder"),
		sp.GetBool("disabled"),
		sp.GetBool("readonly"),
		"",
		sp.GetString("note"),
		sp.GetString("preStr"),
		sp.GetString("postStr"),
		sp.GetString("preFAIcon"),
		sp.GetString("postFAIcon"),
	}
	if len(ctx.QueryArgs().Peek(inp.Name)) > 0 {
		inp.Value = string(ctx.QueryArgs().Peek(inp.Name))
	}
	elements = append(elements, inp)
	return elements
}

func (sp SearchParam) GetTextInput(ctx *fasthttp.RequestCtx) []FormElement {
	var elements []FormElement
	inpText := FElement.InputText{
		sp.Label,
		sp.Name,
		sp.GetSpId(""),
		sp.GetString("class"),
		sp.GetString("placeholder"),
		sp.GetBool("disabled"),
		sp.GetBool("readonly"),
		"",
		sp.GetString("note"),
		sp.GetString("preStr"),
		sp.GetString("postStr"),
		sp.GetString("preFAIcon"),
		sp.GetString("postFAIcon"),
	}
	inpText.Value = string(ctx.QueryArgs().Peek(inpText.Name))
	elements = append(elements, inpText)
	return elements
}

func (sp SearchParam) GetSpId(postFix string) string {
	return fmt.Sprintf("search_param_%v_%v_%v", sp.Name, postFix, h.GetTimeNow().Unix())
}

func (sp SearchParam) GetString(key string) string {
	val := h.GetOption(sp.Option, key)
	if val == nil {
		return ""
	}

	return val.(string)
}

func (sp SearchParam) GetStrings(key string) []string {
	val := h.GetOption(sp.Option, key)
	if val == nil {
		return []string{}
	}

	return val.([]string)
}

func (sp SearchParam) GetBool(key string) bool {
	val := h.GetOption(sp.Option, key)
	if val == nil {
		return false
	}

	return val.(bool)
}

func (sp SearchParam) GetSqlPart(ctx *fasthttp.RequestCtx) string {
	switch sp.Type {
	case "number_range":
		if len(ctx.QueryArgs().PeekMulti(sp.Name)) > 0 {
			valueF := string(ctx.QueryArgs().PeekMulti(sp.Name)[0])
			valueT := string(ctx.QueryArgs().PeekMulti(sp.Name)[1])
			var wheres []string
			if valueF != "" {
				wheres = append(wheres, fmt.Sprintf(`%v >= %v`, sp.DbField, html.EscapeString(valueF)))
			}
			if valueT != "" {
				wheres = append(wheres, fmt.Sprintf(`%v <= %v`, sp.DbField, html.EscapeString(valueT)))
			}
			return strings.Join(wheres, " AND ")
		}
		break
	case "number":
		if len(ctx.QueryArgs().Peek(sp.Name)) > 0 {
			value := string(ctx.QueryArgs().Peek(sp.Name))
			return fmt.Sprintf(`%v = %v`, sp.DbField, html.EscapeString(value))
		}
		break
	case "text":
		if len(ctx.QueryArgs().Peek(sp.Name)) > 0 {
			value := string(ctx.QueryArgs().Peek(sp.Name))
			return fmt.Sprintf(`INSTR(%v,"%v") != 0`, sp.DbField, html.EscapeString(value))
		}
	case "bool":
		return sp.getBoolTempSp(ctx).GetSqlPart(ctx)
		break
	case "select":
		if len(ctx.QueryArgs().Peek(sp.Name)) > 0 {
			value := string(ctx.QueryArgs().Peek(sp.Name))
			return fmt.Sprintf(`%v = "%v"`, sp.DbField, html.EscapeString(value))
		}
		break
	default:
		break
	}

	return ""
}
