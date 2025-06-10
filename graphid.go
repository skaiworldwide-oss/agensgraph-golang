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
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// GraphId is a unique ID for a vertex and an edge.
type GraphId struct {
	// Valid is true if GraphId is not NULL
	Valid bool

	b []byte
}

var nullGraphId = GraphId{}

var graphIdRegexp = regexp.MustCompile(`^(\d+)\.(\d+)$`)

// NewGraphId returns GraphId of str if str is between "1.1" and
// "65535.281474976710656". If str is "NULL", it returns GraphId whose Valid is
// false. Otherwise, it returns an error.
func NewGraphId(str string) (GraphId, error) {
	if str == "NULL" {
		return nullGraphId, nil
	}

	err := validateGraphId(str)
	if err != nil {
		return GraphId{}, err
	}

	return GraphId{true, []byte(str)}, nil
}

const (
	labelBit = 16
	localBit = 48
)

func validateGraphId(str string) error {
	m := graphIdRegexp.FindStringSubmatch(str)
	if m == nil {
		return fmt.Errorf("bad graphid representation: %q", str)
	}

	i, err := strconv.ParseUint(m[1], 10, labelBit)
	if err != nil {
		return errors.New("invalid label ID: " + err.Error())
	}
	if i == 0 {
		return fmt.Errorf("invalid label ID: %d", i)
	}

	i, err = strconv.ParseUint(m[2], 10, localBit)
	if err != nil {
		return errors.New("invalid local ID: " + err.Error())
	}
	if i == 0 {
		return fmt.Errorf("invalid local ID: %d", i)
	}

	return nil
}

// Equal reports whether gid and x are the same GraphId.
func (gid GraphId) Equal(x GraphId) bool {
	if !gid.Valid || !x.Valid {
		return false
	}
	return bytes.Equal(gid.b, x.b)
}

func (gid GraphId) String() string {
	if gid.Valid {
		return string(gid.b)
	} else {
		return "NULL"
	}
}

// Scan implements the database/sql Scanner interface.
func (gid *GraphId) Scan(src interface{}) error {
	if src == nil {
		gid.Valid, gid.b = false, nil
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid source for graphid: %T", src)
	}
	if len(b) < 1 {
		return fmt.Errorf("invalid source for graphid: %v", b)
	}

	err := validateGraphId(string(b))
	if err != nil {
		return err
	}

	gid.Valid, gid.b = true, append([]byte(nil), b...)
	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (gid GraphId) Value() (driver.Value, error) {
	if gid.Valid {
		return gid.b, nil
	} else {
		return nil, nil
	}
}

type graphIdArray []GraphId

// separated by comma (see graphid in pg_type.h)
const graphIdSeparator = byte(054)

func (a *graphIdArray) Scan(src interface{}) error {
	if src == nil {
		*a = nil
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid source for _graphid: %T", src)
	}
	if len(b) < 1 {
		return fmt.Errorf("invalid source for _graphid: %v", b)
	}

	// remove surrounding braces
	b = b[1 : len(b)-1]

	// bytes.Split() returns [][]byte{[]byte{}} even if len(b) < 1.
	// In this case, return empty []GraphId to distinguish between NULL and
	// empty _graphid.
	if len(b) < 1 {
		*a = []GraphId{}
		return nil
	}

	ss := bytes.Split(b, []byte{graphIdSeparator})

	gids := make([]GraphId, len(ss))
	for i, s := range ss {
		if bytes.Equal(s, nullElementValue) {
			gids[i] = nullGraphId
			continue
		}

		err := gids[i].Scan(s)
		if err != nil {
			return errors.New("bad _graphid representation: " + err.Error())
		}
	}

	*a = gids
	return nil
}

func (a graphIdArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// '{' + "d.d,"*n - ',' + '}' = 1+4*n-1+1 = 4*n+1
		b := make([]byte, 1, 4*n+1)
		b[0] = byte('{')
		for i := 0; i < n; i++ {
			if i > 0 {
				b = append(b, graphIdSeparator)
			}
			if a[i].Valid {
				val, _ := a[i].Value()
				b = append(b, val.([]byte)...)
			} else {
				b = append(b, nullElementValue...)
			}
		}
		b = append(b, '}')

		return b, nil
	}

	return []byte("{}"), nil
}
