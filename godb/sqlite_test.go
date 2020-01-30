package db

import (
	"db"
	"fmt"
)

func (test) TestSqlite() {
	dd, err := db.Open("sqlite3", "./my.db")
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	fmt.Println(dd.Hello())
	dd.Orm().LogMode(true)
	dd.Orm().AutoMigrate(&MyTables{})
	dd.Orm().Exec("PRAGMA foreign_keys = ON;")
	dd.Orm().SingularTable(true)
	meta, err := dd.GetMeta("my_tables")
	if err != nil {
		fmt.Println("Erro Meta ", err.Error())
	}
	if m, ok := meta.([]db.TableInfo); ok {
		for _, mv := range m {
			fmt.Println("INFO", mv)
		}
	}
	v, err := dd.GetTables()
	if err != nil {
		fmt.Println("Erro Table ", err.Error())
	}
	fmt.Println(fmt.Printf("Tables %s", v))
}
