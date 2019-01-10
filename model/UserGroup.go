package model

import (
	dbHelper "baseadmin/db"
	h "baseadmin/helper"
	"baseadmin/model/FElement"
	"errors"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/valyala/fasthttp"
	"reflect"
	"strconv"
	"strings"
)

type UserGroup struct {
	Id         int64  `db:"id, primarykey, autoincrement"`
	Name       string `db:"name, size:254"`
	Identifier string `db:"identifier, size:254"`
}

func (_ UserGroup) Get(id int64) (UserGroup, error) {
	var usergroup UserGroup
	if id == 0 {
		return usergroup, errors.New(fmt.Sprintf("Could not retrieve usergroup to ID %v", id))
	}

	err := dbHelper.DbMap.SelectOne(&usergroup, fmt.Sprintf("SELECT * FROM %v WHERE %v = ?", usergroup.GetTable(), usergroup.GetPrimaryKey()[0]), id)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	if err != nil {
		return usergroup, err
	}

	if usergroup.Id == 0 {
		return usergroup, errors.New(fmt.Sprintf("Could not retrieve usergroup to ID %v", id))
	}

	return usergroup, nil
}

func (_ UserGroup) GetByIdentifier(identifier string) (UserGroup, error) {
	var usergroup UserGroup
	if identifier == "" {
		return usergroup, errors.New(fmt.Sprintf("Could not retrieve usergroup to Identifier %v", identifier))
	}

	err := dbHelper.DbMap.SelectOne(&usergroup, fmt.Sprintf("SELECT * FROM %v WHERE %s = ?", usergroup.GetTable(), "identifier"), identifier)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	if err != nil {
		return usergroup, err
	}

	if usergroup.Id == 0 {
		return usergroup, errors.New(fmt.Sprintf("Could not retrieve usergroup to Identifier %s", identifier))
	}

	return usergroup, nil
}

// implement the PreInsert and PreUpdate hooks
func (ug *UserGroup) PreInsert(s gorp.SqlExecutor) error {
	return nil
}

func (ug *UserGroup) PreUpdate(s gorp.SqlExecutor) error {
	return nil
}

func (ug UserGroup) GetOptions(defOption map[string]string) []map[string]string {
	var groups = ug.GetAll()
	var options []map[string]string
	if defOption != nil {
		_, okl := defOption["label"]
		_, okv := defOption["value"]
		if okl || okv {
			options = append(options, defOption)
		}
	}
	for _, group := range groups {
		options = append(options, map[string]string{"label": group.Name, "value": strconv.Itoa(int(group.Id))})
	}
	return options
}

func (ug UserGroup) GetAll() []UserGroup {
	var groups []UserGroup
	query := fmt.Sprintf("SELECT * FROM %v ORDER BY %v", ug.GetTable(), "name")
	h.PrintlnIf(query, h.GetConfig().Mode.Debug)
	_, err := dbHelper.DbMap.Select(&groups, query)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	return groups
}

func (ug *UserGroup) ModifyRoles(roles []string) {
	var ur UserRole
	fmt.Println(roles)
	h.PrintlnIf(fmt.Sprintf("Modify usergroup roles %v", ug.Id), h.GetConfig().Mode.Debug)
	if ug.Id > 0 {
		_, err := dbHelper.DbMap.Exec(fmt.Sprintf("DELETE FROM %v WHERE user_group_id = ?", ur.GetTable()), ug.Id)
		if err != nil {
			h.Error(err, "", h.ERROR_LVL_ERROR)
			return
		}
	}

	var SkipSubRoles = make(map[string]bool)

	for _, strRole := range roles {
		var role UserRole
		if strRole == "*" {
			role = UserRole{UserGroupId: ug.Id, Role: strRole}
			err := dbHelper.DbMap.Insert(&role)
			if err != nil {
				h.Error(err, "", h.ERROR_LVL_ERROR)
			}
			return
		}
		roleExp := strings.Split(strRole, "/")
		//if SkipSubRoles[user] exists, skip all user subrole from insert
		_, ok := SkipSubRoles[roleExp[0]]
		if ok {
			continue
		}
		if roleExp[1] == "*" {
			//if user/* is the role, add user to skipsubrole, to skip inserting subroles
			SkipSubRoles[roleExp[0]] = true
		}

		role.UserGroupId = ug.Id
		role.Role = strRole

		err := dbHelper.DbMap.Insert(&role)
		if err != nil {
			h.Error(err, "", h.ERROR_LVL_ERROR)
		}
	}
}

