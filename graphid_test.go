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
	"bytes"
	"testing"
)

func TestNewGraphIdError(t *testing.T) {
	tests := []string{
		"",
		"0.1",
		"1.0",
		"65536.281474976710655",
		"65535.281474976710656",
	}
	for _, s := range tests {
		_, err := NewGraphId(s)
		if err == nil {
			t.Errorf("error expected for %q", s)
		}
	}
}

func mustNewGraphId(str string) GraphId {
	gid, err := NewGraphId(str)
	if err != nil {
		panic(err)
	}
	return gid
}

func TestGraphIdEqual(t *testing.T) {
	tests := []struct {
		x     GraphId
		y     GraphId
		equal bool
	}{
		{mustNewGraphId("NULL"), mustNewGraphId("1.1"), false},
		{mustNewGraphId("1.1"), mustNewGraphId("NULL"), false},
		{mustNewGraphId("1.1"), mustNewGraphId("1.1"), true},
		{mustNewGraphId("1.1"), mustNewGraphId("65535.281474976710655"), false},
	}
	for _, c := range tests {
		equal := c.x.Equal(c.y)
		if equal != c.equal {
			t.Errorf("got %q.Equal(%q) == %t, want %t", c.x, c.y, equal, c.equal)
		}
	}
}

func TestGraphIdScanNil(t *testing.T) {
	var gid GraphId
	err := gid.Scan(nil)
	if err != nil {
		t.Error(err)
	} else if gid.Valid {
		t.Errorf("got %q, want NULL", gid)
	}
}

func TestGraphIdScanType(t *testing.T) {
	src := 0
	var gid GraphId
	err := gid.Scan(src)
	if err == nil {
		t.Errorf("error expected for %T", src)
	}
}

func TestGraphIdScanZero(t *testing.T) {
	var src interface{} = []byte(nil)
	var gid GraphId
	err := gid.Scan(src)
	if err == nil {
		t.Errorf("error expected for %v", src)
	}
}

func TestGraphIdScanDuplicate(t *testing.T) {
	src := []byte("1.1")

	var gid GraphId
	_ = gid.Scan(src)

	if &gid.b[0] == &src[0] {
		t.Error("GraphId references underlying array")
	}
}

func TestGraphIdArrayScanType(t *testing.T) {
	src := 0
	var gids []GraphId
	a := Array(&gids)
	err := a.Scan(src)
	if err == nil {
		t.Errorf("error expected for %T", src)
	}
}

var graphIdArrayTests = []struct {
	val  interface{}
	gids []GraphId
}{
	{
		[]byte("{NULL,1.1,65535.281474976710655}"),
		[]GraphId{
			mustNewGraphId("NULL"),
			mustNewGraphId("1.1"),
			mustNewGraphId("65535.281474976710655"),
		},
	},
	{
		[]byte("{}"),
		[]GraphId{},
	},
	{
		nil,
		nil,
	},
}

func TestGraphIdArrayScan(t *testing.T) {
	for _, c := range graphIdArrayTests {
		var gids []GraphId
		a := Array(&gids)
		err := a.Scan(c.val)
		if err != nil {
			t.Error(err)
			continue
		}

		if n, cn := len(gids), len(c.gids); n != cn {
			t.Errorf("got len(gids) == %d, want %d", n, cn)
			continue
		}

		for i, gid := range gids {
			if !gid.Valid && !c.gids[i].Valid {
				continue
			}
			if !gid.Equal(c.gids[i]) {
				t.Errorf("got %q, want %q", gid, c.gids[i])
				break
			}
		}
	}
}

func TestGraphIdArrayValue(t *testing.T) {
	for _, c := range graphIdArrayTests {
		val, err := Array(c.gids).Value()
		if err != nil {
			t.Error(err)
			continue
		}

		if val == nil {
			if c.val != nil {
				t.Errorf("error expected for %v", c.gids)
			}
		} else {
			b, ok := val.([]byte)
			if !ok {
				t.Errorf("got %T, want []byte", val)
				continue
			}
			if cb := c.val.([]byte); !bytes.Equal(b, cb) {
				t.Errorf("got %q, want %q", b, cb)
			}
		}
	}
}

func TestServerGraphId(t *testing.T) {
	skipUnlessServerTest(t)

	db := mustOpenAndSetGraph(t)
	defer db.Close()

	_, err := db.Exec(`CREATE (:gid)`)
	if err != nil {
		t.Fatal(err)
	}

	var gid GraphId
	err = db.QueryRow(`MATCH (n:gid) RETURN id(n)`).Scan(&gid)
	if err != nil {
		t.Error(err)
	} else if !gid.Valid {
		t.Errorf("got NULL, want Valid %T", gid)
	}

	var cnt int64
	q := `MATCH (n:gid) WHERE id(n) = $1 RETURN count(*)`
	err = db.QueryRow(q, gid).Scan(&cnt)
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Errorf("got %d, want %d", cnt, 1)
	}

	gids := []GraphId{mustNewGraphId("NULL"), mustNewGraphId("1.1"), mustNewGraphId("65535.281474976710655")}
	var gidsOut []GraphId
	err = db.QueryRow(`SELECT $1::_graphid`, Array(gids)).Scan(Array(&gidsOut))
	if err == nil {
		for i, gid := range gidsOut {
			if !gid.Valid && !gids[i].Valid {
				continue
			}
			if !gid.Equal(gids[i]) {
				t.Errorf("got %q, want %q", gid, gids[i])
				break
			}
		}
	} else {
		t.Error(err)
	}

	err = db.QueryRow(`SELECT NULL::_graphid`).Scan(Array(&gidsOut))
	if err != nil {
		t.Error(err)
	} else if gidsOut != nil {
		t.Errorf("got %v, want nil", gidsOut)
	}
}
