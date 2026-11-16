//
//  Copyright 2026 The InfiniFlow Authors. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//

package cad

import (
	"encoding/json"
	"testing"

	"github.com/unitedrhino/go-cad/internal/entity"
)

// The DXF reader used to leave BaseEntity.TypeName empty, so DumpEntitiesDoc
// emitted UNKNOWN for every row and the dimension/MLEADER compensation could
// not match anything.
const dxfTypeNames = `0
SECTION
2
HEADER
9
$ACADVER
1
AC1015
0
ENDSEC
0
SECTION
2
ENTITIES
0
LINE
5
20
8
0
10
0
20
0
30
0
11
10
21
10
31
0
0
TEXT
5
21
8
0
10
0
20
0
30
0
40
2
1
hello
0
DIMENSION
5
22
8
0
2
*D1
10
0
20
0
30
0
11
50
21
50
31
0
70
32
0
ENDSEC
0
EOF
`

func TestParseDXFEntityTypeNames(t *testing.T) {
	doc, err := ParseDXF([]byte(dxfTypeNames))
	if err != nil {
		t.Fatalf("ParseDXF: %v", err)
	}
	for _, ent := range doc.ModelSpace {
		switch e := ent.(type) {
		case *entity.EntLine:
			if e.TypeName != "LINE" {
				t.Fatalf("LINE TypeName = %q", e.TypeName)
			}
		case *entity.EntText:
			if e.TypeName != "TEXT" {
				t.Fatalf("TEXT TypeName = %q", e.TypeName)
			}
		case *entity.EntDimension:
			if e.TypeName != "DIM_LINEAR" {
				t.Fatalf("DIMENSION TypeName = %q, want DIM_LINEAR", e.TypeName)
			}
		}
	}
}

func TestParseDXFDumpTypesAreKnown(t *testing.T) {
	doc, err := ParseDXF([]byte(dxfTypeNames))
	if err != nil {
		t.Fatalf("ParseDXF: %v", err)
	}
	raw, err := DumpEntitiesDoc(doc)
	if err != nil {
		t.Fatalf("DumpEntitiesDoc: %v", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		t.Fatalf("decode dump: %v", err)
	}
	types := map[string]int{}
	for _, row := range rows {
		name, _ := row["type"].(string)
		types[name]++
	}
	if types["UNKNOWN"] != 0 {
		t.Fatalf("dump still reports UNKNOWN rows: %v", types)
	}
	for _, want := range []string{"LINE", "TEXT", "DIMENSION_LINEAR"} {
		if types[want] == 0 {
			t.Fatalf("dump type %q missing: %v", want, types)
		}
	}
}
