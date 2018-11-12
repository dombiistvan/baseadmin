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

type Block struct {
	Id         int64  `db:"id, primarykey, autoincrement"`
	Identifier string `db:"identifier, size:255"`
	Content    string `db:"content, size:1000"`
	Lc         string `db:"lc, size:2"`
}

func (b Block) GetAll() []Block {
	var blocks []Block
	_, err := db.DbMap.Select(&blocks, fmt.Sprintf("select * from %v order by %v", b.GetTable(), b.GetPrimaryKey()[0]))
	h.Error(err, "", h.ERROR_LVL_ERROR)
	return blocks
}

func (_ Block) Get(blockId int64) (Block, error) {
	var block Block
	if blockId == 0 {
		return block, errors.New(fmt.Sprintf("Could not retrieve block to ID %v", blockId))
	}

	err := db.DbMap.SelectOne(&block, fmt.Sprintf("SELECT * FROM %v WHERE %v = ?", block.GetTable(), block.GetPrimaryKey()[0]), blockId)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	if err != nil {
		return block, err
	}

	if block.Id == 0 {
		return block, errors.New(fmt.Sprintf("Could not retrieve block to ID %v", blockId))
	}

	return block, nil
}

func (_ Block) IsLanguageModel() bool {
	return true
}

func (_ Block) GetTable() string {
	return "block"
}

func (_ Block) GetPrimaryKey() []string {
	return []string{"id"}
}

func (_ Block) IsAutoIncrement() bool {
	return true
}

func GetBlockForm(data map[string]interface{}, action string) Form {
	var Elements []FormElement
	var id = FElement.InputHidden{"id", "id", "", false, true, data["id"].(string)}
	Elements = append(Elements, id)
	var lc = FElement.InputHidden{"lc", "lc", "", false, true, data["lc"].(string)}
	Elements = append(Elements, lc)
	var identifier = FElement.InputText{"identifier", "identifier", "", "", "fe.: iden-ti-fier", false, false, data["identifier"].(string), "Unique per language (this will be used to load the block)", "", "", "", ""}
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

func NewBlock(Id int64, Identifier string, Content string, Lang string) Block {
	return Block{
		Id:         Id,
		Identifier: Identifier,
		Content:    Content,
		Lc:         Lang,
	}
}

func NewEmptyBlock() Block {
	return NewBlock(0, "", "", h.DefLang)
}

func GetBlockFormValidator(ctx *fasthttp.RequestCtx, Block Block) Validator {
	var Validator Validator
	Validator = Validator.New(ctx)
	Validator.AddField("id", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": false,
		},
	})
	Validator.AddField("identifier", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
			"format": map[string]interface{}{
				"type":          "regexp",
				"patternregexp": "^([a-zA-Z0-9\\-\\_]*)+$",
			},
		},
	})
	Validator.AddField("content", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
		},
	})
	return Validator
}

func (b Block) GetByIdentifier(identifier string, languageCode string) (Block, error) {
	var block Block
	var query string = fmt.Sprintf("SELECT * FROM %v WHERE %v= ? AND %v = ?", b.GetTable(), "lc", "identifier")
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)
	err := db.DbMap.SelectOne(&block, query, languageCode, identifier)
	if err != sql.ErrNoRows {
		return block, err
	}

	return block, nil
}

func (b Block) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()
	if Conf.Mode.Rebuild_structure {
		h.PrintlnIf(fmt.Sprintf("Drop %v table", b.GetTable()), Conf.Mode.Rebuild_structure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", b.GetTable()))
	}

	h.PrintlnIf(fmt.Sprintf("Create %v table", b.GetTable()), Conf.Mode.Rebuild_structure)
	dbmap.CreateTablesIfNotExists()
	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	indexes = map[int]map[string]interface{}{
		0: {
			"name":   "IDX_BLOCK_IDENTIFIER_LC",
			"type":   "hash",
			"field":  []string{"identifier", "lc"},
			"unique": true,
		},
	}
	tablemap, err := dbmap.TableFor(reflect.TypeOf(Block{}), false)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	for _, index := range indexes {
		h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.Rebuild_structure)
		tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
	}

	dbmap.CreateIndex()
	var blockCont map[string]map[string]string = map[string]map[string]string{
		/*"introduction": {
			"en": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque fermentum est a tempor ullamcorper. Nunc tristique maximus velit non cursus. Praesent eget ipsum lobortis dolor consequat ultrices sed in mi. Proin dignissim pulvinar dolor et varius. Duis id tempus libero, sit amet finibus nulla. Nunc ut leo interdum, vehicula nulla eget, sollicitudin metus. Donec accumsan nulla in nibh accumsan convallis. In lacinia pretium varius. Nullam ullamcorper velit in neque aliquam, in tempor augue convallis. Integer dictum ex non odio molestie faucibus.",
			"de": "In hac habitasse platea dictumst. Praesent in ex sit amet nisl ornare sodales. In hac habitasse platea dictumst. Phasellus semper luctus augue sed molestie. Maecenas sit amet urna laoreet, faucibus dui a, dignissim leo. Vivamus dictum luctus nisl, ac lacinia mauris posuere a. Sed mollis mi metus, at rutrum metus iaculis at. Morbi auctor volutpat nunc, ut dapibus est viverra at. In bibendum faucibus magna, a pretium libero scelerisque sed. Aenean eget mi eleifend, blandit diam quis, vulputate arcu.",
			"hu": "Donec vitae nulla dolor. Fusce nec nulla non turpis egestas vestibulum non a enim. Phasellus tempor arcu a magna porttitor, eu mollis orci pellentesque. Curabitur eget vulputate tellus. In gravida massa nec tellus varius mollis. Vivamus ultricies urna in odio semper, at accumsan est semper. Donec at ornare urna, vel ultricies diam. Praesent pellentesque aliquet enim, pellentesque semper elit dapibus sollicitudin. Integer id gravida neque, et hendrerit sem.",
		},*/
	}
	for k, lMap := range blockCont {
		for lk, c := range lMap {
			block, err := b.GetByIdentifier(k, lk)
			h.Error(err, "", h.ERROR_LVL_ERROR)
			if block.Id == 0 {
				block.Identifier = k
				block.Lc = lk
				block.Content = c
				err := dbmap.Insert(&block)
				h.Error(err, "", h.ERROR_LVL_ERROR)
			}
		}
	}
}
