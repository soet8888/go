package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"

	"github.com/gorilla/mux"
)

type Status struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Type    string      `json:"type"`
}

func NewOkStatus() *Status {
	return &Status{Status: "Ok"}
}
func NewErrorStatus() *Status {
	return &Status{Status: "Error"}
}

func (s *Status) SetData(data interface{}) *Status {
	s.Data = data
	return s
}

func (s *Status) SetMessage(message string) *Status {
	s.Message = message
	return s
}
func heartbeatHandler(w http.ResponseWriter, r *http.Request) Return {
	var m *runtime.MemStats = new(runtime.MemStats)
	runtime.ReadMemStats(m)

	hb := &Heartbeat{Status: "Ok", Message: "Heartbeat!", Mem: m.Alloc,
		AppName: NAME, AppVersion: VERSION}
	return hb
}
func autoCompleteHandler(w http.ResponseWriter, r *http.Request) Return {
	if r.Method == "GET" {
		vars := mux.Vars(r)
		obj := vars["obj"]
		field := vars["field"]
		termstr := vars["term"]
		s, err := url.QueryUnescape(termstr)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		term, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		if len(obj) == 0 {
			return NewErrorStatus().SetMessage("Unknown object")
		}
		err = CheckObj(obj)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		data, err := db.GetSearchTerm(obj, field, string(term), 10)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		return NewOkStatus().SetData(data)
	} else if r.Method == "OPTIONS" {
		return NewOkStatus()
	}
	return NewErrorStatus().SetMessage("Invalid Method")
}
func searchHandler(w http.ResponseWriter, r *http.Request) Return {
	if r.Method == "GET" {
		vars := mux.Vars(r)
		obj := vars["obj"]

		if len(obj) == 0 {
			return NewErrorStatus().SetMessage("Unknown object")
		}
		lim := vars["limit"]
		limit, err := strconv.ParseUint(lim, 10, 64)
		if err != nil {
			return NewErrorStatus().SetMessage("Invalid Limit")
		}
		termstr := vars["term"]
		s, err := url.QueryUnescape(termstr)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		term, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		if len(obj) == 0 {
			return NewErrorStatus().SetMessage("Unknown object")
		}
		fObj := obj + "_table"
		data, err := GetFullTextSearch(fObj, string(term), limit)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		fData := GetFullTextSearchData(obj, data)
		return NewOkStatus().SetData(fData)
	} else if r.Method == "OPTIONS" {
		return NewOkStatus()
	}
	return NewErrorStatus().SetMessage("Invalid Method")
}
func metaHandler(w http.ResponseWriter, r *http.Request) Return {
	if r.Method == "GET" {
		vars := mux.Vars(r)
		obj := vars["obj"]
		if len(obj) == 0 {
			return NewErrorStatus().SetMessage("Unknown object")
		}
		err := CheckObj(obj)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		data := db.GetMeta(obj)
		return NewOkStatus().SetData(data)
	} else if r.Method == "OPTIONS" {
		return NewOkStatus()
	}
	return NewErrorStatus().SetMessage("Invalid Method")
}
func chartHandler(w http.ResponseWriter, r *http.Request) Return {
	if r.Method == "GET" {
		data, err := db.GetAll("po_product_view")
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		return NewOkStatus().SetData(data)
	} else if r.Method == "OPTIONS" {
		return NewOkStatus()
	}
	return NewErrorStatus().SetMessage("Invalid Method")
}
func objHandler(w http.ResponseWriter, r *http.Request) Return {
	if r.Method == "GET" {
		objs := getObjects()
		if len(objs) <= 0 {
			return NewErrorStatus().SetMessage("Empty Object.")
		}
		return NewOkStatus().SetData(objs)
	} else if r.Method == "OPTIONS" {
		return NewOkStatus()
	}
	return NewErrorStatus().SetMessage("Invalid Method")
}
func resyncHandler(w http.ResponseWriter, r *http.Request) Return {
	if r.Method == "POST" {
		vars := mux.Vars(r)
		obj := vars["obj"]
		if len(obj) == 0 {
			return NewErrorStatus().SetMessage("Unknown object")
		}
		err := CheckObj(obj)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		err = resyncData(obj)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		return NewOkStatus()
	} else if r.Method == "OPTIONS" {

		return NewOkStatus()
	}
	return NewErrorStatus().SetMessage("Invalid Method")
}
func rowsCountHandler(w http.ResponseWriter, r *http.Request) Return {
	if r.Method == "POST" {
		vars := mux.Vars(r)
		obj := vars["obj"]
		if len(obj) == 0 {
			return NewErrorStatus().SetMessage("Unknown object")
		}
		err := CheckObj(obj)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		var parser AggParser
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&parser)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		data, err := ValidFomat(obj, parser)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		count, err := db.GetCount(obj, data)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		return NewOkStatus().SetData(count)
	} else if r.Method == "OPTIONS" {
		return NewOkStatus()
	}
	return NewErrorStatus().SetMessage("Invalid Method")
}
func rowsGetHandler(w http.ResponseWriter, r *http.Request) Return {
	if r.Method == "POST" {
		vars := mux.Vars(r)
		obj := vars["obj"]
		if len(obj) == 0 {
			return NewErrorStatus().SetMessage("Unknown object")
		}
		err := CheckObj(obj)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		var parser AggParser
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&parser)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		data, err := ValidFomat(obj, parser)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		d, err := db.GetFilterValue(obj, data)
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		return NewOkStatus().SetData(d)
	} else if r.Method == "OPTIONS" {
		return NewOkStatus()
	}
	return NewErrorStatus().SetMessage("Invalid Method")
}

