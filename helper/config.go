package helper

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

var config Conf
var confParsed bool

type UserGroups []UserGroup

func (ug UserGroups) GetOptions(defOption map[string]string) []map[string]string {
	var options []map[string]string

	if defOption != nil {
		options = append(options, defOption)
	}

	for _, g := range ug {
		options = append(options, map[string]string{
			"value": g.Value,
			"label": g.Label,
		})
	}

	return options
}

type UserGroup struct {
	Label       string `yml:"label"`
	Value       string `yml:"value"`
	Description string `yml:"description"`
}

type Conf struct {
	ViewDir    string `yml:"viewdir"`
	ListenPort string `yml:"listenport"`
	BrandUrl   string `yml:"brandurl"`
	ChiefAdmin map[int]struct {
		Email      string `yml:"email"`
		Password   string `yml:"password"`
		SuperAdmin bool   `yml:"superadmin"`
	} `yml:"chiefadmin"`
	Og struct {
		Url         string `yml:"url"`
		Type        string `yml:"type"`
		Title       string `yml:"title"`
		Description string `yml:"description"`
		Image       string `yml:"image"`
	} `yml:"og"`
	Environment string `yml:"environment"`
	Db          struct {
		Environment map[string]struct {
			Host     string `yml:"host"`
			Username string `yml:"username"`
			Password string `yml:"password"`
			Name     string `yml:"name"`
		} `yml:"environment"`
		MaxIdleCons            int `yml:"maxidleconns"`
		MaxOpenCons            int `yml:"maxopenconns"`
		MaxConnLifetimeMinutes int `yml:"maxconnlifetimeminutes"`
	} `yml:"db"`
	Server struct {
		ReadTimeoutSeconds  int    `yml:"readtimeoutseconds"`
		WriteTimeoutSeconds int    `yml:"writetimeoutseconds"`
		SessionKey          string `yml:"sessionkey"`
		MaxRPS              int    `yml:"maxrps"`
		BanMinutes          int    `yml:"banminutes"`
		BanActive           bool   `yml:"banactive"`
		Name                string `yml:"name"`
	} `yml:"server"`
	Mode struct {
		Live              bool `yml:"live"`
		Debug             bool `yml:"debug"`
		Rebuild_structure bool `yml:"rebuild_structure"`
	} `yml:"mode"`
	Cache struct {
		Enabled bool   `yml:"enabled"`
		Type    string `yml:"type"`
		Dir     string `yml:"dir"`
	} `yml:"cache"`
	AdminRouter  string            `yml:"adminrouter"`
	ConfigValues map[string]string `yml:"configvalues"`
	Language     struct {
		Allowed []string `yml:"allowed"`
	} `yml:"language"`
	Ug UserGroups `yml:"ug"`
}

var resourceCfg *ResourceConfig

func SetResourceConfig(config ResourceConfig) {
	if err := config.Prepare(); err != nil {
		Error(err, "", ERROR_LVL_ERROR)
	}
	resourceCfg = &config
}

func GetConfig() Conf {
	var err error
	if confParsed {
		return config
	}
	config, err = parseConfig()
	if nil != err {
		Error(err, "Could not retrieve config", ERROR_LVL_ERROR)
	}
	return config
}

func parseConfig() (Conf, error) {
	var Config Conf
	var err error
	var dat []byte

	dat, err = ioutil.ReadFile(resourceCfg.GetConfigFilePath())
	Error(err, "", ERROR_LVL_ERROR)
	if err != nil {
		Error(err, "", ERROR_LVL_ERROR)
	}

	err = yaml.Unmarshal(dat, &Config)
	Error(err, "", ERROR_LVL_ERROR)
	if err != nil {
		return Conf{}, err
	}

	Config.Cache.Dir = TrimPath(Config.Cache.Dir)
	return Config, nil
}

var (
	ErrorEmptyPath = errors.New("the path is empty")
)

type ResourceConfig struct {
	ResourcePath string
	ConfigFile   string
	RolesFile    string
	MenuFile     string
	LanguagePath string

	configFilePath string
	rolesFilePath  string
	menuFilePath   string
}

func (rc ResourceConfig) GetConfigFilePath() string {
	return rc.configFilePath
}

func (rc ResourceConfig) GetRoleFilePath() string {
	return rc.rolesFilePath
}

func (rc ResourceConfig) GetMenuFilePath() string {
	return rc.menuFilePath
}

func (rc ResourceConfig) GetLanguageFilesPath() string {
	return rc.LanguagePath
}

func (rc *ResourceConfig) Prepare() error {
	if rc.ResourcePath == "" {
		return ErrorEmptyPath
	}

	if rc.ConfigFile == "" {
		return ErrorEmptyPath
	}

	if rc.RolesFile == "" {
		return ErrorEmptyPath
	}

	if rc.MenuFile == "" {
		return ErrorEmptyPath
	}

	rc.ResourcePath = strings.Trim(rc.ResourcePath, "/")

	_, err := ioutil.ReadDir(rc.ResourcePath)
	if err != nil {
		return err
	}

	rc.LanguagePath = strings.Trim(rc.LanguagePath, " ")
	rc.LanguagePath = strings.Trim(rc.LanguagePath, "/")

	if strings.Index(rc.LanguagePath, "%ResourcePath%") == 0 {
		splitted := strings.Split(rc.LanguagePath, "%ResourcePath%")
		rc.LanguagePath = fmt.Sprintf("%s/%s", rc.ResourcePath, strings.Trim(splitted[1], "/"))
	}

	_, err = ioutil.ReadDir(rc.LanguagePath)
	if err != nil {
		return err
	}

	rc.configFilePath = fmt.Sprintf("%s/%s", rc.ResourcePath, rc.ConfigFile)
	rc.rolesFilePath = fmt.Sprintf("%s/%s", rc.ResourcePath, rc.RolesFile)
	rc.menuFilePath = fmt.Sprintf("%s/%s", rc.ResourcePath, rc.MenuFile)

	if _, err := ioutil.ReadFile(rc.configFilePath); err != nil {
		return err
	}

	if _, err := ioutil.ReadFile(rc.rolesFilePath); err != nil {
		return err
	}

	if _, err := ioutil.ReadFile(rc.menuFilePath); err != nil {
		return err
	}

	return nil
}
