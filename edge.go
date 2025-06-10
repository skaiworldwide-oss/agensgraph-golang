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

// Edge gives any struct an ability to read the value from the database
// driver as an edge if the struct has Edge as its embedded field.
type Edge struct{}

// EdgeCore represents essential data to identify an edge.
type EdgeCore struct {
	Label string
	Id    GraphId
	Start GraphId
	End   GraphId
}

var edgeCoreRegexp = regexp.MustCompile(`^(.+?)\[(\d+\.\d+)\]\[(\d+\.\d+),(\d+\.\d+)\]`)

func (_ Edge) readEntity(b []byte) (*entityData, error) {
	m := edgeCoreRegexp.FindSubmatch(b)
	if m == nil {
		return nil, fmt.Errorf("bad edge representation: %s", b)
	}

	return makeEdgeData(m[1], m[2], m[3], m[4], b[len(m[0]):])
}

func makeEdgeData(label, id, start, end, props []byte) (*entityData, error) {
	var c EdgeCore

	c.Label = string(label)

	err := c.Id.Scan(id)
	if err != nil {
		return nil, errors.New("invalid edge ID: " + err.Error())
	}

	err = c.Start.Scan(start)
	if err != nil {
		return nil, errors.New("invalid edge start ID: " + err.Error())
	}

	err = c.End.Scan(end)
	if err != nil {
		return nil, errors.New("invalid edge end ID: " + err.Error())
	}

	return &entityData{c, props}, nil
}

func (_ Edge) readElements(b []byte) ([]interface{}, error) {
	return readEdgeElements(b)
}

func readEdgeElements(b []byte) ([]interface{}, error) {
	// remove surrounding brackets
	b = b[1 : len(b)-1]

	var ds []interface{}
	for len(b) > 0 {
		if len(ds) > 0 {
			// remove comma
			b = b[1:]
		}

		advance, data, err := readEdgeElement(b)
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

func readEdgeElement(b []byte) (advance int, data *entityData, err error) {
	if bytes.HasPrefix(b, nullElementValue) {
		advance = len(nullElementValue)
		return
	}

	m := edgeCoreRegexp.FindSubmatch(b)
	if m == nil {
		err = fmt.Errorf("bad edge representation: %s", b)
		return
	}
	advance = len(m[0])

	props, err := readJSONObject(b[advance:])
	if err != nil {
		err = errors.New("invalid edge properties: " + err.Error())
		return
	}
	advance += len(props)

	data, err = makeEdgeData(m[1], m[2], m[3], m[4], props)
	return
}

// EdgeHeader may be used as an embedded field of any struct to change the
// struct to an entity for edge.
type EdgeHeader struct {
	Edge
	Valid    bool // Valid is true if the edge is not NULL
	EdgeCore      // EdgeCore is valid only if Valid is true
}

// SaveEntity implements EntitySaver interface.
func (h *EdgeHeader) SaveEntity(valid bool, core interface{}) error {
	h.Valid = valid
	if !valid {
		return nil
	}

	c, ok := core.(EdgeCore)
	if !ok {
		return fmt.Errorf("invalid edge core: %T", core)
	}

	h.EdgeCore = c

	return nil
}

// BasicEdge can be used to scan the value from the database driver as an edge.
//
// This is a reference implementation of an entity for edge using all the basic
// building blocks(Edge, EdgeCore, EdgeHeader, EntitySaver, PropertiesSaver,
// and ScanEntity.)
type BasicEdge struct {
	EdgeHeader
	Properties map[string]interface{}
}

func (e BasicEdge) String() string {
	if e.Valid {
		p, _ := json.Marshal(e.Properties)
		return fmt.Sprintf("%s[%s][%s,%s]%s", e.Label, e.Id, e.Start, e.End, p)
	} else {
		return "NULL"
	}
}

// SaveProperties implements PropertiesSaver interface. It calls json.Unmarshal
// to unmarshal b and store the result in Properties.
func (e *BasicEdge) SaveProperties(b []byte) error {
	err := json.Unmarshal(b, &e.Properties)
	if err != nil {
		return errors.New("invalid edge properties: " + err.Error())
	}
	return nil
}

// Scan implements the database/sql Scanner interface. It calls ScanEntity.
func (e *BasicEdge) Scan(src interface{}) error {
	return ScanEntity(src, e)
}

type basicEdgeArray []BasicEdge

func (a *basicEdgeArray) Scan(src interface{}) error {
	if src == nil {
		*a = nil
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid source for _edge: %T", src)
	}
	if len(b) < 1 {
		return fmt.Errorf("invalid source for _edge: %v", b)
	}

	ds, err := readEdgeElements(b)
	if err != nil {
		return errors.New("failed to read edge elements: " + err.Error())
	}

	es := make([]BasicEdge, len(ds))
	for i, d := range ds {
		err = es[i].Scan(d)
		if err != nil {
			return err
		}
	}

	*a = es
	return nil
}

func (a basicEdgeArray) Value() (driver.Value, error) {
	return nil, errors.New("Value() on an array of edge is not supported")
}
