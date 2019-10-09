package connectors

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

var dbConnections map[string]*sql.DB
var dbOnce map[string]bool
var dbOnceMutex sync.Mutex

// DbSettings is the struct that is used for registering a connection
type DbSettings struct {
	username string
	password string
	host     string
	port     string
	database string
}

var allDbSettings map[string]DbSettings

// RegisterDb registers a db connection
func RegisterDb(name string, dbSettings DbSettings) {
	allDbSettings[name] = dbSettings
}

// Db returns a connection
func Db(name string) *sql.DB {

	if name == "" {
		panic("DB name cannot be empty")
		return nil
	}

	dbSettings, ok := allDbSettings[name]
	if !ok {
		panic("This DB '" + name + "' was not registered")
		return nil
	}

	if len(dbOnce) == 0 {
		dbOnce = make(map[string]bool, 15)
		dbConnections = make(map[string]*sql.DB, 15)
	}

	dbOnceMutex.Lock()
	if !dbOnce[name] {
		dbOnce[name] = true
		connectionStringArr := []string{
			dbSettings.username,
			":",
			dbSettings.password,
			"@(",
			dbSettings.host,
			":",
			dbSettings.port,
			")/",
			dbSettings.database,
			"?parseTime=true",
		}
		dbConnections[name], _ = sql.Open("mysql", strings.Join(connectionStringArr, ""))
		err := dbConnections[name].Ping()
		if err != nil {
			fmt.Println(err.Error())
		}
		dbConnections[name].SetConnMaxLifetime(time.Second)
		dbOnceMutex.Unlock()
	} else {
		dbOnceMutex.Unlock()
	}

	return dbConnections[name]

}
