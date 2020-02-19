package adapt_test

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/iliyanmotovski/adapt"

	"github.com/stretchr/testify/assert"
)

var adapter = adapt.NewAdapter()

func init() {
	adapter.RegisterAdaptFunc("boolToNullBool", boolToNullBoolAdapter)
	adapter.RegisterAdaptFunc("sqlNullBoolToBool", sqlNullBoolToBoolAdapter)
	adapter.RegisterAdaptFunc("nullBoolToSQLNullBool", nullBoolToSQLNullBoolAdapter)
}

func TestAdapter(t *testing.T) {
	src := Src{
		Valid:     "yes",
		Locale:    "en_US",
		IsDone:    sql.NullBool{Bool: true, Valid: true},
		IsBroken:  NullBool{Bool: false, Valid: true},
		IsWorking: true,
		Skip:      true,
	}

	dst := Dst{}

	err := adapter.SrcToDst(src, &dst)
	assert.Nil(t, err)

	want := Dst{
		IsDone:     true,
		Locale:     "en_US",
		IsValid:    "yes",
		IsWorking:  NullBool{Bool: true, Valid: true},
		IsItBroken: sql.NullBool{Bool: false, Valid: true},
	}

	assert.Equal(t, want, dst)
}

type Src struct {
	Locale    string
	Valid     string       `dstName:"IsValid"`
	IsWorking bool         `adapter:"boolToNullBool"`
	IsDone    sql.NullBool `adapter:"sqlNullBoolToBool"`
	IsBroken  NullBool     `dstName:"IsItBroken" adapter:"nullBoolToSQLNullBool"`
	Skip      bool         `adapter:"skip"`
}

type Dst struct {
	IsWorking  NullBool
	Locale     string
	IsValid    string
	IsItBroken sql.NullBool
	IsDone     bool
}

type NullBool struct {
	Bool  bool
	Valid bool
}

func boolToNullBoolAdapter(src reflect.Value) interface{} {
	b := src.Interface().(bool)
	return NullBool{Bool: b, Valid: true}
}

func sqlNullBoolToBoolAdapter(src reflect.Value) interface{} {
	b := src.Interface().(sql.NullBool)
	return b.Bool && b.Valid
}

func nullBoolToSQLNullBoolAdapter(src reflect.Value) interface{} {
	b := src.Interface().(NullBool)
	return sql.NullBool{Bool: b.Bool, Valid: b.Valid}
}
