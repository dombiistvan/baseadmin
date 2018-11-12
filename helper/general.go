package helper

import (
	"log"
)

var (
	renderFloatPrecisionMultipliers = [...]float64{
		1,
		10,
		100,
		1000,
		10000,
		100000,
		1000000,
		10000000,
		100000000,
		1000000000,
	}

	renderFloatPrecisionRounders = [...]float64{
		0.5,
		0.05,
		0.005,
		0.0005,
		0.00005,
		0.000005,
		0.0000005,
		0.00000005,
		0.000000005,
		0.0000000005,
	}
)

func PrintlnIf(txt string, condition bool){
	if(condition){
		log.Println(txt);
	}
}

func GetOption(object map[string]interface{},key string) interface{}{
	val,ok := object[key];
	if(ok){
		switch val.(type){
		case string:
			return val.(string);
			break;
		case []string:
			return val.([]string);
			break;
		case int:
			return val.(int);
			break;
		case bool:
			return val.(bool);
			break;
		default:
			return val;
			break;
		}
	}
	return nil;
}