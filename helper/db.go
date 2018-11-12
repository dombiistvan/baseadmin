package helper

import (
	"github.com/go-gorp/gorp"
	"fmt"
	"strings"
)

func AddIndexes(dbmap *gorp.DbMap,table string, indexes map[int]map[string]interface{}){
	for _,index := range(indexes){
		var uniq string = "";
		var idxtype string = " USING %s";
		if(index["unique"].(bool)){
			uniq = " UNIQUE";
		}
		if(len(index["type"].(string))>0){
			idxtype = fmt.Sprintf(idxtype,strings.ToUpper(index["type"].(string)));
		}
		var queryString string = fmt.Sprintf(
			"CREATE%v INDEX %v ON %v(%v)%v",
			uniq,
			index["name"].(string),
			table,
			strings.Join(index["field"].([]string),","),
			idxtype,
		);
		PrintlnIf(queryString,GetConfig().Mode.Debug)
		_,err := dbmap.Query(queryString)
		Error(err,"",ERROR_LVL_ERROR);
	}
}
