package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"strings"
	"sync"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FB struct {
	client     *firestore.Client
	ctx        context.Context
	authClient *auth.Client
	mutex      *sync.Mutex
}

func runFB(jsonPath string) *FB {
	var fb *FB
	ctx := context.Background()
	sa := option.WithCredentialsFile(jsonPath)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
		panic("Faile to connect firebase")
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
		panic("Failed to start client connection")
	}
	auth, err := app.Auth(ctx)
	if err != nil {
		log.Fatalln(err)
		panic("Failed to auth")
	}
	fb = &FB{client: client, ctx: ctx, authClient: auth, mutex: &sync.Mutex{}}
	return fb
}

func (fb FB) closeFB() {
	defer fb.client.Close()
}

func (fb FB) Listener(collection Collection) {
	path, f, fts, ftsFields, groupCollection := collection.Path, collection.Filter, collection.FTS, collection.FtsField, collection.IsGroupCollection
	iter := fb.MakeIterator(path, f, groupCollection)
	count := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			log.Printf("[Object] %s [ERROR] %s", GetObj(path), err.Error())
			break
		}
		if doc == nil {
			log.Printf("[Object] %s [ERROR] %s", GetObj(path), err.Error())
			break
		}
		m := doc.Changes
		for _, ff := range m {
			obj, ft := FixObjectAndFields(GetObj(path), ff.Doc.Data())
			if count != -1 {
				if collection.IsGroupCollection {
					for index, unique := range ft {
						if unique.Field == "id" {
							unique.Association = ""
							ft[index] = unique
						}
					}
				}
				if err := MakeObj(obj, ft); err != nil {
					log.Printf("[Object] %s [ERROR] %s", GetObj(path), err.Error())
					break
				} else {
					count = -1
				}
				if fts {
					if err := createTrigger(GetObj(path), ftsFields); err != nil {
						log.Printf("[Object] %s [ERROR] %s", GetObj(path), err.Error())
					}
				}
			}
			switch ff.Kind {
			case 0:
				log.Printf("[Object] %s  [Action] %s", GetObj(path), "INSERT")
				if err := InsertObj(GetObj(path), ff.Doc.Data(), ff.Doc.Ref.ID, ff.Doc.Ref.Path); err != nil {
					AlterObjColumn(GetObj(path), ff.Doc.Data())
					InsertObj(GetObj(path), ff.Doc.Data(), ff.Doc.Ref.ID, ff.Doc.Ref.Path)
				}
			case 1:
				log.Printf("[Object] %s  [Action] %s", GetObj(path), "DELETE")
				if err := DeleteRecord(GetObj(path), ff.Doc.Ref.Path); err != nil {
					log.Printf("Object", GetObj(path), "Error Delete", err.Error())
				}
			case 2:
				log.Printf("[Object] %s  [Action] %s", GetObj(path), "UPDATE")
				if err := UpdateRecord(GetObj(path), ff.Doc.Ref.ID, ff.Doc.Ref.Path, ff.Doc.Data()); err != nil {
					AlterObjColumn(GetObj(path), ff.Doc.Data())
					UpdateRecord(GetObj(path), ff.Doc.Ref.ID, ff.Doc.Ref.Path, ff.Doc.Data())
				}
			}
		}
		if err != nil {
			log.Printf("[Object] %s [ERROR] %v", GetObj(path), err.Error())
			break
		}
	}
}
func (fb FB) MakeIterator(path string, f []Filter, gpCollection bool) *firestore.QuerySnapshotIterator {
	var iter *firestore.QuerySnapshotIterator
	if gpCollection {
		if len(f) == 1 {
			iter = fb.client.CollectionGroup(path).
				Where(f[0].Field, f[0].Compare, f[0].Value).Snapshots(fb.ctx)
		}
		if len(f) == 2 {
			iter = fb.client.CollectionGroup(path).
				Where(f[0].Field, f[0].Compare, f[0].Value).
				Where(f[1].Field, f[1].Compare, f[1].Value).
				Snapshots(fb.ctx)
		}
		if len(f) == 3 {
			iter = fb.client.CollectionGroup(path).
				Where(f[0].Field, f[0].Compare, f[0].Value).
				Where(f[1].Field, f[1].Compare, f[1].Value).
				Where(f[2].Field, f[2].Compare, f[2].Value).
				Snapshots(fb.ctx)
		}
		if len(f) <= 0 {
			iter = fb.client.CollectionGroup(path).Snapshots(fb.ctx)
		}
	} else {
		if len(f) == 1 {
			iter = fb.client.Collection(path).
				Where(f[0].Field, f[0].Compare, f[0].Value).Snapshots(fb.ctx)
		}
		if len(f) == 2 {
			iter = fb.client.Collection(path).
				Where(f[0].Field, f[0].Compare, f[0].Value).
				Where(f[1].Field, f[1].Compare, f[1].Value).
				Snapshots(fb.ctx)
		}
		if len(f) == 3 {
			iter = fb.client.Collection(path).
				Where(f[0].Field, f[0].Compare, f[0].Value).
				Where(f[1].Field, f[1].Compare, f[1].Value).
				Where(f[2].Field, f[2].Compare, f[2].Value).
				Snapshots(fb.ctx)
		}
		if len(f) <= 0 {
			iter = fb.client.Collection(path).Snapshots(fb.ctx)
		}
	}
	return iter
}
func FixObjectAndFields(obj string, data map[string]interface{}) (string, []FieldType) {
	var ft []FieldType
	var typ string
	if len(data) > 0 {
		ft = append(ft, FieldType{Field: "id", Type: DEFAULT_SQL_STRING, Association: "UNIQUE"})
		ft = append(ft, FieldType{Field: "path", Type: DEFAULT_SQL_STRING, Association: "UNIQUE"})
	}
	for key, value := range data {
		switch value.(type) {
		case int64, int:
			typ = DEFAULT_SQL_INT
		case float64, float32:
			typ = DEFAUTL_SQL_FLOAT
		case bool:
			typ = DEAFULT_SQL_BOOL
		case time.Time:
			typ = DEFAULT_SQL_DATE
		default:
			typ = DEFAULT_SQL_STRING
		}
		ft = append(ft, FieldType{Field: key, Type: typ})
	}
	return obj, ft
}
func SqlDataType(value interface{}) string {
	var typ string
	switch value.(type) {
	case int64, int:
		typ = DEFAULT_SQL_INT
		break
	case float64, float32:
		typ = DEFAUTL_SQL_FLOAT
		break
	case bool:
		typ = DEAFULT_SQL_BOOL
		break
	case time.Time:
		typ = DEFAULT_SQL_DATE
		break
	default:
		typ = DEFAULT_SQL_STRING
		break
	}
	return typ
}
func MakeObj(obj string, ftype []FieldType) error {
	var q []string
	if len(ftype) < 0 {
		return errors.New("Cannot creat without at least one field")
	}
	for _, f := range ftype {
		if strings.Contains(f.Field, " ") || f.Type == "" {
			return errors.New("Field contains space" + f.Field)
		}
		qfix := "\"" + f.Field + "\"" + " " + f.Type
		if f.Association == "UNIQUE" {
			qfix = "\"" + f.Field + "\"" + " " + f.Type + " " + "NOT NULL UNIQUE"
		}
		q = append(q, qfix)
	}
	query := `drop table if exists ` + obj + ` ; create table ` + obj + `(` + strings.Join(q, ",") + `)`
	err := db.orm.Exec(query).Error
	if err != nil {
		return err
	}
	return nil
}
func GetObj(path string) string {
	if strings.Contains(path, "/") {
		p := strings.Split(path, "/")
		return p[len(p)-1]
	}
	return path
}
func AlterObjColumn(obj string, data map[string]interface{}) error {
	var oldField []string
	var newField []string
	objMap := getObjectsAndFields()
	if objMapValue, ok := objMap[obj]; ok {
		if f, ok := objMapValue.([]string); ok {
			oldField = f
		}
	}
	for key, _ := range data {
		newField = append(newField, key)
	}
	alterField := MatchField(oldField, newField)
	if len(alterField) > 0 {
		for _, field := range alterField {
			fieldValue := data[field]
			dataType := SqlDataType(fieldValue)
			query := `alter table ` + obj + ` add ` + field + ` ` + dataType + ` ;`
			err := db.orm.Exec(query).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func MatchField(oldField []string, newField []string) []string {
	var alterField []string
	for _, n := range newField {
		if !find(oldField, n) {
			alterField = append(alterField, n)
		}
	}
	return alterField
}

func InsertObj(obj string, data map[string]interface{}, docID, path string) error {
	var attrArr []string
	var valueArr []string
	index := 0
	attrArr = make([]string, len(data)+2)
	valueArr = make([]string, len(data)+2)
	attrArr[index] = "id"
	valueArr[index] = "\"" + docID + "\""
	index = index + 1
	attrArr[index] = "path"
	valueArr[index] = "\"" + path + "\""
	index = index + 1
	for key, value := range data {
		var v string
		switch value.(type) {
		case int64, int:
			v = fmt.Sprintf("%d", value)
		case float64, float32:
			v = fmt.Sprintf("%f", value)
		case bool:
			v = fmt.Sprintf("%t", value)
		case time.Time:
			if timeValue, ok := value.(time.Time); ok {
				v = timeValue.Format(time.RFC3339)
			}
		default:
			v = fmt.Sprintf("%s", value)
		}
		if key != "" || v != "" {
			attrArr[index] = key
			valueArr[index] = "\"" + v + "\""
		}
		index++
	}
	query := `insert into ` + obj + ` (` + strings.Join(attrArr, ",") + `) values ( ` + strings.Join(valueArr, ",") + `);`
	err := db.orm.Exec(query).Error
	if err != nil {
		return err
	}

	return nil
}
func DeleteRecord(obj string, path string) error {
	path = "\"" + path + "\""
	query := `delete from ` + obj + ` where path=` + path + `;`
	err := db.orm.Exec(query).Error
	if err != nil {
		return err
	}
	return nil
}
func UpdateRecord(obj string, docID, path string, para map[string]interface{}) error {
	for key, value := range para {
		if vv, ok := value.(map[string]interface{}); ok {
			para[key] = fmt.Sprintf("%v", vv)
		}
	}
	para["id"] = docID
	para["path"] = path
	tx := db.orm.Begin()
	err := tx.Table(obj).Where("path = ?", path).Updates(para).Error
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
func find(arr []string, key string) bool {
	for _, finder := range arr {
		if finder == key {
			return true
		}
	}
	return false
}
func createTrigger(obj string, ftsFields []Field) error {
	var fields []string
	objMap := getObjectsAndFields()
	if objMapValue, ok := objMap[obj]; ok {
		if f, ok := objMapValue.([]string); ok {
			fields = f
		}
	}
	if len(ftsFields) > 0 {
		var ftsfs []string
		for _, ftsf := range ftsFields {
			if !find(fields, ftsf.Name) {
				return errors.New("Invalid FTS fields" + ftsf.Name)
			}
			ftsfs = append(ftsfs, ftsf.Name)
		}
		fields = ftsfs
	}
	if !find(fields, "id") {
		fields = append(fields, "id")
	}
	inFieldValue := make([]string, len(fields))
	for index, value := range fields {
		value = "new." + value
		inFieldValue[index] = value
	}
	upFeildValue := make([]string, len(fields))
	for index, value := range fields {
		value = value + "=new." + value
		upFeildValue[index] = value
	}
	fields = append(fields, "action")
	inStatus := "\"insert\""
	tableName := obj + "_" + "table"
	upTriggerName := obj + "_au"
	inTriggerName := obj + "_ai"
	delTriggerName := obj + "_ad"
	vString := `drop table if exists ` + tableName + ` ; create virtual table ` + tableName + ` using fts5(` + strings.Join(fields, ",") + `)`
	inString := `drop trigger if exists ` + inTriggerName + `; 
	create trigger ` + inTriggerName + ` after insert 
	on ` + obj + ` begin 
	insert into ` + tableName + `(` + strings.Join(fields, ",") + `) 
	values(` + strings.Join(inFieldValue, ",") + `,` + inStatus + `); 
	end;`
	delString := `drop trigger if exists ` + delTriggerName + `;
	create trigger ` + delTriggerName + ` 
	after delete 
	on ` + obj + ` begin delete from 
	` + tableName + ` where id=old.id; end;`

	upString := `drop trigger if exists ` + upTriggerName + `;
	create trigger ` + upTriggerName + `
	after update 
	on ` + obj + ` begin update ` + tableName + ` 
	set ` + strings.Join(upFeildValue, ",") + ` 
	where id=new.id; end;`
	err := db.orm.Exec(vString).Error
	if err != nil {
		return err
	}
	err = db.orm.Exec(inString).Error
	if err != nil {
		return err
	}
	err = db.orm.Exec(delString).Error
	if err != nil {
		return err
	}
	err = db.orm.Exec(upString).Error
	if err != nil {
		return err
	}
	return nil
}
func (fb FB) VerifedToken(token string) error {
	_, err := fb.authClient.VerifyIDToken(fb.ctx, token)
	if err != nil {
		return err
	}
	return nil
}
func checkToken(r *http.Request) error {
	token := r.Header.Get("Token")
	if token == "" {
		return errors.New("Request token is null.")
	}
	err := fb.VerifedToken(token)
	if err != nil {
		return err
	}
	return nil
}
