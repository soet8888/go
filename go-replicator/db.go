package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DEFAULT_SQL_STRING = "VARCHAR(255)"
	DEFAUTL_SQL_FLOAT  = "REAL"
	DEAFULT_SQL_BOOL   = "BOOL"
	DEFAULT_SQL_INT    = "INTEGER"
	DEFAULT_SQL_DATE   = "DATETIME"
)

type FieldType struct {
	Field       string `json:"field"`
	Type        string `json:"type"`
	Association string `json:"association"`
}
type TableInfo struct {
	Cid       uint        `json:"cid"`
	Type      interface{} `json:"type"`
	Name      string      `json:"name"`
	Notnull   string      `json:"not_null"`
	Pk        int         `json:"pk"`
	DfltValue interface{} `json:"deflt_value"`
}
type Db struct {
	orm   *gorm.DB
	mutex *sync.Mutex
}

func NewDb(driver string, param string, dbMigration bool, logMode bool) *Db {
	orm, err := gorm.Open(driver, param)
	if err != nil {
		fmt.Println("Panic", err.Error())
		panic("failed to connect database")
	}
	// assign driver
	dbDriver = driver
	// Ping server
	err = orm.DB().Ping()
	if err != nil {
		log.Fatal("Error: Could not establish a connection with the database - ", err)
	}

	orm.Exec("PRAGMA foreign_keys = ON;")
	orm.LogMode(logMode)
	orm.SingularTable(true)
	orm.AutoMigrate()
	db := &Db{orm: orm, mutex: &sync.Mutex{}}
	return db
}

func (db Db) Close() {
	db.orm.Close()
}

