package helper

import (
	"github.com/valyala/fasthttp"
	"strings"
	"html"
	"fmt"
	"os"
	"time"
	"log"
	"io"
)

/**
 * Return the pramIdx-nth param from request
 * For example /admin/user/edit/3 request GetParamFromCtxPath(ctx,3) will return "3"
 * For example /admin/user/edit/3/redirect/urlencodedstring request GetParamFromCtxPath(ctx,5) will return "urlencodedstring"
 */
func GetParamFromCtxPath(ctx *fasthttp.RequestCtx, paramIdx int64, defaultValue string) string {
	var pathArr []string = strings.Split(TrimPath(string(ctx.Path())), "/");
	if(len(pathArr)<int(paramIdx+1)){
		return defaultValue;
	}
	if(pathArr[paramIdx] == ""){
		return defaultValue;
	}
	return html.EscapeString(pathArr[paramIdx]);
}

func FormValue(ctx *fasthttp.RequestCtx, key string) string {
	return html.EscapeString(string(ctx.FormValue(key)));
}

func PostValue(ctx *fasthttp.RequestCtx, key string) string {
	return html.EscapeString(string(ctx.PostArgs().Peek(key)));
}

func GetUrl(controllerAction string, params []string, withScope bool, scope string) string {
	var scopeReplace string = "";

	if(withScope){
		scopeReplace = "/" + scope;
	}

	var paramStr = "";
	if(len(params)>0){
		paramStr = "/"+strings.Join(params,"/");
	}

	return fmt.Sprintf("%v/%v%v", scopeReplace, TrimPath(controllerAction),paramStr);
}

func GetFormData(ctx *fasthttp.RequestCtx, field string, multi bool) interface{} {
	if (!multi) {
		return string(ctx.FormValue(field));
	} else {
		retVal := []string{}
		for _, val := range ctx.PostArgs().PeekMulti(field) {
			retVal = append(retVal, string(val));
		}
		return retVal;
	}
}

func SetLog() *os.File{
	path := "./system/";
	err := os.MkdirAll(path,0775);
	Error(err,"",ERROR_LVL_WARNING);
	f, err := os.OpenFile(fmt.Sprintf("%v/%v.log",path,time.Now().Format("06_01_02__15")), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	Error(err,"",ERROR_LVL_WARNING);
	mw := io.MultiWriter(os.Stdout, f)

	log.SetOutput(mw)

	return f;
}
