package helper

import (
	"html/template"
	"bytes"
	"path"
	"fmt"
	"time"
	"strconv"
	"strings"
)

func GetScopeTemplateString(filePath string, data interface{}, scope string) string {
	var tplBuffer bytes.Buffer

	filePath = TrimPath(filePath)

	var fileFullPath string = fmt.Sprintf("%v/%v/%v", GetConfig().ViewDir, scope, filePath);
	parseName := strings.Replace(strings.Trim(fileFullPath, "/"), "/", "_", -1);
	t, err := template.New(parseName).Funcs(template.FuncMap{
		"addOne": func(val int) int {
			return  val+1;
		},
	}).ParseFiles("./" + fileFullPath)
	Error(err, "", ERROR_LVL_ERROR)
	err = t.ExecuteTemplate(&tplBuffer, path.Base("./"+fileFullPath), data);
	Error(err, "", ERROR_LVL_ERROR)

	return tplBuffer.String();
}

func GetTemplateString(templateString string, data interface{}) string {
	var tplTempName string;
	var tplBuffer bytes.Buffer

	tplTempName = fmt.Sprintf("%v", strconv.Itoa(int(time.Now().Unix())), "tpl_tmp_");
	tmpl, err := template.New(tplTempName).Parse(templateString)
	Error(err, "", ERROR_LVL_ERROR)
	err = tmpl.Execute(&tplBuffer, data);
	Error(err, "", ERROR_LVL_ERROR)

	return tplBuffer.String();
}
