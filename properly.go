package properly

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/scanner"
)

var (
	// ErrInvalidIndex returned if an invalid index is used for an array.
	ErrInvalidIndex = errors.New("Invalid index")
	// ErrNotSupported returned if an unsupported value is encountered while
	// evaluating the expression.
	ErrNotSupported = errors.New("Value not supported")
	// ErrNotFound returned if a property was not found.
	ErrNotFound = errors.New("Property not found")
)

// Value gets the value of the object graph at the specified path.
func Value(object interface{}, expr string) (v interface{}, ok bool, err error) {
	v = object
	ok = true
	if object == nil {
		return
	}

	var keys []string
	keys, err = split(expr)
	if err != nil {
		ok = false
		return
	}

	for _, key := range keys {
		v, ok, err = value(v, key)
		if err != nil {
			ok = false
			break
		}
	}
	return
}

func value(value interface{}, key string) (interface{}, bool, error) {
	if key == "" {
		return value, true, nil
	}
	ov := reflect.ValueOf(value)

	kind := ov.Kind()
	if kind == reflect.Ptr {
		ov = ov.Elem()
		kind = ov.Kind()
	} else if kind == reflect.Interface {
		kind = ov.Elem().Kind()
	}

	var nv reflect.Value
	ok := true
	switch kind {
	case reflect.Map:
		nv = ov.MapIndex(reflect.ValueOf(key))
		if !nv.IsValid() {
			nv = reflect.Zero(ov.Type().Elem())
			ok = false
		}

	case reflect.Struct:
		nv = ov.FieldByName(key)
		ok = nv.IsValid()

	case reflect.Slice, reflect.Array, reflect.String:
		i, err := strconv.Atoi(key)
		if err != nil {
			return value, false, err
		}
		nv = ov.Index(i)

	case reflect.Invalid:
		return value, false, ErrNotFound

	default:
		return value, false, ErrNotSupported
	}

	if nv.IsValid() {
		value = nv.Interface()
	} else {
		value = nil
	}

	return value, ok, nil
}

func split(expr string) (keys []string, err error) {

	var msgs []string
	var s scanner.Scanner
	s.Init(strings.NewReader(expr))
	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanStrings
	s.Error = func(s *scanner.Scanner, msg string) { msgs = append(msgs, fmt.Sprintf("%s %s", s.Pos(), msg)) }

	key := ""
	keys = []string{}
	for err == nil {
		t := s.Peek()
		// fmt.Printf(">>> %s: %s %s\n", s.Pos(), scanner.TokenString(t), s.TokenText())
		switch t {
		case '[':
			key, err = scanBracketedKey(&s)
		case '.':
			s.Scan()
			continue
		case scanner.EOF:
			goto end
		default:
			key, err = scanKey(&s)
		}
		if len(msgs) > 0 {
			err = errors.New(strings.Join(msgs, "\n"))
		}
		if err == nil {
			keys = append(keys, key)
		}
	}
end:
	return
}

func scanKey(s *scanner.Scanner) (key string, err error) {
	t := s.Scan()
	switch t {
	case scanner.Ident, scanner.Int, scanner.Float:
		key = s.TokenText()
	case scanner.String:
		key = strings.Trim(s.TokenText(), "\"")
	default:
		err = fmt.Errorf("Unexpected token at %s. Expected ident, number or string, had %s", s.Pos(), scanner.TokenString(t))
	}
	return
}

func scanBracketedKey(s *scanner.Scanner) (key string, err error) {
	s.Scan() // scan the '['
	key, err = scanKey(s)
	if err == nil {
		t := s.Scan()
		if t != ']' {
			err = fmt.Errorf("Unexpected token at %s. Expected ']', had %s", s.Pos(), scanner.TokenString(t))
		}
	}
	return
}
