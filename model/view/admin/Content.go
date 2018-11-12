package admin

import "html/template"

type Content struct {
	Title      string
	SubTitle   string
	Content    template.HTML
}
