package helper

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"strings"
)

type AuthHelper struct{}

type RoleGroupStruct struct{
	Title string `yml:"title"`
	Value string `yml:"value"`
	Children map[string]struct{
		Title string `yml:"title"`
		Value string `yml:"value"`
	} `yml:"children"`
}
type RolesStruct struct {
	Roles map[string]RoleGroupStruct `yml:"roles"`
}

var Roles RolesStruct;

var RolesConfigPath string = "./resource/roles.yml"

func (a *AuthHelper) HasRights(requiredRoles []string, session *Session) bool {
	for _, role := range requiredRoles {
		if (CanAccess(role, session)) {
			return true;
		}
	}

	if (session.Value(USER_SESSION_SUPERADMIN_KEY) == true && !GetConfig().Mode.Debug) {
		PrintlnIf("Superadmin user -> GODMODE -> Let me know your wishes", GetConfig().Mode.Debug);
		return true;
	}

	return false;
}

func GetRoles() RolesStruct {
	succ, err := parseRolesConfig();
	if (nil != err || !succ) {
		Error(err, "Could not retrieve roles config", ERROR_LVL_ERROR);
	}
	return Roles;
}

func parseRolesConfig() (bool, error) {
	dat, err := ioutil.ReadFile(RolesConfigPath);
	Error(err, "", ERROR_LVL_ERROR);
	if (err != nil) {
		return false, err;
	}

	err = yaml.Unmarshal(dat, &Roles)
	Error(err, "", ERROR_LVL_ERROR);
	if (err != nil) {
		return false, err;
	}

	return true, nil;
}

func CanAccess(role string, session *Session) bool {
	switch(role) {
	case "*":
		PrintlnIf("Anyone is allowed", GetConfig().Mode.Debug);
		return true; //anyone
		break;
	case "!@":
		//only NOT logged in user @TODO make it works
		PrintlnIf("Logged out user allowed", GetConfig().Mode.Debug);
		if (session.IsLoggedIn() == false) {
			return true;
		}
		break;
	case "@":
		//only logged in user @TODO make it works
		PrintlnIf("Loggedin user is allowed (admin users)", GetConfig().Mode.Debug);
		if (session.IsLoggedIn()) {
			return true;
		}
		break;
	case "@sa":
		PrintlnIf("Only superadmin allowed (chiefadmin)", GetConfig().Mode.Debug);
		if (session.IsSuperAdmin()) {
			return true;
		}
		break;
	default:
		roleGroup := strings.Split(role,"/")[0]; //for example in case of user/exmample, it is user
		for _,uRole := range session.GetRoles(){
			if(uRole == role || uRole == fmt.Sprintf("%v/*",roleGroup)){
				PrintlnIf(fmt.Sprintf("Required role: %v, rolegroup: %v, user role: %v",role,roleGroup,uRole),GetConfig().Mode.Debug);
				return true;
			}
		}
		break;
	}

	if (session.IsSuperAdmin()) {
		return true;
	}

	return false;
}
