package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func ValidFomat(obj string, paser AggParser) (DataParser, error) {
	var d DataParser
	if strings.Contains(paser.Field, " ") {
		return d, errors.New("Fields contain space")
	}
	if strings.Contains(paser.AggFuns, " ") {
		return d, errors.New("Aggeration functions contain space")
	}
	if strings.Contains(paser.GroupBy, " ") {
		return d, errors.New("Group By contain space")
	}
	fArr := strings.Split(paser.Field, ",")
	aggArr := strings.Split(paser.AggFuns, ",")

	if len(fArr) != len(aggArr) {
		return d, errors.New("Mismatch Fields and Aggreation Functions")
	}
	d.Field = QuerySelectFieldFomat(fArr, aggArr)
	d.Filter = QueryDateFileterFomat(obj, paser.Filter)
	d.Limit = paser.Limit
	d.Offset = paser.Offset
	d.Sort = paser.Sort
	d.GroupBy = QueryGroupFomat(paser.GroupBy)
	return d, nil
}
func QueryDateFileterFomat(obj string, f []Filter) []Filter {
	if dbDriver == "sqlite3" {
		if info, ok := db.GetMeta(obj).([]TableInfo); ok {
			if len(f) <= 0 {
				return f
			}
			for fkey, fvalue := range f {
				for _, infoValue := range info {
					if infoValue.Name == fvalue.Field {
						switch infoValue.Type {
						case DEFAULT_SQL_DATE:
							fvalue.Raw = true
						default:
							fvalue.Raw = false
						}
					}
					f[fkey] = fvalue
				}
			}
		}
	}
	for _, fType := range f {
		fmt.Println("Return Filter", reflect.TypeOf(fType.Value))
	}

	return f
}
func QueryGroupFomat(g string) []string {
	return strings.Split(g, ",")
}
func QuerySelectFieldFomat(f []string, agg []string) []string {
	var qArr []string
	for index, fvalue := range f {
		if agg[index] != "" {
			s_field := agg[index] + "(" + fvalue + ") as" + " " + fvalue + "_" + agg[index]
			qArr = append(qArr, s_field)
		} else {
			qArr = append(qArr, fvalue)
		}
	}
	return qArr
}
func reportGetHandler(w http.ResponseWriter, r *http.Request) {
	Cors(w, r)
	if r.Method == "POST" {
		vars := mux.Vars(r)
		obj := vars["obj"]
		if len(obj) == 0 {
			ErrorResponse(w, errors.New("Unknown object"))
			return
		}
		err := checkToken(r)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		err = CheckObj(obj)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		var parser AggParser
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&parser)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		data, err := ValidFomat(obj, parser)
		if err != nil {
			ErrorResponse(w, err)
			return
		} else {
			db.GetReport(obj, data, w)
		}
	} else if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"Ok","message":"","data":}`)
	} else {
		ErrorResponse(w, errors.New("Invalid method"))
		return
	}

}

// get report
func (db Db) GetReport(obj string, filter DataParser, w http.ResponseWriter) {
	isColumn := false
	var d []interface{}
	gstr := strings.Join(filter.GroupBy, ",")
	con := db.orm
	con = con.Table(obj)
	con = con.Select(filter.Field)
	con = con.Group(gstr)
	for _, f := range filter.Filter {
		if f.Raw {
			fValue := f.RawDate()
			con = con.Where(fValue)
		} else {
			con = con.Where(f.Field+" "+f.Compare+" ?", f.Value)
		}
	}

	row, err := con.Rows()
	if err != nil {
		log.Printf("Row Errors %s", err.Error())
		return
	}
	columns, err := row.Columns()
	if err != nil {
		log.Printf("Colums Error %s", err.Error())
		ErrorResponse(w, err)
		return
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
			ErrorResponse(w, err)
			return
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
		d = append(d, data)
		if len(d) > 0 {
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", "attachment;filename=report.csv")
			for _, out := range d {
				strAr := make([]string, len(columns))
				output, ok := out.(map[string]interface{})
				if ok {
					for key, in := range output {
						var i int
						for index, value := range columns {
							if value == key {
								i = index
							}
						}
						strValue := fmt.Sprintf("%s", in)
						strAr[i] = strValue
					}
				}
				wr := csv.NewWriter(w)
				if !isColumn {
					wr.Write(columns)
					isColumn = true
				}
				err = wr.Write(strAr)
				if err != nil {
					http.Error(w, "Error sending csv: "+err.Error(), http.StatusInternalServerError)
					return
				}
				wr.Flush()
			}
			d = nil
			continue
		}
	}
}

func ErrorResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, `{"status":"Error","message":"%s"}`, err)
}
