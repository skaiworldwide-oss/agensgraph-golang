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

import "testing"

type testElement struct{}

func (_ testElement) readElements(b []byte) ([]interface{}, error) {
	return []interface{}{}, nil
}

func TestArrayScanError(t *testing.T) {
	tests := []interface{}{
		nil,
		0,
		(*byte)(nil),
		&testElement{},
		&[]byte{},
		&[]testElement{},
	}
	for _, i := range tests {
		err := Array(i).Scan(nil)
		if err == nil {
			t.Errorf("error expected for %T", i)
		}
	}
}

type testElementEx struct {
	testElement
}

func (_ testElementEx) Scan(src interface{}) error {
	return nil
}

func TestArrayScanNullSlice(t *testing.T) {
	var es []testElementEx
	err := Array(&es).Scan(nil)
	if err != nil {
		t.Fatalf("got %v, want NULL", es)
	}

	if es != nil {
		t.Fatalf("got %v, want nil", es)
	}
}

func TestArrayScanNullArray(t *testing.T) {
	es := [1]testElementEx{}
	err := Array(&es).Scan(nil)
	if err == nil {
		t.Fatalf("got %v, want error", es)
	}

	if _, ok := err.(NullArrayError); !ok {
		t.Errorf("NullArrayError expected")
	}
}

func TestArrayScanType(t *testing.T) {
	src := 0
	var es []testElementEx
	err := Array(&es).Scan(src)
	if err == nil {
		t.Errorf("error expected for %T", src)
	}
}

func TestArrayScanZero(t *testing.T) {
	var src interface{} = []byte(nil)
	var es []testElementEx
	err := Array(&es).Scan(src)
	if err == nil {
		t.Errorf("error expected for %T", src)
	}
}

func TestArrayScanArrayLen(t *testing.T) {
	es := [1]testElementEx{}
	err := Array(&es).Scan([]byte("dummy"))
	if err == nil {
		t.Errorf("error expected for %T", es)
	}
}

func TestArrayValue(t *testing.T) {
	var es []testElementEx
	_, err := Array(&es).Value()
	if err == nil {
		t.Errorf("error expected for Value() on Array")
	}
}
