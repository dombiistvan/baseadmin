package helper

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"fmt"
)

var MenuFilePath string = "./resource/menu.yml"

type MenuGroup struct {
	Label      string           `yml:"label"`
	Group      string           `yml:"group"`
	Url        string           `yml:"url"`
	Children   map[int]MenuItem `yml:"children"`
	Icon       string           `yml:"icon"`
	Visibility string           `yml:"visibility"`
	IsVisible  bool             `yml:"-"`
}

func (mg MenuGroup) GetUrl() string {
	return GetUrl(mg.Url, nil,true,"admin");
}

func (mg *MenuGroup) SetIsVisible(visible bool) {
	mg.IsVisible = visible;
}

type MenuItem struct {
	Label      string `yml:"label"`
	Url        string `yml:"url"`
	Type       string `yml:"type"`
	Visibility string `yml:"visibility"`
	Icon       string `yml:"icon"`
	IsVisible  bool   `yml:"-"`
}

func (mi MenuItem) GetUrl() string {
	return GetUrl(mi.Url,  nil,true,"admin");
}

func (mi *MenuItem) SetIsVisible(visible bool) {
	mi.IsVisible = visible;
}

type Menu struct {
	Menu        []MenuGroup `yml:"menu"`
	LogoutUrl   string
	LogoutLabel string
	IsLoggedIn  bool
	Title string
	Lang string
}

func (m *Menu) Init(session *Session) {
	for gi, group := range m.Menu {
		for ii, item := range group.Children {
			item.SetIsVisible(CanAccess(item.Visibility, session))
			group.Children[ii] = item;
		}
		group.SetIsVisible(CanAccess(group.Visibility, session));
		m.Menu[gi] = group;
	}
	m.appendLang(session);
}

func (m *Menu) appendLang(session *Session) {
	langGroup := MenuGroup{
		"Active Store",
		"language",
		"",
		nil,
		"fa fa-link",
		"@",
		session.IsLoggedIn(),
	}

	langGroup.Children = map[int]MenuItem{};
	for _,lc := range Lang.GetAvailableLanguageCodes() {
		langGroup.Children[len(langGroup.Children)] = MenuItem{
			lc,
				GetUrl("user/switchlanguage",[]string{lc},false,"admin"),
			"",
			"@",
			"fa fa-flag",
			session.IsLoggedIn(),
		};
	}

	m.Menu = append(m.Menu, langGroup);
}

func GetMenu(session *Session) Menu {
	var menu Menu;
	succ, err := parseMenu(&menu);
	if (nil != err || !succ) {
		Error(err, "", ERROR_LVL_ERROR);
	}

	if (succ) {
		menu.Init(session);
	}

	menu.LogoutUrl = GetUrl("user/logout",  nil,true,"admin");
	menu.LogoutLabel = "Log out";
	menu.IsLoggedIn = session.IsLoggedIn();

	menu.Title = GetConfig().Og.Title;
	menu.Lang = fmt.Sprintf("[%v]",session.GetActiveLang());

	return menu;
}

func parseMenu(menu *Menu) (bool, error) {
	dat, err := ioutil.ReadFile(MenuFilePath);
	Error(err, "Menu file reading error", ERROR_LVL_ERROR);
	if (err != nil) {
		return false, err;
	}

	err = yaml.Unmarshal(dat, menu)
	Error(err, "Yaml reading error", ERROR_LVL_ERROR);
	if (err != nil) {
		return false, err;
	}

	return true, nil;
}
