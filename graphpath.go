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
	"errors"
	"fmt"
	"strings"
)

// PathSaver is an interface used by ScanPath.
type PathSaver interface {
	// SavePath assigns a path from the database driver.
	//
	// valid is true if the path is not NULL.
	//
	// ds is a series of connected vertices and edges. Each element of ds
	// can be stored in an entity for vertex or edge by calling ScanEntity.
	// If valid is false, ds will be nil.
	//
	// An error should be returned if the path cannot be stored without
	// loss of information.
	SavePath(valid bool, ds []interface{}) error
}

// ScanPath reads a path from src and stores the result by calling SavePath.
//
// An error will be returned if the type of src is not []byte, or src is
// invalid.
func ScanPath(src interface{}, saver PathSaver) error {
	if src == nil {
		return saver.SavePath(false, nil)
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid source for graphpath: %T", src)
	}

	n := len(b)
	if n < 1 {
		return fmt.Errorf("invalid source for graphpath: %v", b)
	}

	advance, ds, err := readPath(b)
	if err != nil {
		return err
	}
	if advance != n {
		return fmt.Errorf("bad graphpath representation: %s", b)
	}

	return saver.SavePath(true, ds)
}

func readPath(b []byte) (advance int, ds []interface{}, err error) {
	if bytes.HasPrefix(b, nullElementValue) {
		advance = len(nullElementValue)
		return
	}

	if b[0] != byte('[') {
		err = fmt.Errorf("bad graphpath representation: %s", b)
		return
	}
	advance = 1

	read, readNext := readVertexElement, readEdgeElement
	for b[advance] != byte(']') {
		if len(ds) > 0 {
			// remove comma
			advance++
		}

		n, d, r := read(b[advance:])
		if err != nil {
			err = errors.New("invalid path element: " + r.Error())
			return
		}

		advance += n
		if d == nil {
			ds = append(ds, nil)
		} else {
			ds = append(ds, d)
		}

		read, readNext = readNext, read
	}
	advance++

	return
}

// BasicPath can be used to scan the value from the database driver as a path.
//
// This is a reference implementation that uses PathSaver and ScanPath.
type BasicPath struct {
	Valid    bool
	Vertices []BasicVertex
	Edges    []BasicEdge
}

func (p BasicPath) String() string {
	if !p.Valid {
		return "NULL"
	}

	nv := len(p.Vertices)
	if nv < 1 {
		return "[]"
	}

	ne := len(p.Edges)

	s := make([]string, 0, nv+ne)
	for i := 0; i < ne; i++ {
		s = append(s, p.Vertices[i].String(), p.Edges[i].String())
	}
	s = append(s, p.Vertices[nv-1].String())

	return fmt.Sprintf("[%s]", strings.Join(s, ","))
}

// SavePath implements PathSaver interface.
func (p *BasicPath) SavePath(valid bool, ds []interface{}) error {
	p.Valid = valid
	if !valid {
		return nil
	}

	n := len(ds)
	if n < 1 {
		return nil
	}

	ne := n / 2
	p.Vertices = make([]BasicVertex, ne+1)
	if ne > 0 {
		p.Edges = make([]BasicEdge, ne)
	}

	i, j := 0, 0
	for i < ne {
		err := p.Vertices[i].Scan(ds[j])
		if err != nil {
			return err
		}
		err = p.Edges[i].Scan(ds[j+1])
		if err != nil {
			return err
		}

		i, j = i+1, j+2
	}
	p.Vertices[i].Scan(ds[j])

	return nil
}

// Scan implements the database/sql Scanner interface. It calls ScanPath.
func (p *BasicPath) Scan(src interface{}) error {
	return ScanPath(src, p)
}
