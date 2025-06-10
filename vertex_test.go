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

// ScanEntity - case nil
func TestBasicVertexScanNil(t *testing.T) {
	var v BasicVertex
	err := v.Scan(nil)
	if err != nil {
		t.Error(err)
	} else if v.Valid {
		t.Errorf("got %s, want NULL", v)
	}
}

// ScanEntity - default
func TestBasicVertexScanType(t *testing.T) {
	src := 0
	var v BasicVertex
	err := v.Scan(src)
	if err == nil {
		t.Errorf("error expected for %T", src)
	}
}

// (Vertex).readEntity, makeVertexData
func TestBasicVertexScanError(t *testing.T) {
	tests := [][]byte{
		[]byte(nil),
		[]byte(""),
		[]byte("v"),
		[]byte("v[3.1]"),
		[]byte("v[0.0]{}"),
	}
	for _, b := range tests {
		var v BasicVertex
		err := v.Scan(b)
		if err == nil {
			t.Errorf("error expected for %s", b)
		}
	}
}

// (Vertex).readEntity, saveEntityData, (*VertexHeader).SaveEntity,
// (*BasicVertex).SaveProperties
func TestBasicVertexScan(t *testing.T) {
	b := []byte(`v[3.1]{"s": "", "n": 0, "b": false, "a": [], "o": {}}`)
	var v BasicVertex
	err := v.Scan(b)
	if err != nil {
		t.Error(err)
	} else if !v.Valid {
		t.Errorf("got NULL, want Valid %T", v)
	}
}

type userVertex struct {
	VertexHeader `json:"-"`
	Name         string
}

func (v *userVertex) Scan(src interface{}) error {
	return ScanEntity(src, v)
}

// saveEntityData - json.Unmarshal
func TestPropertiesDefault(t *testing.T) {
	b := []byte(`v[3.1]{"name": "go"}`)
	var v userVertex
	err := v.Scan(b)
	if err != nil {
		t.Error(err)
	} else if !v.Valid {
		t.Errorf("got NULL, want Valid %T", v)
	} else if v.Name != "go" {
		t.Errorf(`got %q, want "go"`, v.Name)
	}
}

func TestBasicVertexArrayScanNil(t *testing.T) {
	var vs []BasicVertex
	err := Array(&vs).Scan(nil)
	if err != nil {
		t.Error(err)
	} else if vs != nil {
		t.Errorf("got %v, want nil", vs)
	}
}

func TestBasicVertexArrayScanType(t *testing.T) {
	src := 0
	var vs []BasicVertex
	err := Array(&vs).Scan(src)
	if err == nil {
		t.Errorf("error expected for %T", src)
	}
}

func TestBasicVertexArrayScanZero(t *testing.T) {
	var src interface{} = []byte(nil)
	var vs []BasicVertex
	err := Array(&vs).Scan(src)
	if err == nil {
		t.Errorf("error expected for %v", src)
	}
}

var vertexArrayTests = []struct {
	src interface{}
	n   int
}{
	{
		[]byte(`[NULL,v[3.1]{"name": "go"},v[3.2]{"name": "go"}]`),
		3,
	},
	{
		[]byte("[]"),
		0,
	},
	{
		nil,
		0,
	},
}

// (*basicVertexArray).Scan, ScanEntity - case *entityData
func TestBasicVertexArrayScan(t *testing.T) {
	for _, c := range vertexArrayTests {
		var vs []BasicVertex
		err := Array(&vs).Scan(c.src)
		if err != nil {
			t.Error(err)
			continue
		}

		if n := len(vs); n != c.n {
			t.Errorf("got len(vs) == %d, want %d", n, c.n)
		}
	}
}

func TestBasicVertexArrayValue(t *testing.T) {
	var vs []BasicVertex
	_, err := Array(&vs).Value()
	if err == nil {
		t.Error("error expected for Value() on basicVertexArray")
	}
}

func TestUserVertexArrayScan(t *testing.T) {
	for _, c := range vertexArrayTests {
		var vs []userVertex
		err := Array(&vs).Scan(c.src)
		if err != nil {
			t.Error(err)
			continue
		}

		if n := len(vs); n != c.n {
			t.Errorf("got len(vs) == %d, want %d", n, c.n)
		}
	}
}

func TestServerVertex(t *testing.T) {
	skipUnlessServerTest(t)

	db := mustOpenAndSetGraph(t)
	defer db.Close()

	_, err := db.Exec(`CREATE (:vv)-[:ve]->(:vv)`)
	if err != nil {
		t.Fatal(err)
	}

	var v BasicVertex
	err = db.QueryRow(`MATCH (n:vv) RETURN n LIMIT 1`).Scan(&v)
	if err != nil {
		t.Error(err)
	} else if !v.Valid {
		t.Errorf("got NULL, want Valid %T", v)
	}

	var vs []BasicVertex
	q := `MATCH p=(:vv)-[:ve]->(:vv) RETURN nodes(p) LIMIT 1`
	err = db.QueryRow(q).Scan(Array(&vs))
	if err != nil {
		t.Error(err)
	} else if vs == nil {
		t.Errorf("got nil, want non-NULL %T", vs)
	} else if n := len(vs); n != 2 {
		t.Errorf("got len(vs) == %d, want 2", n)
	}

	err = db.QueryRow(`SELECT ARRAY[NULL]::_vertex`).Scan(Array(&vs))
	if err != nil {
		t.Error(err)
	} else if vs == nil {
		t.Errorf("got nil, want non-NULL %T", vs)
	} else if vs[0].Valid {
		t.Errorf("got %v, want NULL", vs[0])
	}
}
