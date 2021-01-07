package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var MenuFilePath string = "./resource/menu.json"

type MenuGroup struct {
	Label      string     `json:"label"`
	Group      string     `json:"group"`
	URL        string     `json:"url"`
	Children   []MenuItem `json:"children"`
	Icon       string     `json:"icon"`
	Visibility string     `json:"visibility"`
	IsVisible  bool       `json:"-"`
}

func (mg MenuGroup) GetURL() string {
	return GetURL(mg.URL, nil, true, "admin")
}

func (mg *MenuGroup) SetIsVisible(visible bool) {
	mg.IsVisible = visible
}

type MenuItem struct {
	Label      string `json:"label"`
	URL        string `json:"url"`
	Type       string `json:"type"`
	Visibility string `json:"visibility"`
	Icon       string `json:"icon"`
	IsVisible  bool   `json:"-"`
}

func (mi MenuItem) GetURL() string {
	return GetURL(mi.URL, nil, true, "admin")
}

func (mi *MenuItem) SetIsVisible(visible bool) {
	mi.IsVisible = visible
}

type Menu struct {
	Menu        []MenuGroup `json:"menu"`
	LogoutURL   string
	LogoutLabel string
	IsLoggedIn  bool
	Title       string
	Lang        string
}

func (m *Menu) Init(session *Session) {
	for gi, group := range m.Menu {
		for ii, item := range group.Children {
			item.SetIsVisible(CanAccess(item.Visibility, session))
			group.Children[ii] = item
		}
		group.SetIsVisible(CanAccess(group.Visibility, session))
		m.Menu[gi] = group
	}
}

func (m *Menu) AddMenuGroup(group MenuGroup) {
	m.Menu = append(m.Menu, group)
}

func GetMenu(session *Session) Menu {
	var menu Menu
	succ, err := parseMenu(&menu)
	if nil != err || !succ {
		Error(err, "", ErrorLvlError)
	}

	if succ {
		menu.Init(session)
	}

	menu.LogoutURL = GetURL("user/logout", nil, true, "admin")
	menu.LogoutLabel = "Log out"
	menu.IsLoggedIn = session.IsLoggedIn()

	menu.Title = GetConfig().OpenGraph.Title
	menu.Lang = fmt.Sprintf("[%v]", session.GetActiveLang())

	return menu
}

func parseMenu(menu *Menu) (bool, error) {
	dat, err := ioutil.ReadFile(MenuFilePath)
	Error(err, "Menu file reading error", ErrorLvlError)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(dat, menu)
	Error(err, "Yaml reading error", ErrorLvlError)
	if err != nil {
		return false, err
	}

	return true, nil
}
