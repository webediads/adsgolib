package wconnectors

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	// mysql
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/webediads/adsgolib/wconfig"
)

var dbConnections map[string]*sql.DB
var dbMocks map[string]sqlmock.Sqlmock
var dbOnce map[string]bool
var dbOnceMutex sync.Mutex

// DbSettings is the struct that is used for registering a connection
type DbSettings struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
	IsMock   bool
}

var allDbSettings = make(map[string]DbSettings)

// RegisterDb registers a db connection
func RegisterDb(name string) {
	allDbSettings[name] = DbSettings{
		Username: wconfig.Config.GetUnsafe("db", name+".username"),
		Password: wconfig.Config.GetUnsafe("db", name+".password"),
		Host:     wconfig.Config.GetUnsafe("db", name+".host"),
		Port:     wconfig.Config.GetUnsafe("db", name+".port"),
		Database: wconfig.Config.GetUnsafe("db", name+".database"),
		IsMock:   false,
	}
}

// RegisterMockDb registers a mocked db connection
func RegisterMockDb(name string) {
	allDbSettings[name] = DbSettings{
		Username: "mock",
		Password: "mock",
		Host:     "3306",
		Port:     "mock",
		Database: "mock",
		IsMock:   true,
	}
}

// Db returns a connection
func Db(name string) *sql.DB {

	if name == "" {
		panic("DB name cannot be empty")
	}

	dbSettings, ok := allDbSettings[name]
	if !ok {
		panic("This DB '" + name + "' was not registered")
	}

	if len(dbOnce) == 0 {
		dbOnce = make(map[string]bool, 15)
		dbConnections = make(map[string]*sql.DB, 15)
		dbMocks = make(map[string]sqlmock.Sqlmock, 15)
	}

	dbOnceMutex.Lock()
	if !dbOnce[name] {
		dbOnce[name] = true
		if !dbSettings.IsMock {
			connectionStringArr := []string{
				dbSettings.Username,
				":",
				dbSettings.Password,
				"@(",
				dbSettings.Host,
				":",
				dbSettings.Port,
				")/",
				dbSettings.Database,
				"?parseTime=true",
			}
			dbConnections[name], _ = sql.Open("mysql", strings.Join(connectionStringArr, ""))
			err := dbConnections[name].Ping()
			if err != nil {
				fmt.Println(err.Error())
			}
			dbConnections[name].SetConnMaxLifetime(time.Second)
		} else {
			dbConnections[name], dbMocks[name], _ = sqlmock.New()
			err := dbConnections[name].Ping()
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		dbOnceMutex.Unlock()
	} else {
		dbOnceMutex.Unlock()
	}

	return dbConnections[name]

}

// DbMock returns a mock previously defined with Db
func DbMock(name string) sqlmock.Sqlmock {
	return dbMocks[name]
}
