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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// Vertex gives any struct an ability to read the value from the database
// driver as a vertex if the struct has Vertex as its embedded field.
type Vertex struct{}

// VertexCore represents essential data to identify a vertex.
type VertexCore struct {
	Label string
	Id    GraphId
}

var vertexCoreRegexp = regexp.MustCompile(`^(.+?)\[(\d+\.\d+)\]`)

func (_ Vertex) readEntity(b []byte) (*entityData, error) {
	m := vertexCoreRegexp.FindSubmatch(b)
	if m == nil {
		return nil, fmt.Errorf("bad vertex representation: %s", b)
	}

	return makeVertexData(m[1], m[2], b[len(m[0]):])
}

func makeVertexData(label, id, props []byte) (*entityData, error) {
	var c VertexCore

	c.Label = string(label)

	err := c.Id.Scan(id)
	if err != nil {
		return nil, errors.New("invalid vertex ID: " + err.Error())
	}

	return &entityData{c, props}, nil
}

func (_ Vertex) readElements(b []byte) ([]interface{}, error) {
	return readVertexElements(b)
}

func readVertexElements(b []byte) ([]interface{}, error) {
	// remove surrounding brackets
	b = b[1 : len(b)-1]

	var ds []interface{}
	for len(b) > 0 {
		if len(ds) > 0 {
			// remove comma
			b = b[1:]
		}

		advance, data, err := readVertexElement(b)
		if err != nil {
			return nil, err
		}
		if data == nil {
			ds = append(ds, nil)
		} else {
			ds = append(ds, data)
		}

		b = b[advance:]
	}

	return ds, nil
}

func readVertexElement(b []byte) (advance int, data *entityData, err error) {
	if bytes.HasPrefix(b, nullElementValue) {
		advance = len(nullElementValue)
		return
	}

	m := vertexCoreRegexp.FindSubmatch(b)
	if m == nil {
		err = fmt.Errorf("bad vertex representation: %s", b)
		return
	}
	advance = len(m[0])

	props, err := readJSONObject(b[advance:])
	if err != nil {
		err = errors.New("invalid vertex properties: " + err.Error())
		return
	}
	advance += len(props)

	data, err = makeVertexData(m[1], m[2], props)
	return
}

// VertexHeader may be used as an embedded field of any struct to change the
// struct to an entity for vertex.
type VertexHeader struct {
	Vertex
	Valid      bool // Valid is true if the vertex is not NULL
	VertexCore      // VertexCore is valid only if Valid is true
}

// SaveEntity implements EntitySaver interface.
func (h *VertexHeader) SaveEntity(valid bool, core interface{}) error {
	h.Valid = valid
	if !valid {
		return nil
	}

	c, ok := core.(VertexCore)
	if !ok {
		return fmt.Errorf("invalid vertex core: %T", core)
	}

	h.VertexCore = c

	return nil
}

// BasicVertex can be used to scan the value from the database driver as a
// vertex.
//
// This is a reference implementation of an entity for vertex using all the
// basic building blocks(Vertex, VertexCore, VertexHeader, EntitySaver,
// PropertiesSaver, and ScanEntity.)
type BasicVertex struct {
	VertexHeader
	Properties map[string]interface{}
}

func (v BasicVertex) String() string {
	if v.Valid {
		p, _ := json.Marshal(v.Properties)
		return fmt.Sprintf("%s[%s]%s", v.Label, v.Id, p)
	} else {
		return "NULL"
	}
}

// SaveProperties implements PropertiesSaver interface. It calls json.Unmarshal
// to unmarshal b and store the result in Properties.
func (v *BasicVertex) SaveProperties(b []byte) error {
	err := json.Unmarshal(b, &v.Properties)
	if err != nil {
		return errors.New("invalid vertex properties: " + err.Error())
	}
	return nil
}

// Scan implements the database/sql Scanner interface. It calls ScanEntity.
func (v *BasicVertex) Scan(src interface{}) error {
	return ScanEntity(src, v)
}

type basicVertexArray []BasicVertex

func (a *basicVertexArray) Scan(src interface{}) error {
	if src == nil {
		*a = nil
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid source for _vertex: %T", src)
	}
	if len(b) < 1 {
		return fmt.Errorf("invalid source for _vertex: %v", b)
	}

	ds, err := readVertexElements(b)
	if err != nil {
		return errors.New("failed to read vertex elements: " + err.Error())
	}

	vs := make([]BasicVertex, len(ds))
	for i, d := range ds {
		err = vs[i].Scan(d)
		if err != nil {
			return err
		}
	}

	*a = vs
	return nil
}

func (a basicVertexArray) Value() (driver.Value, error) {
	return nil, errors.New("Value() on an array of vertex is not supported")
}
