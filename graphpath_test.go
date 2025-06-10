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

func TestBasicPathScanNil(t *testing.T) {
	var p BasicPath
	err := p.Scan(nil)
	if err != nil {
		t.Error(err)
	} else if p.Valid {
		t.Errorf("got %s, want NULL", p)
	}
}

func TestBasicPathScanType(t *testing.T) {
	src := 0
	var p BasicPath
	err := p.Scan(src)
	if err == nil {
		t.Errorf("error expected for %T", src)
	}
}

func TestBasicPathScanZero(t *testing.T) {
	var src interface{} = []byte(nil)
	var p BasicPath
	err := p.Scan(src)
	if err == nil {
		t.Errorf("error expected for %v", src)
	}
}

func TestBasicPathScan(t *testing.T) {
	tests := []struct {
		b  []byte
		nv int
		ne int
	}{
		{[]byte("[]"), 0, 0},
		{[]byte(`[v[3.1]{},e[4.1][3.1,3.2]{},v[3.2]{},NULL,NULL]`), 3, 2},
	}
	for _, c := range tests {
		var p BasicPath
		err := p.Scan(c.b)
		if err != nil {
			t.Error(err)
		} else if !p.Valid {
			t.Errorf("got NULL, want Valid %T", p)
		} else {
			if nv := len(p.Vertices); nv != c.nv {
				t.Errorf("got len(p.Vertices) == %d, want %d", nv, c.nv)
			} else if ne := len(p.Edges); ne != c.ne {
				t.Errorf("got len(p.Edges) == %d, want %d", ne, c.ne)
			}
		}
	}
}

func TestServerGraphpath(t *testing.T) {
	skipUnlessServerTest(t)

	db := mustOpenAndSetGraph(t)
	defer db.Close()

	_, err := db.Exec(`CREATE (:pv)-[:pe]->(:pv)-[:pe]->(:pv)`)
	if err != nil {
		t.Fatal(err)
	}

	var p BasicPath
	q := `MATCH p=(:pv)-[:pe]->(:pv)-[:pe]->(:pv) RETURN p LIMIT 1`
	err = db.QueryRow(q).Scan(&p)
	if err != nil {
		t.Error(err)
	} else if !p.Valid {
		t.Errorf("got NULL, want Valid %T", p)
	} else {
		if nv := len(p.Vertices); nv != 3 {
			t.Errorf("got len(p.Vertices) == %d, want %d", nv, 3)
		} else if ne := len(p.Edges); ne != 2 {
			t.Errorf("got len(p.Edges) == %d, want %d", ne, 2)
		}
	}
}
