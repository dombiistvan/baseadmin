package helper

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

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
}

var ConfigFilePath string = "./resource/config.yml"

func GetConfig() Conf {
	Config, err := parseConfig()
	if nil != err {
		Error(err, "Could not retrieve config", ErrorLvlError)
	}
	return Config
}

func parseConfig() (Conf, error) {
	var Config Conf
	var err error
	var dat []byte
	dat, err = ioutil.ReadFile(ConfigFilePath)
	Error(err, "", ErrorLvlError)
	if err != nil {
		Error(err, "", ErrorLvlError)
	}

	err = yaml.Unmarshal(dat, &Config)
	Error(err, "", ErrorLvlError)
	if err != nil {
		return Conf{}, err
	}

	Config.Cache.Dir = TrimPath(Config.Cache.Dir)
	return Config, nil
}
