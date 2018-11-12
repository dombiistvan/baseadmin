package model

import (
	"base/db"
	h "base/helper"
	"base/model/FElement"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/valyala/fasthttp"
	"reflect"
)

type Config struct {
	Id    int64  `db:"id, primarykey, autoincrement"`
	Path  string `db:"path, size:255"`
	Value string `db:"value, size:1000"`
}

func (c Config) GetAll() []Config {
	var Configs []Config
	_, err := db.DbMap.Select(&Configs, fmt.Sprintf("select * from %v order by %v", c.GetTable(), c.GetPrimaryKey()[0]))
	h.Error(err, "", h.ERROR_LVL_ERROR)
	return Configs
}

func (_ Config) Get(ConfigId int64) (Config, error) {
	var Config Config
	if ConfigId == 0 {
		return Config, errors.New(fmt.Sprintf("Could not retrieve Config to ID %v", ConfigId))
	}

	err := db.DbMap.SelectOne(&Config, fmt.Sprintf("SELECT * FROM %v WHERE %v = ?", Config.GetTable(), Config.GetPrimaryKey()[0]), ConfigId)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	if err != nil {
		return Config, err
	}

	if Config.Id == 0 {
		return Config, errors.New(fmt.Sprintf("Could not retrieve Config to ID %v", ConfigId))
	}

	return Config, nil
}

func (_ Config) IsLanguageModel() bool {
	return false
}

func (_ Config) GetTable() string {
	return "config"
}

func (_ Config) GetPrimaryKey() []string {
	return []string{"id"}
}

func GetConfigForm(data map[string]interface{}, action string) Form {
	var Elements []FormElement
	var id = FElement.InputHidden{"id", "id", "", false, true, data["id"].(string)}
	Elements = append(Elements, id)
	var identifier = FElement.InputText{"identifier", "identifier", "", "", "fe.: iden-ti-fier", false, false, data["identifier"].(string), "Unique per language (this will be used to load the Config)", "", "", "", ""}
	Elements = append(Elements, identifier)
	var content = FElement.InputTextarea{"Content", "content", "content", "", "Content to display", false, false, data["content"].(string), "", 80, 5}
	Elements = append(Elements, content)
	var fullColMap = map[string]string{"lg": "12", "md": "12", "sm": "12", "xs": "12"}
	var Fieldsets []Fieldset
	Fieldsets = append(Fieldsets, Fieldset{"left", Elements, fullColMap})
	button := FElement.InputButton{"Submit", "submit", "submit", "pull-right", false, "", true, false, false, nil}
	Fieldsets = append(Fieldsets, Fieldset{"bottom", []FormElement{button}, fullColMap})
	var form = Form{h.GetUrl(action, nil, true, "admin"), "POST", false, Fieldsets, false, nil, nil}

	return form
}

func NewConfig(Id int64, Path string, Value string) Config {
	return Config{
		Id:    Id,
		Path:  Path,
		Value: Value,
	}
}

func NewEmptyConfig() Config {
	return NewConfig(0, "", "")
}

func GetConfigFormValidator(ctx *fasthttp.RequestCtx, Config Config) Validator {
	var Validator Validator
	Validator = Validator.New(ctx)
	Validator.AddField("id", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": false,
		},
	})
	Validator.AddField("path", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
			"format": map[string]interface{}{
				"type":   "regexp",
				"regexp": "^([a-zA-Z0-9\\-\\_\\/]*)+$",
			},
		},
	})
	Validator.AddField("value", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
	})
	return Validator
}

func (c Config) GetByPath(path string, languageCode string) (Config, error) {
	var Config Config
	var query string = fmt.Sprintf("SELECT * FROM %v WHERE %v = ?", c.GetTable(), "path")
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)
	err := db.DbMap.SelectOne(&Config, query, path)
	if err != sql.ErrNoRows {
		return Config, err
	}

	return Config, nil
}

func (c Config) GetValueByPath(path string) string {
	var query string = fmt.Sprintf("SELECT `value` FROM %v WHERE %v = ?", c.GetTable(), "path")
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)
	value, err := db.DbMap.SelectStr(query, path)
	h.Error(err, "", h.ERROR_LVL_WARNING)

	return value
}

func (c Config) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()
	if Conf.Mode.Rebuild_structure {
		h.PrintlnIf(fmt.Sprintf("Drop %v table", c.GetTable()), Conf.Mode.Rebuild_structure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", c.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %v table", c.GetTable()), Conf.Mode.Rebuild_structure)
	dbmap.CreateTablesIfNotExists()
	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "IDX_CONFIG_PATH",
			"type":   "hash",
			"field":  []string{"path"},
			"unique": true,
		},
	}
	tablemap, err := dbmap.TableFor(reflect.TypeOf(Config{}), false)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.Rebuild_structure)
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
	}

	for path, val := range h.GetConfig().ConfigValues {
		var conf Config
		conf.Path = path
		conf.Value = val
		err = db.DbMap.Insert(&conf)
		h.Error(err, "", h.ERROR_LVL_ERROR)
	}

	dbmap.CreateIndex()
}

func (_ Config) IsAutoIncrement() bool {
	return true
}
