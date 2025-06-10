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

// (Edge).readEntity, makeEdgeData
func TestBasicEdgeScanError(t *testing.T) {
	tests := [][]byte{
		[]byte(nil),
		[]byte(""),
		[]byte("e"),
		[]byte("e[]"),
		[]byte("e[4.1][3.1,3.2]"),
		[]byte("e[0.0][3.1,3.2]{}"),
		[]byte("e[4.1][0.0,3.2]{}"),
		[]byte("e[4.1][3.1,0.0]{}"),
	}
	for _, b := range tests {
		var e BasicEdge
		err := e.Scan(b)
		if err == nil {
			t.Errorf("error expected for %s", b)
		}
	}
}

// (Edge).readEntity, (*EdgeHeader).SaveEntity, (*BasicEdge).SaveProperties
func TestBasicEdgeScan(t *testing.T) {
	b := []byte(`e[4.1][3.1,3.2]{"s": "", "n": 0, "b": false, "a": [], "o": {}}`)
	var e BasicEdge
	err := e.Scan(b)
	if err != nil {
		t.Error(err)
	} else if !e.Valid {
		t.Errorf("got NULL, want Valid %T", e)
	}
}

func TestBasicEdgeArrayScanNil(t *testing.T) {
	var es []BasicEdge
	err := Array(&es).Scan(nil)
	if err != nil {
		t.Error(err)
	} else if es != nil {
		t.Errorf("got %v, want nil", es)
	}
}

func TestBasicEdgeArrayScanType(t *testing.T) {
	src := 0
	var es []BasicEdge
	err := Array(&es).Scan(src)
	if err == nil {
		t.Errorf("error expected for %T", src)
	}
}

func TestBasicEdgeArrayScanZero(t *testing.T) {
	var src interface{} = []byte(nil)
	var es []BasicEdge
	err := Array(&es).Scan(src)
	if err == nil {
		t.Errorf("error expected for %v", src)
	}
}

// (*basicEdgeArray).Scan
func TestBasicEdgeArrayScan(t *testing.T) {
	tests := []struct {
		src interface{}
		n   int
	}{
		{
			[]byte(`[NULL,e[4.1][3.1,3.2]{"name": "go"},e[4.2][3.3,3.4]{"name": "go"}]`),
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
	for _, c := range tests {
		var es []BasicEdge
		err := Array(&es).Scan(c.src)
		if err != nil {
			t.Error(err)
			continue
		}

		if n := len(es); n != c.n {
			t.Errorf("got len(es) == %d, want %d", n, c.n)
		}
	}
}

func TestBasicEdgeArrayValue(t *testing.T) {
	var es []BasicEdge
	_, err := Array(&es).Value()
	if err == nil {
		t.Error("error expected for Value() on basicEdgeArray")
	}
}

func TestServerEdge(t *testing.T) {
	skipUnlessServerTest(t)

	db := mustOpenAndSetGraph(t)
	defer db.Close()

	_, err := db.Exec(`CREATE (:ev)-[:ee]->(:ev)-[:ee]->(:ev)`)
	if err != nil {
		t.Fatal(err)
	}

	var e BasicEdge
	err = db.QueryRow(`MATCH (:ev)-[r:ee]->(:ev) RETURN r LIMIT 1`).Scan(&e)
	if err != nil {
		t.Error(err)
	} else if !e.Valid {
		t.Errorf("got NULL, want Valid %T", e)
	}

	var es []BasicEdge
	q := `MATCH p=(:ev)-[:ee]->(:ev)-[:ee]->(:ev) RETURN relationships(p) LIMIT 1`
	err = db.QueryRow(q).Scan(Array(&es))
	if err != nil {
		t.Error(err)
	} else if es == nil {
		t.Errorf("got nil, want non-NULL %T", es)
	} else if n := len(es); n != 2 {
		t.Errorf("got len(vs) == %d, want 2", n)
	}

	err = db.QueryRow(`SELECT ARRAY[NULL]::_edge`).Scan(Array(&es))
	if err != nil {
		t.Error(err)
	} else if es == nil {
		t.Errorf("got nil, want non-NULL %T", es)
	} else if es[0].Valid {
		t.Errorf("got %v, want NULL", es[0])
	}
}
