package toumin

import (
	"fmt"
	"strings"
)

type Cond map[string]map[string]interface{}

type Connective struct {
	Operator string
	Operands []interface{}
}

func And(ops ...interface{}) Connective {
	return Connective{"AND", ops}
}

func Or(ops ...interface{}) Connective {
	return Connective{"OR", ops}
}

type SqlParam struct {
	Operator string
	Value    interface{}
	Values   []interface{}
}

type Selectable struct {
	Entity *Entity
	Field  string
	Param  SqlParam
}

func (e *Selectable) Eq(value interface{}) *Selectable {
	e.Param = SqlParam{Operator: "=", Value: value}
	return e
}

func (e *Selectable) Ne(value interface{}) *Selectable {
	e.Param = SqlParam{Operator: "!=", Value: value}
	return e
}

func (e *Selectable) Gt(value interface{}) *Selectable {
	e.Param = SqlParam{Operator: ">", Value: value}
	return e
}

func (e *Selectable) Gte(value interface{}) *Selectable {
	e.Param = SqlParam{Operator: ">=", Value: value}
	return e
}

func (e *Selectable) Lt(value interface{}) *Selectable {
	e.Param = SqlParam{Operator: "<", Value: value}
	return e
}

func (e *Selectable) Lte(value interface{}) *Selectable {
	e.Param = SqlParam{Operator: "<=", Value: value}
	return e
}

func (e *Selectable) In(values []interface{}) *Selectable {
	e.Param = SqlParam{Operator: "IN", Values: values}
	return e
}

func (e *Selectable) Nin(values []interface{}) *Selectable {
	e.Param = SqlParam{Operator: "NOT IN", Values: values}
	return e
}

type Query struct {
	model    string
	registry *Registry
	sql      string
	filter   []interface{}
	params   []interface{}
}

func NewQuery(model string, registry *Registry) *Query {
	q := new(Query)
	q.model = model
	q.registry = registry
	q.params = make([]interface{}, 0)
	return q
}

func (q *Query) Filter(f ...interface{}) *Query {
	q.filter = f
	return q
}

func (q *Query) processConnective(con Connective) string {
	args := make([]string, 0)

	for _, op := range con.Operands {
		switch c := op.(type) {
		case Connective:
			r := q.processConnective(c)
			if r != "" {
				args = append(args, r)
			}
		case *Selectable:
			r := q.processSelectable(c)
			if r != "" {
				args = append(args, r)
			}
		}
	}
	return fmt.Sprintf("(%s)", strings.Join(args, fmt.Sprintf(" %s ", con.Operator)))
}

func (q *Query) processSelectable(s *Selectable) string {
	if s.Param.Operator == "IN" || s.Param.Operator == "NOT IN" {
		l := make([]string, 0)
		for _, v := range s.Param.Values {
			l = append(l, "?")
			q.params = append(q.params, v)
		}
		return fmt.Sprintf("%s.%s %s (%s)", s.Entity.Name, s.Field, s.Param.Operator, strings.Join(l, ", "))
	} else {
		q.params = append(q.params, s.Param.Value)
		return fmt.Sprintf("%s.%s %s ?", s.Entity.Name, s.Field, s.Param.Operator)
	}
}

func (q *Query) applyFilter() string {
	c := make([]string, 0)

	for _, f := range q.filter {
		switch e := f.(type) {
		case Connective:
			r := q.processConnective(e)
			if r != "" {
				c = append(c, r)
			}
		case *Selectable:
			r := q.processSelectable(e)
			if r != "" {
				c = append(c, r)
			}
		}
	}

	return strings.Join(c, " AND ")
}

func (q *Query) Join() *Query {
	return q
}

func (q *Query) Get(keyValue interface{}) IModel {
	entity := q.registry.Entity(q.model)
	if entity == nil {
		fmt.Println("entity == nil")
		return nil
	}
	key := entity.Key()
	if key == nil {
		fmt.Println("key == nil")
		return nil
	}
	model := q.registry.Model(q.model)(q.model)
	model.SetOwner(model)
	model.SetRegistry(q.registry)

	sql := fmt.Sprintf(`SELECT * 
		FROM %s
		WHERE %s = ?`, entity.Name, key.Name)

	db, err := q.registry.Db()
	if err != nil {
		// TODO: log err
		return nil
	}
	rows, err := db.Query(sql, keyValue)
	if err != nil {
		// TODO: log err
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		model.Scan(rows)
		break
	}

	return model
}

func (q *Query) Columns(cols ...string) *Query {
	return q
}

func (q *Query) FromSql(sql string, params ...interface{}) *Query {
	q.sql = sql
	q.params = append(q.params, params...)
	return q
}

func (q *Query) Sql() string {
	if q.sql != "" {
		return q.sql
	}
	e := q.registry.Entity(q.model)
	if e == nil {
		return ""
	}
	sql := fmt.Sprintf(`SELECT * 
		FROM %s`, e.Name)
	f := q.applyFilter()
	if f != "" {
		sql += fmt.Sprintf("\nWHERE %s", f)
	}

	q.sql = sql

	return sql
}

func (q *Query) All() []IModel {
	models := make([]IModel, 0)
	entity := q.registry.Entity(q.model)
	if entity == nil {
		fmt.Println("Query.All(), entity == nil")
		return models
	}
	key := entity.Key()
	if key == nil {
		fmt.Println("Query.All(), key == nil")
		return models
	}

	fieldPrefix := strings.Replace(q.registry.FieldPrefix(), "{model}", q.model, 1)
	keyName := strings.TrimPrefix(key.Name, fieldPrefix)

	db, err := q.registry.Db()
	if err != nil {
		// TODO: log err
		fmt.Println("Query.All(), db err")
		return models
	}
	rows, err := db.Query(q.Sql(), q.params...)
	if err != nil {
		// TODO: log err
		fmt.Println("Query.All(), query error: ", err.Error())
		return models
	}
	defer rows.Close()

	var lastKey string
	modelConstructor := q.registry.Model(q.model)

	for rows.Next() {
		model := modelConstructor(q.model)
		model.SetOwner(model)
		model.SetRegistry(q.registry)
		model.Scan(rows)
		keyValue := model.Field(keyName).String()
		if lastKey != keyValue {
			models = append(models, model)
			lastKey = keyValue
		}
	}

	return models
}
