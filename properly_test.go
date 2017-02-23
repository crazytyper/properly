package properly

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type embeddedType struct {
	Int      int
	IntArray []int
	Map      map[string]interface{}
	BoolMap  map[string]bool
}

type sampleType struct {
	embeddedType
	String string
	Nested *sampleType
}

var sampleData = &sampleType{
	String: "Hello world!",
	embeddedType: embeddedType{
		Int:      42,
		IntArray: []int{1, 2, 4, 8, 16},
		Map: map[string]interface{}{
			"shoeSize": 42.5,
			"42.5":     "shoeSize",
		},
		BoolMap: map[string]bool{
			"truthy": true,
		},
	},
}

func TestValue(t *testing.T) {
	is := assert.New(t)

	v := func(v interface{}, ok bool, err error) interface{} {
		is.True(ok)
		is.NoError(err)
		return v
	}
	n := func(v interface{}, ok bool, err error) interface{} {
		is.False(ok)
		is.NoError(err)
		return v
	}
	e := func(v interface{}, ok bool, err error) error {
		is.False(ok)
		return err
	}

	var value interface{}
	is.Equal(42, v(Value(42, "")))
	is.Equal(42, v(Value(42, ".")))

	value = map[string]int{"a": 42}
	is.Equal(value, v(Value(value, ".")))
	is.Equal(value, v(Value(value, "..")))
	is.Equal(value, v(Value(value, "...")))
	is.Equal(value, v(Value(value, "")))
	is.Equal(value, v(Value(value, "\"\"")))
	is.Equal(42, v(Value(value, "a")))
	is.Equal(42, v(Value(value, ".a")))

	value = []int{1, 2, 3}
	is.Equal(1, v(Value(value, "0")))
	is.Equal(2, v(Value(value, "1")))
	is.Equal(3, v(Value(value, "2")))
	is.Equal(1, v(Value(value, "[0]")))
	is.Equal(2, v(Value(value, "[1]")))
	is.Equal(3, v(Value(value, "[2]")))
	is.IsType(&strconv.NumError{}, e(Value(value, "a")))

	is.Equal("Hello world!", v(Value(sampleData, "String")))
	is.Equal("Hello"[2], v(Value(sampleData, "String[2]")))
	is.Equal(42, v(Value(sampleData, "Int")))
	is.Equal(sampleData.IntArray, v(Value(sampleData, "IntArray")))
	is.Equal(4, v(Value(sampleData, "IntArray[\"2\"]")))
	is.Equal(4, v(Value(sampleData, "IntArray[2]")))

	is.Equal(42.5, v(Value(sampleData, "Map.shoeSize")))
	is.Equal(nil, n(Value(sampleData, "Map.age")))
	is.Equal(true, v(Value(sampleData, "BoolMap.truthy")))
	is.Equal(false, n(Value(sampleData, "BoolMap.falsy")))

	is.Equal("shoeSize", v(Value(sampleData, "Map[\"42.5\"]")))

	is.Equal(nil, v(Value(nil, "whatever")))

	is.Equal(ErrNotFound, e(Value(sampleData, "Map.age.test.x.y.z")))
	is.Equal(ErrNotSupported, e(Value(new(chan bool), "whatever")))

	is.Equal("Unexpected token at Map[a.b]:1:7. Expected ']', had \".\"", e(Value(sampleData, "Map[a.b]")).Error())
	is.Equal("Unexpected token at -:1:2. Expected ident, number or string, had \"-\"", e(Value(sampleData, "-")).Error())
	is.Equal("\"Test:1:6 literal not terminated", e(Value(sampleData, "\"Test")).Error())
}

func TestString(t *testing.T) {
	is := assert.New(t)
	ok := func(v string, ok bool, err error) string {
		is.True(ok)
		is.NoError(err)
		return v
	}
	nok := func(v string, ok bool, err error) string {
		is.False(ok)
		is.NoError(err)
		return v
	}

	is.Equal("Hello world!", ok(String(sampleData, "String")))
	is.Equal("shoeSize", ok(String(sampleData, "Map[\"42.5\"]")))
	is.Equal("", nok(String(sampleData, "Int")))
}
