package toumin

// Toumin - Japans voor Hibernate

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unsafe"
)

// ModelConstructor defines the signature of Model constructors.
type ModelConstructor func(string) IModel

// Underscore2Camel translates strings with underscores
// to strings with camel case: the_word -> TheWord.
func Underscore2Camel(underscores string) string {
	els := strings.Split(strings.ToLower(underscores), "_")

	for i, el := range els {
		r := []rune(el)
		r[0] = unicode.ToUpper(r[0])
		els[i] = string(r)
	}
	return strings.Join(els, "")
}

// IModel defines the interface that Models need to implement.
type IModel interface {
	Name() string
	Entity() *Entity
	FieldMapping(string) string
	SetFieldMapping(string, string)
	Field(string) *FieldValue
	Fields() FieldData
	FieldNames() []string
	Ref(string) (IModel, bool)
	BackRef(string, ...string) *Query
	Owner() IModel
	SetOwner(IModel)
	Registry() *Registry
	SetRegistry(*Registry)
	Scan(*sql.Rows)
}

// Model defines the default Model. Implements IModel.
type Model struct {
	name         string
	registry     *Registry
	fields       FieldData
	fieldMapping map[string]string
	owner        IModel
}

// NewModel constructs a new Model instance.
func NewModel(name string) IModel {
	m := new(Model)
	m.name = name
	m.fields = make(FieldData)
	m.fieldMapping = make(map[string]string)
	m.owner = m
	return m
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) Entity() *Entity {
	if m.registry == nil {
		return nil
	}
	return m.registry.Entity(m.name)
}

func (m *Model) FieldNames() []string {
	names := make([]string, 0)
	for name, _ := range m.fields {
		names = append(names, name)
	}
	return names
}

func (m *Model) Field(name string) *FieldValue {
	field, ok := m.fields[name]
	if ok {
		return field
	}
	field = &FieldValue{}
	field.Set("")
	return field
}

func (m *Model) FieldMapping(name string) string {
	return m.fieldMapping[name]
}

func (m *Model) SetFieldMapping(modelField, entityField string) {
	m.fieldMapping[modelField] = entityField
}

func (m *Model) Fields() FieldData {
	return m.fields
}

func (m *Model) Registry() *Registry {
	return m.registry
}

func (m *Model) SetRegistry(r *Registry) {
	m.registry = r
}

func (m *Model) Owner() IModel {
	return m.owner
}

func (m *Model) SetOwner(owner IModel) {
	m.owner = owner
}

func (m *Model) Key() *FieldValue {
	r := m.Registry()
	if r == nil {
		return nil
	}
	e := m.Entity()
	if e == nil {
		return nil
	}
	key := e.Key()
	if key == nil {
		return nil
	}

	fieldPrefix := strings.Replace(r.FieldPrefix(), "{model}", m.Name(), 1)
	keyName := strings.TrimPrefix(key.Name, fieldPrefix)
	return m.Field(keyName)
}

// Ref returns the model that foreign key fk refers to.
func (m *Model) Ref(fk string) (IModel, bool) {
	entity := m.Entity()
	if entity == nil {
		fmt.Println("Model.Ref(): geen entity")
		return m, false
	}
	// TODO: support multiple foreign key name schemes
	fkName := fmt.Sprintf("%s_%s", m.Name(), fk)
	relationship, ok := entity.Relationship(fkName)
	if !ok {
		return m, false
	}
	fkValue := m.Field(fk).Get()
	registry := m.Registry()
	if registry == nil {
		return m, false
	}
	refModel := registry.TrimTableAffixes(relationship.ReferencedTable)
	fkModel := registry.Query(refModel).Get(fkValue)

	return fkModel, true
}

// BackRef returns a Query instance representing all records
// of br referring to m.
// Unless fks is provided, m's primary key is used as foreign key.
func (m *Model) BackRef(br string, fks ...string) *Query {
	r := m.Registry()
	if r == nil {
		return &Query{}
	}
	e := r.Entity(br)
	if e == nil {
		return &Query{}
	}
	var fk string
	if len(fks) > 0 {
		fk = fks[0]
	} else {
		fk = fmt.Sprintf("%s_%s", br, m.Name())
	}
	key := m.Key()
	if key == nil {
		return &Query{}
	}
	q := r.Query(br).Filter(e.Col(fk).Eq(key.Get()))

	return q
}

func (m *Model) Scan(rows *sql.Rows) {
	refElem := reflect.ValueOf(m.owner).Elem()
	registry := m.Registry()
	var entity *Entity
	fieldPrefix := ""

	if registry != nil {
		fieldPrefix = strings.Replace(registry.FieldPrefix(), "{model}", m.Name(), 1)
		entity = registry.Entity(m.Name())
	}

	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	pValues := make([]interface{}, len(columns))

	for i := range columns {
		pValues[i] = &values[i]
	}

	rows.Scan(pValues...)

	for i := range columns {
		belongs := true
		// if the model is registered, only use the fields that belong to this model.
		if entity != nil {
			field := entity.Fields[columns[i]]
			if field == nil {
				belongs = false
			}
		}
		if belongs {
			value := new(FieldValue)
			modelAttr := Underscore2Camel(strings.TrimPrefix(columns[i], fieldPrefix))
			// Check if the model has an attribute that matches the name
			// of the column. Underscores are  translated to CamelCase:
			// the_name -> TheName
			structField := refElem.FieldByName(modelAttr)

			if structField.IsValid() {
				x := ""
				scanValue := &x

				switch s := values[i].(type) {
				case string:
					scanValue = &s
				case []byte:
					*scanValue = string(s)
				case int, int16, int32, int64:
					*scanValue = fmt.Sprintf("%d", s)
				default:
					fmt.Println("?? onbekend scan type ??")
				}

				switch structField.Interface().(type) {
				case int:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (*int)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					*up = x
				case int8:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (*int8)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					*up = (int8)(x)
				case int16:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (*int16)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					*up = (int16)(x)
				case int32:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (*int32)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					*up = (int32)(x)
				case int64:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (*int64)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					*up = (int64)(x)
				case *int:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (**int)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					**up = x
				case *int8:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (**int8)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					**up = (int8)(x)
				case *int16:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (**int16)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					**up = (int16)(x)
				case *int32:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (**int32)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					**up = (int32)(x)
				case *int64:
					x, err := strconv.Atoi(*scanValue)
					if err != nil {
						x = 0
					}
					up := (**int64)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					**up = (int64)(x)
				case string:
					up := (*string)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					*up = *scanValue
				case *string:
					up := (**string)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					**up = *scanValue
				case []byte:
					up := (*[]byte)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					*up = []byte(*scanValue)
				case *[]byte:
					up := (**[]byte)(unsafe.Pointer(structField.UnsafeAddr()))
					value.SetAddr(up)
					**up = []byte(*scanValue)
				case nil:
				}

			} else {
				if values[i] != nil {
					//fmt.Println(string(values[i].([]byte)))
					value.SetAddr(values[i])
					//fmt.Println(value.String())
				}
			}

			modelField := strings.TrimPrefix(columns[i], fieldPrefix)
			m.SetFieldMapping(modelField, columns[i])
			m.Fields()[modelField] = value
		}
	}

}
