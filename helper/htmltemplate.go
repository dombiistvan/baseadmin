package helper

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
	"strconv"
	"strings"
	"time"
)

func GetScopeTemplateString(filePath string, data interface{}, scope string) string {
	var tplBuffer bytes.Buffer

	filePath = TrimPath(filePath)

	var fileFullPath string = fmt.Sprintf("%v/%v/%v", GetConfig().ViewDir, scope, filePath)
	parseName := strings.Replace(strings.Trim(fileFullPath, "/"), "/", "_", -1)
	t, err := template.New(parseName).Funcs(template.FuncMap{
		"addOne": func(val int) int {
			return val + 1
		},
	}).ParseFiles("./" + fileFullPath)
	Error(err, "", ErrorLvlError)
	err = t.ExecuteTemplate(&tplBuffer, path.Base("./"+fileFullPath), data)
	Error(err, "", ErrorLvlError)

	return tplBuffer.String()
}

func GetTemplateString(templateString string, data interface{}) string {
	var tplTempName string
	var tplBuffer bytes.Buffer

	tplTempName = fmt.Sprintf("%v", strconv.Itoa(int(time.Now().Unix())), "tpl_tmp_")
	tmpl, err := template.New(tplTempName).Parse(templateString)
	Error(err, "", ErrorLvlError)
	err = tmpl.Execute(&tplBuffer, data)
	Error(err, "", ErrorLvlError)

	return tplBuffer.String()
}
