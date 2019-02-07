package helper

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
	return GetUrl(mg.Url, nil, true, "admin")
}

func (mg *MenuGroup) SetIsVisible(visible bool) {
	mg.IsVisible = visible
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
	return GetUrl(mi.Url, nil, true, "admin")
}

func (mi *MenuItem) SetIsVisible(visible bool) {
	mi.IsVisible = visible
}

type Menu struct {
	Menu        []MenuGroup `yml:"menu"`
	LogoutUrl   string
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

	menu.LogoutUrl = GetUrl("user/logout", nil, true, "admin")
	menu.LogoutLabel = "Log out"
	menu.IsLoggedIn = session.IsLoggedIn()

	menu.Title = GetConfig().Og.Title
	menu.Lang = fmt.Sprintf("[%v]", session.GetActiveLang())

	return menu
}

func parseMenu(menu *Menu) (bool, error) {
	dat, err := ioutil.ReadFile(MenuFilePath)
	Error(err, "Menu file reading error", ErrorLvlError)
	if err != nil {
		return false, err
	}

	err = yaml.Unmarshal(dat, menu)
	Error(err, "Yaml reading error", ErrorLvlError)
	if err != nil {
		return false, err
	}

	return true, nil
}
