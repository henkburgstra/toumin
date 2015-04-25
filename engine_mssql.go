package toumin

import (
	"fmt"
)

type mssqlDriver struct{}

var MssqlDriver mssqlDriver

func (d mssqlDriver) Name() string {
	return "mssql"
}

func (d mssqlDriver) Version() string {
	return "0.0.1"
}

func (d mssqlDriver) ConnectionString(e *Engine) string {
	return fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;encrypt=disable",
		e.host, e.user, e.password, e.database)
}

func (d mssqlDriver) TableNames(e *Engine) []string {
	names := make([]string, 0)
	rows, err := e.db.Query(`SELECT table_name FROM information_schema.tables
		WHERE table_type = 'BASE TABLE'`)

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

func (d mssqlDriver) TableStructure(e *Engine, name string, entity *Entity) {
	rows, err := e.db.Query(`select COLUMN_NAME, DATA_TYPE, IS_NULLABLE,
		COLUMN_DEFAULT 
 		from information_schema.columns 
 		where table_name = ?
 		order by ordinal_position`, name)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		ts := NewModel("fieldstructure")
		ts.Scan(rows)
		f := &EntityField{
			Name:   ts.Field("COLUMN_NAME").String(),
			Type:   ts.Field("DATA_TYPE").String(),
			Length: 0,
			//Key:     ts.Field("Key").String() == "PRI",
			Null:    ts.Field("IS_NULLABLE").String() == "YES",
			Default: ts.Field("COLUMN_DEFAULT").String(),
		}
		entity.Fields[f.Name] = f
	}
}

func (d mssqlDriver) LoadRelationships(e *Engine, registry *Registry) {
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
