package db

import (
	"db"
	"fmt"
)
func test struct{
}
func (test) TestPostgres() {
	p := "host=localhost user=forwardtest password=forwardtest dbname=forward sslmode=disable"
	dd, err := db.Open("postgres", p)
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	fmt.Println(dd.Hello())
	dd.Orm().LogMode(true)
	dd.Orm().AutoMigrate(&MyTables{})
	meta, err := dd.GetMeta("my_tables")
	if err != nil {
		fmt.Println("Erro Meta ", err.Error())
	}
	if myInfo, ok := meta.([]db.PostgresTableInfo); ok {
		for _, info := range myInfo {
			fmt.Println("Info", info)
		}
	}
	v, err := dd.GetTables()
	if err != nil {
		fmt.Println("Erro Table ", err.Error())
	}
	fmt.Println(fmt.Printf("Tables %s", v))
}
