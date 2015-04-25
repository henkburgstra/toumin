package toumin

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

type Patient struct {
	*Model
	Achternaam string
}

func NewPatient(name string) IModel {
	p := new(Patient)
	p.Model = NewModel(name).(*Model)
	p.Model.SetOwner(p)
	return p
}

func (p *Patient) EchteNaam() string {
	return "Leeuwerik"
}

type BehandeldagVerrichtingen struct {
	*Model
	Id                        int
	LokaleVerrichtingcodeCode string
	Teller                    int
	BehandeldagKey            string
}

func NewBehandeldagVerrichtingen(name string) IModel {
	p := new(BehandeldagVerrichtingen)
	p.Model = NewModel(name).(*Model)
	p.Model.SetOwner(p)
	return p
}

func makeRegistry(engine *Engine) *Registry {
	registry := NewRegistry(engine)
	registry.SetTableSuffix("_data")
	registry.SetFieldPrefix("{model}_")
	registry.LoadEntities()
	return registry
}

func TestQuery(t *testing.T) {
	engine := makeEngine()
	db, err := engine.Connect()
	if err != nil {
		t.Fatalf("TestQuery(): engine.Connect(): %s", err.Error())
	}
	defer db.Close()
	registry := makeRegistry(engine)
	registry.RegisterModel("patient", NewPatient)

	patient := registry.Query("patient").Get("PJJG-AA0010").(*Patient)
	if patient == nil {
		t.Errorf("Geen patient gevonden")
	} else {
		fmt.Println(patient.Field("achternaam").String())
		fmt.Println(patient.Achternaam)
		fmt.Println(patient.EchteNaam())
	}
}

func TestFromSql(t *testing.T) {
	engine := makeEngine()
	db, err := engine.Connect()
	if err != nil {
		t.Fatalf("TestFromSql(): engine.Connect(): %s", err.Error())
	}
	defer db.Close()
	registry := makeRegistry(engine)
	registry.RegisterModel("behandeldag_verrichtingen", NewBehandeldagVerrichtingen)

	for _, entry := range registry.Query("behandeldag_verrichtingen").FromSql(`
		SELECT *
		FROM behandeldag_verrichtingen
		WHERE behandeldag_verrichtingen_id <= ?`, 2).All() {
		verrichting := entry.(*BehandeldagVerrichtingen)
		if verrichting.Field("lokale_verrichtingcode_code").String() != verrichting.LokaleVerrichtingcodeCode {
			t.Errorf(`Field("lokale_verrichtingcode_code").String() != LokaleVerrichtingcodeCode`)
		}
		verrichting.LokaleVerrichtingcodeCode = "A777"
		if verrichting.Field("lokale_verrichtingcode_code").String() != verrichting.LokaleVerrichtingcodeCode {
			t.Errorf(`(change LokaleVerrichtingcodeCode) Field("lokale_verrichtingcode_code").String() != LokaleVerrichtingcodeCode`)
		}
		verrichting.Field("lokale_verrichtingcode_code").Set("A888")
		if verrichting.Field("lokale_verrichtingcode_code").String() != verrichting.LokaleVerrichtingcodeCode {
			t.Errorf(`(change lokale_verrichtingcode_code) Field("lokale_verrichtingcode_code").String() != LokaleVerrichtingcodeCode`)
		}
		if verrichting.Field("teller").Int() != (int64)(verrichting.Teller) {
			t.Errorf(`verrichting.Field("teller").Int() != verrichting.Teller`)
		}
		verrichting.Field("teller").Set(2)
		if verrichting.Field("teller").Int() != (int64)(verrichting.Teller) {
			t.Errorf(`(change teller) Field("teller").Int() != verrichting.Teller`)
		}
		verrichting.Teller = 3
		if verrichting.Field("teller").Int() != (int64)(verrichting.Teller) {
			t.Errorf(`(change teller) Field("teller").Int() != verrichting.Teller`)
		}
	}
}
