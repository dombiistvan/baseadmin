package helper

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"os"
	"strings"
)

var Lang *Language

var DefLang string = "hu"

var LangQueryKey string = "lang"

func InitLanguage() {
	PrintlnIf("Initializing translator", GetConfig().Mode.Debug)
	Lang = &Language{}
	Lang.Init()
	PrintlnIf("Translator initialization done", GetConfig().Mode.Debug)
}

type Language struct {
	storage   map[string]map[string]string
	available []string
}

func (l *Language) GetStorage() map[string]map[string]string {
	return l.storage
}

func (l *Language) GetAvailableLanguageCodes() []string {
	return l.available
}

func (l *Language) Init() {
	l.storage = make(map[string]map[string]string)
	var path string = "./resource/language"
	dir, err := os.Open(path)
	Error(err, "", ERROR_LVL_ERROR)
	files, err := dir.Readdir(0)
	Error(err, "", ERROR_LVL_ERROR)

	for _, f := range files {
		if !f.IsDir() {
			var parts []string = strings.Split(f.Name(), ".")
			if parts[len(parts)-1] == "json" {
				var mapKey string = strings.Replace(f.Name(), ".json", "", -1)
				if Contains(GetConfig().Language.Allowed, mapKey) {
					PrintlnIf(fmt.Sprintf("Parsing language file %s", f.Name()), GetConfig().Mode.Debug)

					data, err := ioutil.ReadFile(path + "/" + f.Name())
					Error(err, "", ERROR_LVL_ERROR)
					if err != nil {
						continue
					}
					var toData map[string]string
					err = json.Unmarshal(data, &toData)
					l.storage[mapKey] = toData
					Error(err, "", ERROR_LVL_ERROR)
				}
			}
		}
	}

	l.setAvailableLanguages()
}

func (l *Language) setAvailableLanguages() {
	for c, _ := range l.GetStorage() {
		l.available = append(l.available, c)
	}
}

func (l *Language) IsAvailable(lang string) bool {
	_, exists := l.GetStorage()[lang]
	return Contains(l.available, lang) && exists
}

func (l *Language) Trans(txtToTrans string, toLang string) string {
	langMap, ok := l.GetStorage()[toLang]
	if !ok {
		return txtToTrans
	}

	translated, ok := langMap[txtToTrans]
	if !ok {
		return txtToTrans
	}

	return translated
}

func (l *Language) SetLanguage(ctx *fasthttp.RequestCtx, session *Session) {
	var langFromQuery string = string(ctx.FormValue(LangQueryKey))
	if langFromQuery != "" && l.IsAvailable(langFromQuery) {
		PrintlnIf(fmt.Sprintf("Changing language to %s", langFromQuery), GetConfig().Mode.Debug)
		session.SetActiveLang(langFromQuery)
	} else if session.GetActiveLang() == "" || !l.IsAvailable(session.GetActiveLang()) {
		PrintlnIf("Setting default language", GetConfig().Mode.Debug)
		if l.IsAvailable(DefLang) {
			session.SetActiveLang(DefLang)
		} else {
			panic(fmt.Sprintf("The Default language is not allowed: %s", DefLang))
		}
	}
}
