package db

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
)

type PostgresMaster struct {
	MyTable
	TableCatalog              string
	TableSchema               string
	TableType                 string
	SelfReferencingColumnName string
	ReferenceGeneration       string
	UserDefinedTypeName       string
	IsInsertableInto          string
	IsTyped                   string
	CommitAction              string
}
type PostgresTableInfo struct {
	MyTable
	TableCatalog    string
	TableSchema     string
	ColumnName      string
	OrdinalPosition int
	ColumnDefault   string
	IsNullable      string
	DataType        interface{}
}

func (PostgresTableInfo) TableName() string {
	return "information_schema.columns"
}
func (PostgresMaster) TableName() string {
	return "information_schema.tables"
}

type MyTable struct {
	TableName string
}
type Postgres struct {
	orm   *gorm.DB
	mutex *sync.Mutex
}

func (Postgres) Hello() string {
	return "Postgres"
}
func (p Postgres) Orm() *gorm.DB {
	return p.orm
}
func (p Postgres) Mutex() *sync.Mutex {
	return p.mutex
}
func (db Postgres) Insert(obj interface{}) error {
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
func (db Postgres) Delete(obj interface{}, para map[string]interface{}) error {
	tx := db.orm.Begin()

	err := tx.Where(para).Delete(obj).Error
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
func (db Postgres) Get(objs interface{}, where map[string]interface{}, order string, limit uint64, offset uint64, deleted bool) error {
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
func (db Postgres) Update(obj interface{}, para map[string]interface{}) error {
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
func (db Postgres) GetMeta(name interface{}) (interface{}, error) {
	tableName := fmt.Sprint(name)
	var tabInfo []PostgresTableInfo
	err := db.Get(&tabInfo, map[string]interface{}{"table_name": tableName}, "", 0, 0, false)
	if err != nil {
		return tabInfo, err
	}
	return tabInfo, nil
}
func (db Postgres) GetTables() (interface{}, error) {
	var tables []PostgresMaster
	if err := db.Get(&tables, map[string]interface{}{"table_type": "BASE TABLE"}, "", 0, 0, false); err != nil {
		return nil, err
	}
	return tables, nil
}
func (db Postgres) Select(obj interface{}, para string, filter []Filter) ([]interface{}, error) {
	var tables []interface{}
	con := db.orm
	con = con.Table(fmt.Sprint(obj))
	con = con.Select(para)
	for _, f := range filter {
		con = con.Where(f.Field+" "+f.Compare+" ?", f.Value)
	}
	row, err := con.Rows()
	if err != nil {
		return tables, err
	}
	columns, err := row.Columns()
	if err != nil {
		return tables, err
	}
	for row.Next() {
		values := make([]interface{}, len(columns))
		record := make([]interface{}, len(columns))
		for i, _ := range columns {
			values[i] = &record[i]
		}
		err = row.Scan(values...)
		if err != nil {
			return tables, err
		}
		for _, v := range record {
			var str string
			switch value := v.(type) {
			case int64:
				str = fmt.Sprint(value)
			case float64:
				str = fmt.Sprintf("%f", value)
			default:
				str = fmt.Sprintf("%s", value)
			}
			tables = append(tables, str)
		}
	}
	return tables, nil
}
