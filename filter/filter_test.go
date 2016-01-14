package filter

import (
	"strconv"
	"testing"
	"fmt"
)

func selecteerJaar(jaar int) Selectable {
	return Selectable{Entity: "onderzoek", Field: "datum"}.Pfx(strconv.Itoa(jaar))
}

func selecteerCentrum(centrum string) Selectable {
	return Selectable{Entity: "onderzoek", Field: "centrum"}.Eq(centrum)
}

func selecteerGeslacht(geslacht string) Selectable {
	return Selectable{Entity: "patient", Field: "geslacht"}.Eq(geslacht)
}

func TestFilter(t *testing.T) {
	filter := NewFilter(
		And(
			selecteerJaar(2015),
			selecteerCentrum("ACH"),
			selecteerGeslacht("M")))
	fmt.Println(filter)
	for _, v := range filter.Values() {
		fmt.Printf("%s\n", v)
	}
}

