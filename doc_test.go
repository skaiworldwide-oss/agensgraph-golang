/*
Copyright 2018 Bitnine Co., Ltd.

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

package ag_test

import (
	"database/sql"

	"github.com/skaiworldwide-oss/agensgraph-golang"
)

var db *sql.DB

func ExampleGraphId_Scan() {
	var gid ag.GraphId
	err := db.QueryRow(`MATCH (n) RETURN id(n) LIMIT 1`).Scan(&gid)
	if err == nil && gid.Valid {
		// Valid graphid
	} else {
		// An error occurred or the graphid is NULL
	}
}

func ExampleGraphId_Value() {
	gid, _ := ag.NewGraphId("1.1")
	db.QueryRow(`MATCH (n) WHERE id(n) = $1 RETURN n`, gid)
}

func ExampleArray_graphId() {
	var gids []ag.GraphId

	// Scan()
	db.QueryRow(`MATCH ()-[es*]->() RETURN [e IN es | id(e)] LIMIT 1`).Scan(ag.Array(&gids))

	// Value()
	gid0, _ := ag.NewGraphId("1.1")
	gid1, _ := ag.NewGraphId("65535.281474976710655")
	gids = []ag.GraphId{gid0, gid1}
	db.QueryRow(`MATCH (n) WHERE id(n) IN $1 RETURN n LIMIT 1`, ag.Array(gids))
}

func ExampleBasicVertex_Scan() {
	var v ag.BasicVertex
	err := db.QueryRow(`MATCH (n) RETURN n LIMIT 1`).Scan(&v)
	if err == nil && v.Valid {
		// Valid vertex
	} else {
		// An error occurred or the vertex is NULL
	}
}

func ExampleArray_basicVertex() {
	var vs []ag.BasicVertex
	err := db.QueryRow(`MATCH p=(:v)-[:e]->(:v) RETURN nodes(p) LIMIT 1`).Scan(ag.Array(&vs))
	if err == nil && vs != nil {
		// Valid _vertex
	} else {
		// An error occurred or the _vertex is NULL
	}
}

func ExampleBasicEdge_Scan() {
	var e ag.BasicEdge
	err := db.QueryRow(`MATCH ()-[e]->() RETURN e LIMIT 1`).Scan(&e)
	if err == nil && e.Valid {
		// Valid edge
	} else {
		// An error occurred or the edge is NULL
	}
}

func ExampleArray_basicEdge() {
	var es []ag.BasicEdge
	err := db.QueryRow(`MATCH p=(:v)-[:e]->(:v)-[:e]->(:v) RETURN relationships(p) LIMIT 1`).Scan(ag.Array(&es))
	if err == nil && es != nil {
		// Valid _edge
	} else {
		// An error occurred or the _edge is NULL
	}
}

func ExampleBasicPath_Scan() {
	var p ag.BasicPath
	err := db.QueryRow(`MATCH p=(:v)-[:e]->(:v)-[:e]->(:v) RETURN p LIMIT 1`).Scan(&p)
	if err == nil && p.Valid {
		// Valid graphpath
	} else {
		// An error occurred or the graphpath is NULL
	}
}