func (db Db) Insert(obj interface{}) error {
	tx := db.orm.Begin()
	err := tx.Create(obj).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (db Db) Update(obj interface{}, para map[string]interface{}) error {
	tx := db.orm.Begin()

	err := tx.Model(obj).Updates(para).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// get by id
func (db Db) GetById(obj string, id string) (interface{}, error) {
	var d []interface{}
	con := db.orm
	con = con.Table(obj)
	con = con.Where("id = ?", id)
	row, err := con.Rows()
	if err != nil {
		return nil, err
	}
	colums, err := row.Columns()
	if err != nil {
		return nil, err
	}
	for row.Next() {
		values := make([]interface{}, len(colums))
		record := make([]interface{}, len(colums))
		mapData := make(map[string]interface{})
		for i, _ := range colums {
			values[i] = &record[i]
		}
		err = row.Scan(values...)
		if err != nil {
			return nil, err
		}
		for key, v := range record {
			var str interface{}
			switch value := v.(type) {
			case int64:
				str = value
			case float64:
				str = value
			case time.Time:
				str = value
			default:
				str = fmt.Sprintf("%s", value)
			}
			mapData[colums[key]] = str
		}
		d = append(d, mapData)
	}
	if len(d) != 1 {
		return nil, errors.New("Record not found")
	}
	return d[0], nil
}
func (db Db) Get(objs interface{}, where map[string]interface{}, order string, limit uint64, offset uint64, deleted bool) error {
	con := db.orm
	if deleted {
		con = con.Unscoped()
	}

	keys, values := whereExactKeyValues(where)
	con = con.Where(keys, values)
	if deleted {
		con = con.Where("deleted_at is not null")
	}

	if order != "" {
		con = con.Order(order)
	}
	if limit != 0 {
		con = con.Limit(limit)
	}
	if offset != 0 {
		con = con.Offset(offset)
	}
	err := con.Find(objs).Error
	if err != nil {
		return err
	}
	return nil
}
func (f Filter) RawDate() string {
	v := "\"" + fmt.Sprint(f.Value) + "\""
	return f.Field + ` ` + f.Compare + ` date(` + v + `)`
}

// get max value
func (db Db) GetMaxValue(obj interface{}) (uint64, error) {
	var max uint64
	row := db.orm.Select("max(tran_event_count)").Find(obj).Row()
	err := row.Scan(&max)
	if err != nil {
		return max, err
	}
	return max, nil
}
func (db Db) GetCount(obj string, data DataParser) (uint64, error) {
	gstr := strings.Join(data.GroupBy, ",")
	con := db.orm
	con = con.Table(obj)
	con = con.Group(gstr)
	for _, f := range data.Filter {
		if f.Raw {
			fValue := f.RawDate()
			con = con.Where(fValue)
		} else {
			con = con.Where(f.Field+" "+f.Compare+" ?", f.Value)
		}
	}
	var count uint64
	err := con.Count(&count).Error
	if err != nil {
		return count, err
	}
	return count, nil
}
func (db Db) GetFilterValue(obj string, data DataParser) (Return, error) {
	var d []map[string]interface{}
	con := db.orm
	con = con.Table(obj)
	if data.HasFields() {
		con = con.Select(data.Field)
	} else {
		con = con.Select("*")
	}
	if data.HasGroupBys() {
		con = con.Group(strings.Join(data.GroupBy, ","))
	}
	if data.HasFilters() {
		for _, f := range data.Filter {
			if f.Raw {
				fValue := f.RawDate()
				con = con.Where(fValue)
			} else {
				con = con.Where(f.Field+" "+f.Compare+" ?", f.Value)
			}
		}
	}
	if data.HasSorts() {
		for _, sort := range data.Sort {
			con = con.Order(sort)
		}
	}
	if data.Limit != 0 {
		con = con.Limit(data.Limit)
	}
	if data.Offset != 0 {
		con = con.Offset(data.Offset)
	}
	row, err := con.Rows()
	if err != nil {
		return d, err
	}
	colums, err := row.Columns()
	if err != nil {
		return d, err
	}
	for row.Next() {
		values := make([]interface{}, len(colums))
		record := make([]interface{}, len(colums))
		mapData := make(map[string]interface{})
		for i, _ := range colums {
			values[i] = &record[i]
		}
		err = row.Scan(values...)
		if err != nil {
			return d, err
		}
		for key, v := range record {
			var str interface{}
			switch value := v.(type) {
			case int64:
				str = value
			case float64:
				str = value
			case time.Time:
				str = value
			default:
				str = fmt.Sprintf("%s", value)
			}
			mapData[colums[key]] = str
		}
		d = append(d, mapData)
	}
	return d, nil
}
func (db Db) GetMeta(obj interface{}) Return {
	str := fmt.Sprint(obj)
	q := "pragma table_info(" + str + ")"
	row, err := db.orm.DB().Query(q)
	if err != nil {
		log.Println("%v", err.Error())
		return err
	}
	var info []TableInfo
	var cid uint
	var t string
	var name string
	var notnull string
	var pk int
	var dfltValue interface{}
	for row.Next() {
		err = row.Scan(&cid, &name, &t, &notnull, &dfltValue, &pk)
		if err != nil {
			log.Println("Error %v", err.Error())
		}
		table_info := TableInfo{
			Cid:       cid,
			Name:      name,
			Type:      t,
			Notnull:   notnull,
			DfltValue: dfltValue,
			Pk:        pk,
		}
		info = append(info, table_info)
	}
	return info
}
func whereExactKeyValues(where map[string]interface{}) (keys string, values []interface{}) {
	var i = 1
	if where != nil && len(where) > 0 {
		for k, v := range where {
			if keys != "" {
				keys = keys + " and "
			}
			if vv, ok := v.(string); ok {
				values = append(values, vv)                   //  value
				keys = fmt.Sprintf("%s %s = $%d", keys, k, i) // key like $#
			} else {
				values = append(values, v)                    // value
				keys = fmt.Sprintf("%s %s = $%d", keys, k, i) // key = $#
			}
			i = i + 1
		}
	}
	return
}

func getObjects() []string {
	var sqlMaster []SqliteMaster
	err := db.Get(&sqlMaster, nil, "", 0, 0, false)
	if err != nil {
		log.Printf("[Sqlite Master ERROR] :%s", err.Error())
	}
	var objs []string
	if len(sqlMaster) > 0 {
		for _, t := range sqlMaster {
			if t.Type == "table" || t.Type == "view" {
				objs = append(objs, t.TblName)
			}
		}
	}
	return objs
}
func getObjectsAndFields() map[string]interface{} {
	var sqlMaster []SqliteMaster
	err := db.Get(&sqlMaster, nil, "", 0, 0, false)
	if err != nil {
		log.Printf("[Sqlite Master ERROR] :%s", err.Error())
	}
	obj_prv := make(map[string]interface{})
	if len(sqlMaster) > 0 {
		for _, t := range sqlMaster {
			if t.Type == "table" || t.Type == "view" {
				var prvArr []string
				meta := db.GetMeta(t.TblName)
				if m, ok := meta.([]TableInfo); ok {
					for _, prv := range m {
						prvArr = append(prvArr, prv.Name)
					}
				}
				obj_prv[t.TblName] = prvArr
			}
		}
	}
	return obj_prv
}
func CheckObj(obj string) error {
	objs := getObjects()
	if len(objs) < 0 {
		return errors.New("Objects list not found.")
	}
	for _, ob := range objs {
		if obj == ob {
			return nil
		}
	}
	return errors.New("Invalid object.")
}

// get view data

func (db Db) GetAll(obj string) (Return, error) {
	var searchValue []interface{}
	q := "select * from " + obj
	row, err := db.orm.DB().Query(q)
	if err != nil {
		return nil, err
	}
	columns, err := row.Columns()
	if err != nil {
		return nil, err
	}
	for row.Next() {
		values := make([]interface{}, len(columns))
		record := make([]interface{}, len(columns))
		data := make(map[string]interface{})
		for i, _ := range columns {
			values[i] = &record[i]
		}
		err = row.Scan(values...)
		if err != nil {
			return nil, err
		}
		for key, v := range record {
			var str interface{}
			switch value := v.(type) {
			case int64:
				str = fmt.Sprint(value)
			case float64:
				str = fmt.Sprintf("%f", value)
			case time.Time:
				str = value
			default:
				str = fmt.Sprintf("%s", value)
			}
			data[columns[key]] = str
		}
		searchValue = append(searchValue, data)
	}
	return searchValue, nil
}

// search term
func (db Db) GetSearchTerm(obj string, field string, term string, limit uint64) (Return, error) {
	d := " DISTINCT " + field
	w := field + " like '%" + term + "%'"
	var record interface{}
	var value []interface{}
	data := make(map[string]interface{})
	con := db.orm
	con = con.Table(obj)
	con = con.Select(d)
	con = con.Where(w)
	if limit != 0 {
		con = con.Limit(limit)
	}
	row, err := con.Rows()
	if err != nil {
		return value, err
	}
	for row.Next() {
		err := row.Scan(&record)
		if err != nil {
			return value, err
		}
		var str interface{}
		switch v := record.(type) {
		case int64:
			str = v
		case float64:
			str = v
		case time.Time:
			str = value
		default:
			str = fmt.Sprintf("%s", v)
		}
		value = append(value, str)
	}
	data[field] = value
	return data, nil
}
func GetFullTextSearch(obj string, term string, limit uint64) ([]interface{}, error) {
	var searchValue []interface{}
	q := "select * from " + obj + " where " + obj + " match \"" + term + "*\" limit " + fmt.Sprintf("%v", limit)
	fmt.Println("Full Query", q)
	row, err := db.orm.DB().Query(q)
	if err != nil {
		log.Println("%v", err.Error())
		return nil, err
	}
	columns, err := row.Columns()
	if err != nil {
		return nil, err
	}
	for row.Next() {
		values := make([]interface{}, len(columns))
		record := make([]interface{}, len(columns))
		data := make(map[string]interface{})
		for i, _ := range columns {
			values[i] = &record[i]
		}
		err = row.Scan(values...)
		if err != nil {
			fmt.Println("Scan Error", err.Error())
			return nil, err
		}
		for key, v := range record {
			var str interface{}
			switch value := v.(type) {
			case int64:
				str = fmt.Sprint(value)
			case float64:
				str = fmt.Sprintf("%f", value)
			case time.Time:
				str = value
			default:
				str = fmt.Sprintf("%s", value)
			}
			data[columns[key]] = str
		}
		searchValue = append(searchValue, data)
	}
	return searchValue, nil
}

//SELECT * FROM ft WHERE ft MATCH 'b : (uvw AND xyz)';