func (ug UserGroup) BuildStructure(dbmap *gorp.DbMap) {
	Conf := h.GetConfig()

	var indexes map[int]map[string]interface{} = make(map[int]map[string]interface{})

	if Conf.Mode.Rebuild_structure {
		h.PrintlnIf(fmt.Sprintf("Drop %v table", ug.GetTable()), Conf.Mode.Rebuild_structure)
		dbmap.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", ug.GetTable()))

		h.PrintlnIf(fmt.Sprintf("Create %v table", ug.GetTable()), Conf.Mode.Rebuild_structure)
		dbmap.CreateTablesIfNotExists()
		tablemap, err := dbmap.TableFor(reflect.TypeOf(UserGroup{}), false)
		h.Error(err, "", h.ERROR_LVL_ERROR)
		for _, index := range indexes {
			h.PrintlnIf(fmt.Sprintf("Create %s index", index["name"].(string)), Conf.Mode.Rebuild_structure)
			tablemap.AddIndex(index["name"].(string), index["type"].(string), index["field"].([]string)).SetUnique(index["unique"].(bool))
		}
		dbmap.CreateIndex()

		h.PrintlnIf(fmt.Sprintf("Admin group to UserGroups"), Conf.Mode.Debug)

		var adminGroup UserGroup

		adminGroup.Name = "Admin"
		adminGroup.Identifier = "admin"

		err = dbmap.Insert(&adminGroup)
		h.Error(err, "", h.ERROR_LVL_ERROR)

		if err == nil {
			adminGroup.ModifyRoles([]string{"*"})
		}
	}
}

func (ug *UserGroup) GetRoles() []string {
	var UserRoles []UserRole
	var ReturnRoles []string
	_, err := dbHelper.DbMap.Select(&UserRoles, "select * from user_role WHERE user_group_id = ?", ug.Id)
	h.Error(err, "", h.ERROR_LVL_ERROR)
	for _, role := range UserRoles {
		ReturnRoles = append(ReturnRoles, role.Role)
	}
	return ReturnRoles
}

func GetUserGroupForm(data map[string]interface{}, action string) Form {
	var ElementsLeft []FormElement
	var ElementsRight []FormElement
	var Checkboxes []FElement.InputCheckbox

	for roleGroup, properties := range h.GetRoles().Roles {
		checkbox := FElement.InputCheckbox{
			properties.Title,
			"role",
			fmt.Sprintf("role_%v_all", roleGroup),
			"roles_group",
			false,
			false,
			properties.Value,
			data["role"].([]string),
			false,
		}
		Checkboxes = append(Checkboxes, checkbox)
		for sub, role := range properties.Children {
			checkbox = FElement.InputCheckbox{
				"&nbsp;&nbsp;&nbsp;&nbsp;" + role.Title,
				"role", //fmt.Sprintf("role_%v_%v",roleGroup,sub),
				fmt.Sprintf("role_%v_%v", roleGroup, sub),
				"roles_entry",
				false,
				false,
				role.Value,
				data["role"].([]string),
				false,
			}
			Checkboxes = append(Checkboxes, checkbox)
		}
	}

	var id = FElement.InputHidden{"id", "id", "", false, true, data["id"].(string)}
	ElementsLeft = append(ElementsLeft, id)

	var name = FElement.InputText{"Name", "name", "name", "", "f.e.: Product Owner", false, false, data["name"].(string), "", "", "", "", ""}
	ElementsLeft = append(ElementsLeft, name)

	var identifier = FElement.InputText{"Identifier", "identifier", "identifier", "", "f.e.: prodowner", false, false, data["identifier"].(string), "", "", "", "", ""}
	ElementsLeft = append(ElementsLeft, identifier)

	RoleGroup := FElement.CheckboxGroup{"Roles", Checkboxes, FElement.Static{}}
	ElementsRight = append(ElementsRight, RoleGroup)

	var colMap map[string]string = map[string]string{
		"lg": "6",
		"md": "6",
		"sm": "12",
		"xs": "12",
	}

	var Fieldsets []Fieldset

	Fieldsets = append(Fieldsets, Fieldset{"base", ElementsLeft, colMap})
	Fieldsets = append(Fieldsets, Fieldset{"other", ElementsRight, colMap})

	button := FElement.InputButton{"Submit", "submit", "submit", "pull-right", false, "", true, false, false, nil}

	Fieldsets = append(Fieldsets, Fieldset{"bottom", []FormElement{button}, map[string]string{"lg": "12", "md": "12", "sm": "12", "xs": "12"}})

	var form = Form{h.GetUrl(action, nil, true, "admin"), "POST", false, Fieldsets, false, nil, nil}

	return form
}

func GetUserGroupFormValidator(ctx *fasthttp.RequestCtx, UserGroup *UserGroup) Validator {
	var Validator Validator
	Validator.Init(ctx)
	Validator.AddField("id", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": false,
		},
	})
	Validator.AddField("name", map[string]interface{}{
		"roles": map[string]interface{}{
			"required": true,
			"length":   map[string]interface{}{"min": 3},
		},
	})
	Validator.AddField("role", map[string]interface{}{
		"multi": true,
		"roles": map[string]interface{}{
			"required": true,
		},
	})

	return Validator
}

func (_ UserGroup) IsLanguageModel() bool {
	return false
}

func (_ UserGroup) GetTable() string {
	return "user_group"
}

func (_ UserGroup) GetPrimaryKey() []string {
	return []string{"id"}
}

func (_ UserGroup) IsAutoIncrement() bool {
	return true
}
