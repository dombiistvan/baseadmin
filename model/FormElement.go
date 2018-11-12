package model

type FormElement interface{
	Render(errs map[string]error) string
	HasPreOrPost() bool
}