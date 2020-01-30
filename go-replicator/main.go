package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	st "git.mokkon.com/soethu/server"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

const (
	NAME    = "whole_sale_report"
	VERSION = "v0.1.0"

	DEFAULT_PORT           = 8080
	DEFAULT_DB_FILE        = "report.db"
	DEFAULT_DATA_DIRECTORY = "data"
	DEFAULT_GOOGLE_JSON    = "./google.json"
	DEFAULT_CONFIG_FILE    = "./config/config.yml"
	DEFAULT_DB_MIGRATION   = true
	DEFAULT_DB_lOGMODE     = true
)

var (
	db            *Db
	fb            *FB
	configFile    Config
	port          *int
	databaseFile  *string
	dataDirectory *string
	googleJsonUrl *string
	dbMigration   *bool
	dbLogmode     *bool
	configFileUrl *string
	dbDriver      string
)

func PrintParameters() {
	fmt.Println("---------------------------------------------------")
	fmt.Println(NAME + "  -  " + VERSION)
	fmt.Println("Database File Name:", *databaseFile)
	fmt.Println("Data Directory:", *dataDirectory)
	fmt.Println("Database Migration:", *dbMigration)
	fmt.Println("Database Logmode:", *dbLogmode)
	fmt.Println("Google Json File:", *googleJsonUrl)
	fmt.Println("Configuration File:", *configFileUrl)

}
func init() {
	port = flag.Int("port", DEFAULT_PORT, "Port")
	databaseFile = flag.String("dbfile", DEFAULT_DB_FILE, "Database file name")
	dataDirectory = flag.String("datadirectory", DEFAULT_DATA_DIRECTORY, "Database directory name")
	dbMigration = flag.Bool("dbmigration", DEFAULT_DB_MIGRATION, "Database migration")
	dbLogmode = flag.Bool("dblogmode", DEFAULT_DB_lOGMODE, "Database logmode")
	googleJsonUrl = flag.String("json", DEFAULT_GOOGLE_JSON, "google json file")
	configFileUrl = flag.String("config", DEFAULT_CONFIG_FILE, "Configuration file")
	flag.Parse()
	PrintParameters()
}

func main() {
	//sqlite3 database
	db = NewDb("sqlite3", *databaseFile, *dbMigration, *dbLogmode)

	f, err := os.Open(*configFileUrl)
	if err != nil {
		log.Print("Open", err.Error())
	}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&configFile)
	if err != nil {
		log.Print("Parse", err.Error())
	}
	fb = runFB(configFile.CredentialJson.Path)
	for _, coll := range configFile.Collection {
		go fb.Listener(coll)
	}
	db.initViews()
	r := mux.NewRouter()
	r.HandleFunc("/api/hb", corsWrap(jsonWrap(heartbeatHandler))).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/meta/{obj}", corsWrap(authWrap(jsonWrap(metaHandler)))).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/objs", corsWrap(authWrap(jsonWrap(objHandler)))).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/resync/{obj}", corsWrap(authWrap(jsonWrap(resyncHandler)))).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/row/{obj}", corsWrap(authWrap(jsonWrap(rowsCountHandler)))).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/data/{obj}", corsWrap(authWrap(jsonWrap(rowsGetHandler)))).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/fts/{obj}/{term}/{limit}", corsWrap(authWrap(jsonWrap(searchHandler)))).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/search/{obj}/{field}/{term}", corsWrap(authWrap(jsonWrap(autoCompleteHandler)))).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/report/{obj}", reportGetHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/chart", corsWrap(authWrap(jsonWrap(chartHandler)))).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/test/{code}", jsonWrap(testHandler)).Methods("GET", "OPTIONS")
	st.Start(strconv.Itoa(*port), r)
}
