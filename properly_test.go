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

	s := &sampleType{
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
	is.IsType(&strconv.NumError{}, e(Value(value, "a")))

	is.Equal("Hello world!", v(Value(s, "String")))
	is.Equal("Hello"[2], v(Value(s, "String[2]")))
	is.Equal(42, v(Value(s, "Int")))
	is.Equal(s.IntArray, v(Value(s, "IntArray")))
	is.Equal(4, v(Value(s, "IntArray[\"2\"]")))
	is.Equal(4, v(Value(s, "IntArray[2]")))
	is.Equal(4, v(Value(s, "embeddedType.IntArray[2]")))
	is.Equal(4, v(Value(s, "embeddedType[IntArray][2]")))

	is.Equal(42.5, v(Value(s, "Map.shoeSize")))
	is.Equal(nil, n(Value(s, "Map.age")))
	is.Equal(true, v(Value(s, "BoolMap.truthy")))
	is.Equal(false, n(Value(s, "BoolMap.falsy")))

	is.Equal("shoeSize", v(Value(s, "Map[\"42.5\"]")))

	is.Equal(nil, v(Value(nil, "whatever")))

	is.Equal(ErrNotFound, e(Value(s, "Map.age.test.x.y.z")))
	is.Equal(ErrNotSupported, e(Value(new(chan bool), "whatever")))

	is.Equal("Unexpected token at 1:7. Expected ']', had \".\"", e(Value(s, "Map[a.b]")).Error())
	is.Equal("Unexpected token at 1:2. Expected ident, number or string, had \"-\"", e(Value(s, "-")).Error())
	is.Equal("1:6 literal not terminated", e(Value(s, "\"Test")).Error())
}