func resyncData(obj string) error {
	var collection Collection
	for _, coll := range configFile.Collection {
		p := GetObj(coll.Path)
		if p == obj {
			collection = coll
		}
	}
	if len(collection.Path) <= 0 || collection.Path == "" {
		return errors.New("Invalid collection path.")
	}
	q1 := `drop table if exists ` + obj + `;`
	q2 := `drop table if exists ` + obj + `_table ;`
	db.orm.Exec(q1)
	db.orm.Exec(q2)
	go fb.Listener(collection)
	return nil
}
func GetFullTextSearchData(obj string, data []interface{}) []interface{} {
	var fData []interface{}
	for _, outV := range data {
		if vv, ok := outV.(map[string]interface{}); ok {
			for inK, inV := range vv {
				if inK == "id" {
					buy, _ := db.GetById(obj, fmt.Sprintf("%s", inV))
					fData = append(fData, buy)
				}

			}
		}
	}
	fmt.Println("fData", fData)
	return fData
}
func testHandler(w http.ResponseWriter, r *http.Request) Return {
	if r.Method == "GET" {
		vars := mux.Vars(r)
		testCode := vars["code"]
		if testCode != "st" {
			return NewErrorStatus().SetMessage("Wrong test code.")
		}
		data, err := GetFullTextSearch("buyers_table", "Buyer", uint64(4))
		if err != nil {
			return NewErrorStatus().SetMessage(err.Error())
		}
		db.GetCount("buyers", DataParser{
			Filter: []Filter{Filter{Field: "reg_date", Compare: ">", Value: "2020-01-08", Raw: true}, Filter{Field: "biz_name", Compare: ">", Value: 20302}},
		})
		for _, outV := range data {
			if vv, ok := outV.(map[string]interface{}); ok {
				for inK, inV := range vv {
					if inK == "id" {
						buy, err := db.GetById("buyers", fmt.Sprintf("%s", inV))
						if err != nil {
							fmt.Println("Buyer Error", err.Error())
						}
						data = append(data, buy)
					}
				}
			}
		}
		test2, err := db.GetFilterValue("po_product_view", DataParser{Limit: 20, GroupBy: []string{"user_name"}})
		if err != nil {
			fmt.Println("Error Po", err.Error())
		}
		fmt.Println("Po Data", test2)

		fmt.Println("Objects", getObjects())

		fmt.Println("Meta PO", db.GetMeta("po_product_view"))
		return NewOkStatus().SetData(data)
	}
	return NewErrorStatus().SetMessage("Invalid Method")
}
