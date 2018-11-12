package helper

import (
	"log"
	"database/sql"
)

const ERROR_LVL_NOTICE = 0;
const ERROR_LVL_WARNING = 1;
const ERROR_LVL_ERROR = 2;

func Error(err error, msg string, lvl int) {
	if (err != nil && err != sql.ErrNoRows) {
		if (msg == "") {
			msg = err.Error();
		}
		switch(lvl) {
		default:
			log.Println(msg)
		case ERROR_LVL_WARNING:
			panic(err);
			log.Println(msg)
		case ERROR_LVL_ERROR:
			panic(err);
			log.Println(msg);
			break;
		}
	}
}
