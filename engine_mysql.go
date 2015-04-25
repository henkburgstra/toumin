package toumin

import (
	"fmt"
)

type mysqlDriver struct{}

var MysqlDriver mysqlDriver

func (d mysqlDriver) Name() string {
	return "mysql"
}

func (d mysqlDriver) Version() string {
	return "0.0.1"
}

func (d mysqlDriver) ConnectionString(e *Engine) string {
	return fmt.Sprintf("%s:%s@/%s", e.user, e.password, e.database)
	//return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
	//	engine.User, engine.Password, engine.Host, engine.Port, engine.Database)
}

func (d mysqlDriver) TableNames(e *Engine) []string {
	names := make([]string, 0)
	rows, err := e.db.Query(`SHOW TABLES`)

	if err != nil {
		return names
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return names
		}
		names = append(names, name)
	}
	return names
}

func (d mysqlDriver) TableStructure(e *Engine, name string, entity *Entity) {
	rows, err := e.db.Query(fmt.Sprintf("DESCRIBE %s", name))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		ts := NewModel("fieldstructure")
		ts.Scan(rows)
		f := &EntityField{
			Name:    ts.Field("Field").String(),
			Type:    ts.Field("Type").String(),
			Length:  0, // TODO: lengte distilleren uit het type (varchar(8))
			Key:     ts.Field("Key").String() == "PRI",
			Null:    ts.Field("Null").String() == "YES",
			Default: ts.Field("Default").String(),
		}
		entity.Fields[f.Name] = f
	}
}

func (d mysqlDriver) LoadRelationships(e *Engine, registry *Registry) {
	rows, err := e.db.Query(`
		SELECT rc.CONSTRAINT_NAME AS ConstraintName, 
		rc.TABLE_NAME AS TableName, kc.COLUMN_NAME AS ColumnName, 
		rc.REFERENCED_TABLE_NAME AS ReferencedTableName, 
		kc.REFERENCED_COLUMN_NAME AS ReferencedColumnName, 
		rc.UPDATE_RULE AS UpdateRule, rc.DELETE_RULE AS DeleteRule 
		FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS AS rc
		JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE AS kc
		ON rc.CONSTRAINT_NAME = kc.CONSTRAINT_NAME
		WHERE rc.CONSTRAINT_SCHEMA = ?
		ORDER BY rc.TABLE_NAME, kc.COLUMN_NAME`, e.database)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		m := NewModel("relationship")
		m.Scan(rows)
		r := EntityRelationship{
			ForeignKey:       m.Field("ColumnName").String(),
			ReferencedTable:  m.Field("ReferencedTableName").String(),
			ReferencedColumn: m.Field("ReferencedColumnName").String(),
		}
		entity := registry.Entity(registry.TrimTableAffixes(m.Field("TableName").String()))
		if entity != nil {
			entity.AddRelationship(r)
		}
	}
}
