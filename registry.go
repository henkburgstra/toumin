package toumin

import (
	"database/sql"
	"strings"
)

type Registry struct {
	engine      *Engine
	entities    map[string]*Entity
	models      map[string]ModelConstructor
	tablePrefix string
	tableSuffix string
	fieldPrefix string
}

func NewRegistry(engine *Engine) *Registry {
	r := new(Registry)
	r.engine = engine
	r.entities = make(map[string]*Entity)
	r.models = make(map[string]ModelConstructor)
	return r
}

func (r *Registry) Entity(name string) *Entity {
	return r.entities[name]
}

func (r *Registry) RegisterEntity(name string, entity *Entity) {
	r.entities[name] = entity
}

func (r *Registry) LoadEntities() {
	engine, err := r.Engine()
	if err != nil {
		return
	}

	for _, name := range engine.TableNames() {
		entity := engine.TableStructure(name)
		r.entities[r.TrimTableAffixes(name)] = entity
		entity.registry = r
	}
	engine.LoadRelationships(r)
}

func (r *Registry) Model(name string) ModelConstructor {
	model, ok := r.models[name]
	if !ok {
		model = NewModel
	}
	return model
}

func (r *Registry) RegisterModel(name string, model ModelConstructor) {
	r.models[name] = model
}

func (r *Registry) Engine() (*Engine, error) {
	if r.engine.connected {
		return r.engine, nil
	}
	_, err := r.engine.Connect()
	if err != nil {
		return nil, err
	}
	return r.engine, nil
}

func (r *Registry) Db() (*sql.DB, error) {
	engine, err := r.Engine()
	if err != nil {
		return nil, err
	}
	return engine.Db(), nil
}

func (r *Registry) TablePrefix() string {
	return r.tablePrefix
}

func (r *Registry) SetTablePrefix(prefix string) {
	r.tablePrefix = prefix
}

func (r *Registry) TrimTablePrefix(name string) string {
	return strings.TrimPrefix(name, r.tablePrefix)
}

func (r *Registry) TableSuffix() string {
	return r.tableSuffix
}

func (r *Registry) SetTableSuffix(suffix string) {
	r.tableSuffix = suffix
}

func (r *Registry) TrimTableSuffix(name string) string {
	return strings.TrimSuffix(name, r.tableSuffix)
}

func (r *Registry) TrimTableAffixes(name string) string {
	return r.TrimTableSuffix(r.TrimTablePrefix(name))
}

func (r *Registry) FieldPrefix() string {
	return r.fieldPrefix
}

func (r *Registry) SetFieldPrefix(prefix string) {
	r.fieldPrefix = prefix
}

func (r *Registry) TrimFieldPrefix(name string) string {
	return strings.TrimPrefix(name, r.fieldPrefix)
}

func (r *Registry) Query(modelName string) *Query {
	return NewQuery(modelName, r)
}
