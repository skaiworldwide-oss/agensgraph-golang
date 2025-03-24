package ag_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/skaiworldwide-oss/agensgraph-golang"
)

type knows struct {
	ag.Edge
	meta struct {
		null bool
		id   ag.GraphId
	}
	who   ag.GraphId
	whom  ag.GraphId
	since yearMonth
}

type knowsBody struct {
	Type  string
	Since json.RawMessage
}

type yearMonth struct {
	Year  int
	Month int
}

func (e knows) String() string {
	if e.meta.null {
		return "NULL"
	} else {
		return fmt.Sprintf("%s knows %s since %d, %d", e.who, e.whom, e.since.Month, e.since.Year)
	}
}

func (e *knows) SaveEntity(valid bool, core interface{}) error {
	e.meta.null = !valid
	if !valid {
		return nil
	}

	c, ok := core.(ag.EdgeCore)
	if !ok {
		return fmt.Errorf("invalid edge core: %T", core)
	}

	e.meta.id = c.Id
	e.who = c.Start
	e.whom = c.End
	return nil
}

func (e *knows) SaveProperties(b []byte) error {
	var body knowsBody
	err := json.Unmarshal(b, &body)
	if err != nil {
		return err
	}

	switch body.Type {
	case "array":
		var ym [2]int
		err = json.Unmarshal(body.Since, &ym)
		if err != nil {
			return err
		}
		e.since.Year, e.since.Month = ym[0], ym[1]
	case "object":
		err := json.Unmarshal(body.Since, &e.since)
		if err != nil {
			return err
		}
	default:
		log.Panicf("unknown body type: %q", body.Type)
	}

	return nil
}

func (e *knows) Scan(src interface{}) error {
	return ag.ScanEntity(src, e)
}

func ExampleScanEntity() {
	ds := [][]byte{
		[]byte(`knows[4.1][3.1,3.2]{"type": "array", "since": [1970, 1]}`),
		[]byte(`knows[4.2][3.3,3.4]{"type": "object", "since": {"year": 2009, "month": 10}}`),
	}
	for _, d := range ds {
		var r knows
		err := r.Scan(d)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Printf("%s\n", r)
		}
	}
}
