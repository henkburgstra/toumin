package filter

import (
	"fmt"
	"strings"
)


func Values(items []interface{}) []interface{} {
	values := make([]interface{}, 0)
	
	for _, op := range items {
		switch v := op.(type) {
		case Connective:
			values = append(values, v.Values()...)
		case Selectable:
			values = append(values, v.Values()...)
		}
	}
	return values
}

type Filter struct {
	Params []interface{}
	Values []interface{}	
}

// NewFilter expects zero or more arguments of the type *Selectable
// or Connective and returns *Filter
func NewFilter(args ...interface{}) *Filter {
	f := new(Filter)
	f.Params = make([]interface{}, 0)
	for _, arg := range args {
		f.Params = append(f.Params, arg)
	}
	return f
}

func (f *Filter) String() string {
	return Translater{}.Translate(f)	
}

func (f *Filter) Test(item interface{}, t FilterTranslater) bool {
	return true
}

type FilterTranslater interface {
	Translate(f *Filter) string
	TranslateOperator(o string) string
	TranslateEntity(e string) string
	TranslateField(e, f string) string
	ProcessConnective(f *Filter, con Connective) string
	ProcessSelectable(f *Filter, s Selectable) string
}

type Translater struct {}

func (t Translater) ProcessConnective(f *Filter, con Connective) string {
	args := make([]string, 0)

	for _, op := range con.Operands {
		switch c := op.(type) {
		case Connective:
			r := t.ProcessConnective(f, c)
			if r != "" {
				args = append(args, r)
			}
		case Selectable:
			r := t.ProcessSelectable(f, c)
			if r != "" {
				args = append(args, r)
			}
		}
	}
	return fmt.Sprintf("(%s)", strings.Join(args, fmt.Sprintf(" %s ", con.Operator)))
}

func (t Translater) TranslateOperator(o string) string {
	switch o {
	case "EQ":
		return "="
	case "NE":
		return "!="
	case "GT":
		return ">"
	case "GTE":
		return ">="
	case "LT":
		return "<"
	case "LTE":
		return "<="
	case "IN":
		return "IN"
	case "NIN":
		return "NOT IN"
	case "PFX":
		return "LIKE"
	case "SFX":
		return "LIKE"
	}
	return "!OPERATOR ERROR!"
}

func (t Translater) TranslateEntity(e string) string {
	return e
}

func (t Translater) TranslateField(e, f string) string {
	return f
}

func (t Translater) ProcessSelectable(f *Filter, s Selectable) string {
	operator := t.TranslateOperator(s.Param.Operator)
	entity := t.TranslateEntity(s.Entity)
	field := t.TranslateField(s.Entity, s.Field)
	
	switch s.Param.Operator {
	case "IN":
		fallthrough
	case "NIN":
		l := make([]string, 0)
		for _, v := range s.Param.Values {
			l = append(l, "?")
			f.Values = append(f.Values, v)
		}
		return fmt.Sprintf("%s.%s %s (%s)", entity, field, operator, strings.Join(l, ", "))
	case "PFX":
		f.Values = append(f.Values, fmt.Sprintf("%s%%", s.Param.Value))
		return fmt.Sprintf("%s.%s %s ?", entity, field, operator)
	case "SFX":
		f.Values = append(f.Values, fmt.Sprintf("%%%s", s.Param.Value))
		return fmt.Sprintf("%s.%s %s ?", entity, field, operator)
	default:	
		f.Values = append(f.Values, s.Param.Value)
		return fmt.Sprintf("%s.%s %s ?", entity, field, operator)
	}
}

func (t Translater) Translate(f *Filter) string {
	c := make([]string, 0)
	f.Values = make([]interface{}, 0)

	for _, v := range f.Params {
		switch e := v.(type) {
		case Connective:
			r := t.ProcessConnective(f, e)
			if r != "" {
				c = append(c, r)
			}
		case Selectable:
			r := t.ProcessSelectable(f, e)
			if r != "" {
				c = append(c, r)
			}
		}
	}

	return strings.Join(c, " AND ")
}

type Cond map[string]map[string]interface{}

type Connective struct {
	Operator string
	Operands []interface{}
}

func (c Connective) Values() []interface{} {
	return Values(c.Operands)
}

type ConnectFunc func(ops ...interface{}) Connective

func And(ops ...interface{}) Connective {
	return Connective{"AND", ops}
}

func Or(ops ...interface{}) Connective {
	return Connective{"OR", ops}
}

type Param struct {
	Operator string
	Value interface{}
	Values []interface{}
}

type Selectable struct {
	Entity string
	Field string
	Param Param
}

func (s Selectable) Values() []interface{} {
	values := make([]interface{}, 0)
	if len(s.Param.Values) > 0 {
		for _, v := range s.Param.Values {
			values = append(values, v)
		}
	} else {
		values = append(values, s.Param.Value)
	}
	return values
}

func (e Selectable) Eq(value interface{}) Selectable {
	e.Param = Param{Operator: "EQ", Value: value}
	return e
}

func (e Selectable) Ne(value interface{}) Selectable {
	e.Param = Param{Operator: "NE", Value: value}
	return e
}

func (e Selectable) Gt(value interface{}) Selectable {
	e.Param = Param{Operator: "GT", Value: value}
	return e
}

func (e Selectable) Gte(value interface{}) Selectable {
	e.Param = Param{Operator: "GTE", Value: value}
	return e
}

func (e Selectable) Pfx(value interface{}) Selectable {
	e.Param = Param{Operator: "PFX", Value: value}
	return e
}

func (e Selectable) Sfx(value interface{}) Selectable {
	e.Param = Param{Operator: "SFX", Value: value}
	return e
}

func (e Selectable) Lt(value interface{}) Selectable {
	e.Param = Param{Operator: "LT", Value: value}
	return e
}

func (e Selectable) Lte(value interface{}) Selectable {
	e.Param = Param{Operator: "LTE", Value: value}
	return e
}

func (e Selectable) In(values []interface{}) Selectable {
	e.Param = Param{Operator: "IN", Values: values}
	return e
}

func (e Selectable) Nin(values []interface{}) Selectable {
	e.Param = Param{Operator: "NIN", Values: values}
	return e
}
