package helper

import (
	"time"
)

func GetTimeNow() time.Time{
	m := make(map[string]string)
	m["Hungary"] = "+01.00h"

	offSet, err := time.ParseDuration(m["Hungary"])
	Error(err,"",ERROR_LVL_WARNING);
	t := time.Now().UTC().Add(offSet)

	return t;
}
