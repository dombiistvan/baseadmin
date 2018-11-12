package view

import (
	"html/template"
)

type Ajax struct {
	Success bool `json:"success"`
	Html    template.HTML `json:"html"`
	Data	map[string]interface{} `json:"data"`
}
