/*
Copyright 2025 SKAI Worldwide Co., Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ag

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
)

// Array returns database/sql Scanner and database/sql/driver Valuer for dest
// that is a slice or an array of the following types; GraphId, and entities
// for vertex and edge.
//
// If the type of dest is not *[]GraphId and []GraphId, Value of Array returns
// an error since passing entities as parameters is not allowed.
func Array(dest interface{}) interface {
	sql.Scanner
	driver.Valuer
} {
	switch dest := dest.(type) {
	case *[]GraphId:
		return (*graphIdArray)(dest)
	case []GraphId:
		return (*graphIdArray)(&dest)

	case *[]BasicVertex:
		return (*basicVertexArray)(dest)

	case *[]BasicEdge:
		return (*basicEdgeArray)(dest)
	}

	return elementArray{dest}
}

// NullArrayError is returned by Scan if the type of dest for Array(dest) is
// array and the value from the database driver is NULL.
type NullArrayError struct{}

func (_ NullArrayError) Error() string {
	return "NULL"
}

var nullElementValue = []byte("NULL")

type elementsReader interface {
	readElements(b []byte) ([]interface{}, error)
}

type elementArray struct {
	dest interface{}
}

var typeArrayScanner = reflect.TypeOf((*elementsReader)(nil)).Elem()
var typeSQLScanner = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

func (a elementArray) Scan(src interface{}) error {
	// *[]t
	rv := reflect.ValueOf(a.dest)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not a pointer to slice or array", a.dest)
	}
	if rv.IsNil() {
		return fmt.Errorf("%T is nil", a.dest)
	}

	// []t
	rv = rv.Elem()
	rk := rv.Kind()
	if rk != reflect.Slice && rk != reflect.Array {
		return fmt.Errorf("%T is not a pointer to slice or array", a.dest)
	}
	rt := rv.Type()

	// t
	rte := rt.Elem()
	// t.(elementsReader)
	if !rte.Implements(typeArrayScanner) {
		return fmt.Errorf("%s does not implement %s", rte, typeArrayScanner)
	}
	// t.(sql.Scanner)
	if !reflect.PtrTo(rte).Implements(typeSQLScanner) {
		return fmt.Errorf("%s does not implement %s", rte, typeSQLScanner)
	}

	if src == nil {
		switch rk {
		case reflect.Slice:
			// *a.dest = nil
			rv.Set(reflect.Zero(rt))
			return nil
		case reflect.Array:
			return NullArrayError{}
		default:
			panic("cannot happen")
		}
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid source for %s: %T", rt, src)
	}
	if len(b) < 1 {
		return fmt.Errorf("invalid source for %s: %v", rt, b)
	}

	reader := reflect.Zero(rte).Interface().(elementsReader)
	ds, err := reader.readElements(b)
	if err != nil {
		return errors.New("failed to read elements: " + err.Error())
	}
	n := len(ds)

	switch rk {
	case reflect.Slice:
		// *a.dest = make([]t, n)
		rv.Set(reflect.MakeSlice(reflect.SliceOf(rte), n, n))
	case reflect.Array:
		// len(*a.dest)
		if rt.Len() != n {
			return fmt.Errorf("number of elements is %d but %s", n, rt)
		}
	default:
		panic("cannot happen")
	}

	for i := 0; i < n; i++ {
		// a.dest[i].(sql.Scanner).Scan(ds[i])
		e := rv.Index(i).Addr().Interface().(sql.Scanner)
		err := e.Scan(ds[i])
		if err != nil {
			return errors.New("invalid element: " + err.Error())
		}
	}

	return nil
}

func (a elementArray) Value() (driver.Value, error) {
	return nil, fmt.Errorf("Value() on an array of %T is not supported", a.dest)
}
