package view

import (
	"baseadmin/helper"
	"errors"
	"fmt"
	"html/template"
	"strings"
)

type Page struct {
	Redirected  bool
	Scope       string
	Title       string
	Meta        []map[string]string
	Scripts     map[string][]string
	StyleSheet  []string
	Wrapper     []map[string]interface{}
	Content     []string
	ContentType string
	Layout      string
}

func (p Page) Instantiates() *Page {
	viewpage := Page{}
	return &viewpage
}

func (p *Page) AddScript(path string, container string, ifhtml string) {
	ifstart, ifend := "", ""
	if ifhtml != "" {
		if strings.Index(ifhtml, "<!--[") != 0 {
			ifhtml = fmt.Sprintf("<!--[if %v]>", ifhtml)
		}
		ifstart = ifhtml + "\n"
		ifend = "\n<![endif]-->"
	}

	if container != "header" && container != "footer" {
		helper.Error(errors.New("Script container is not eligible, will not display in pageload"), "", helper.ERROR_LVL_ERROR)
	}

	_, ok := p.Scripts[container]
	if !ok {
		p.Scripts = make(map[string][]string)
	}

	p.Scripts[container] = append(p.Scripts[container], fmt.Sprintf("%v\t<script type=\"text/javascript\" src=\"%v\"></script>%v", ifstart, path, ifend))
}

func (p *Page) AddCss(path string) {
	helper.PrintlnIf(fmt.Sprintf("Adding css %v", path), helper.GetConfig().Mode.Debug)
	p.StyleSheet = append(p.StyleSheet, path)
}

func (p *Page) AddMeta(props map[string]string) {
	p.Meta = append(p.Meta, props)
}

func (p *Page) GetScriptHtml(container string) template.HTML {
	var scripts template.HTML
	containerScripts, ok := p.Scripts[container]
	if !ok {
		return ""
	}
	for _, script := range containerScripts {
		scripts += template.HTML(script) + "\n"
	}
	return scripts
}

func (p *Page) AddAdminScripts() {
	p.AddScript("/assets/js/html5shiv.js", "header", "lt IE 9")
	p.AddScript("/assets/js/respond.js", "header", "lt IE 9")
	p.AddScript("/vendor/jquery/jquery.min.js", "header", "")
	p.AddScript("/vendor/bootstrap/js/bootstrap.min.js", "header", "")
	p.AddScript("/vendor/metisMenu/metisMenu.min.js", "header", "")
	p.AddScript("/assets/dist/js/sb-admin-2.js", "header", "")
	p.AddScript("/assets/summernote/summernote.js", "header", "")
	p.AddScript("/assets/dist/js/adminfunctions.js", "header", "")
}

func (p *Page) AddAdminStylesheets() {
	cssToAdd := []string{
		"/vendor/bootstrap/css/bootstrap.min.css",
		"/vendor/metisMenu/metisMenu.min.css",
		"/assets/dist/css/sb-admin-2.css",
		"/vendor/morrisjs/morris.css",
		"/vendor/font-awesome/css/font-awesome.min.css",
		"/assets/summernote/summernote.css",
		"/assets/css/admin.css",
	}
	for _, css := range cssToAdd {
		p.AddCss(css)
	}
}

func (p *Page) AddDefaultMetaData() {
	p.AddMeta(map[string]string{"charset": "utf-8"})
	p.AddMeta(map[string]string{"http-equiv": "X-UA-Compatible", "content": "IE=edge"})
	p.AddMeta(map[string]string{"name": "viewport", "content": "width=device-width, initial-scale=1"})
}

func (p *Page) AddOgMetaData() {
	p.AddMeta(map[string]string{"name": "og:url", "content": helper.GetConfig().Og.Url})
	p.AddMeta(map[string]string{"name": "og:type", "content": helper.GetConfig().Og.Type})
	p.AddMeta(map[string]string{"name": "og:title", "content": helper.GetConfig().Og.Title})
	p.AddMeta(map[string]string{"name": "og:description", "content": helper.GetConfig().Og.Description})
	p.AddMeta(map[string]string{"name": "og:image", "content": helper.GetConfig().Og.Url + "/" + strings.TrimLeft(helper.GetConfig().Og.Image, "/")})
}

func (p Page) GetMetaTags() template.HTML {
	metaHTML := []string{}
	for _, attrs := range p.Meta {
		metaTag := "\t<meta "
		for key, value := range attrs {
			metaTag += fmt.Sprintf(` %s="%s"`, key, value)
		}
		metaTag += " />"
		metaHTML = append(metaHTML, metaTag)
	}

	return template.HTML(strings.Join(metaHTML, "\n"))
}

func (p Page) GetCssHtml() template.HTML {
	cssHTML := []string{}
	for _, css := range p.StyleSheet {
		cssHTML = append(cssHTML, fmt.Sprintf("\t<link rel=\"stylesheet\" type=\"text/css\" href=\"%v\" />", css))
	}

	return template.HTML(strings.Join(cssHTML, "\n"))
}

func (p *Page) AddContent(content string, wrapperTag string, wrapperAttrs map[string]string, closeWrapperAfter bool, closeWrapperCount int) {
	p.Content = append(p.Content, content)
	p.Wrapper = append(p.Wrapper, map[string]interface{}{
		"tag":   wrapperTag,
		"attr":  wrapperAttrs,
		"close": closeWrapperAfter,
		"cwc":   closeWrapperCount,
	})
}

func (p Page) GetContent() template.HTML {
	var contentStr []string
	var closeTag []string
	for i, Content := range p.Content {
		var tagPre, tagPost string = "", ""
		wrapper := p.Wrapper[i]
		if wrapper["tag"] != "" {
			attr, ok := wrapper["attr"].(map[string]string)
			strAttr := ""
			if ok {
				for attrK, attrV := range attr {
					strAttr += fmt.Sprintf(" %v", helper.HtmlAttribute(attrK, attrV))
				}
			}
			tagPre = fmt.Sprintf(`<%v%v>`, wrapper["tag"], strAttr)
			tagPost = fmt.Sprintf("</%v>", wrapper["tag"])
			if !wrapper["close"].(bool) {
				closeTag = append(closeTag, tagPost)
				Content = tagPre + Content
			} else {
				Content = tagPre + Content + tagPost
			}
		}
		contentStr = append(contentStr, Content)
		cwc := wrapper["cwc"].(int)
		for j := 0; j < cwc; j++ {
			helper.PrintlnIf(fmt.Sprintf("Close tag %v", closeTag[len(closeTag)-1]), helper.GetConfig().Mode.Debug)
			contentStr = append(contentStr, closeTag[len(closeTag)-1])
			closeTag = closeTag[:len(closeTag)-1]
		}
	}

	for i := len(closeTag) - 1; i >= 0; i-- {
		contentStr = append(contentStr, closeTag[i])
	}

	return template.HTML(strings.Join(contentStr, "\n"))
}
