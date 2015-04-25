package toumin

import (
	"database/sql"
	//	"fmt"
)

type IEngineDriver interface {
	Name() string
	Version() string
	ConnectionString(*Engine) string
	TableNames(*Engine) []string
	TableStructure(engine *Engine, name string, entity *Entity)
	LoadRelationships(e *Engine, registry *Registry)
}

type Engine struct {
	db        *sql.DB
	driver    IEngineDriver
	host      string
	port      int
	database  string
	user      string
	password  string
	connected bool
}

func NewEngine(d IEngineDriver) *Engine {
	e := new(Engine)
	e.driver = d
	return e
}

func (e *Engine) Db() *sql.DB {
	return e.db
}

func (e *Engine) Driver() IEngineDriver {
	return e.driver
}

func (e *Engine) SetHost(h string) {
	e.host = h
}

func (e *Engine) Host() string {
	return e.host
}

func (e *Engine) SetPort(p int) {
	e.port = p
}

func (e *Engine) Port() int {
	return e.port
}

func (e *Engine) SetDatabase(d string) {
	e.database = d
}

func (e *Engine) Database() string {
	return e.database
}

func (e *Engine) SetUser(u string) {
	e.user = u
}

func (e *Engine) User() string {
	return e.user
}

func (e *Engine) SetPassword(p string) {
	e.password = p
}

func (e *Engine) Password() string {
	return e.password
}

//func (engine *Engine) ConnectionString() string {
//	switch engine.Driver {
//	case "sqlite3":
//		return engine.Database
//	case "mysql":
//		return fmt.Sprintf("%s:%s@/%s", engine.User, engine.Password, engine.Database)
//		//return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
//		//	engine.User, engine.Password, engine.Host, engine.Port, engine.Database)
//	case "mssql":
//		return fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;encrypt=disable",
//			engine.Host, engine.User, engine.Password, engine.Database)
//	default:
//		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
//			engine.User, engine.Password, engine.Host, engine.Port, engine.Database)
//	}
//}

func (e *Engine) Connect() (*sql.DB, error) {
	db, err := sql.Open(e.driver.Name(), e.driver.ConnectionString(e))
	if err == nil {
		e.db = db
		e.connected = true
	}
	return db, err
}

func (e *Engine) LoadRelationships(registry *Registry) {
	e.driver.LoadRelationships(e, registry)
}

func (e *Engine) TableNames() []string {
	return e.driver.TableNames(e)
}

func (e *Engine) TableStructure(name string) *Entity {
	entity := NewEntity(name)
	e.driver.TableStructure(e, name, entity)
	return entity
}
