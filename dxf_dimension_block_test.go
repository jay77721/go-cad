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
	"testing"

	"github.com/unitedrhino/go-cad/internal/entity"
)

// A DIMENSION's group code 2 names its anonymous block (*D1, ...). The block
// holds the geometry and text AutoCAD computed for the dimension. The DXF
// reader used to drop that name, so the block was never expanded.
const dxfDimensionWithBlock = `0
SECTION
2
HEADER
9
$ACADVER
1
AC1015
9
$DWGCODEPAGE
3
ANSI_1252
0
ENDSEC
0
SECTION
2
BLOCKS
0
BLOCK
5
10
2
*D1
70
0
10
0
20
0
30
0
0
TEXT
5
11
8
0
10
10
20
20
30
0
40
2
1
DIMTEXT
0
ENDBLK
5
12
0
ENDSEC
0
SECTION
2
ENTITIES
0
DIMENSION
5
20
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

func TestParseDXFDimensionLinksAnonymousBlock(t *testing.T) {
	doc, err := ParseDXF([]byte(dxfDimensionWithBlock))
	if err != nil {
		t.Fatalf("ParseDXF: %v", err)
	}
	var dim *entity.EntDimension
	for _, ent := range doc.ModelSpace {
		if d, ok := ent.(*entity.EntDimension); ok {
			dim = d
			break
		}
	}
	if dim == nil {
		t.Fatal("no DIMENSION in model space")
	}
	if dim.AnonymousBlock == 0 {
		t.Fatal("DIMENSION.AnonymousBlock = 0, want the *D1 block handle")
	}
	if got := len(doc.Blocks[dim.AnonymousBlock]); got == 0 {
		t.Fatalf("anonymous block %d has no entities", dim.AnonymousBlock)
	}

	texts := SheetTexts(doc)
	found := false
	for _, text := range texts {
		if text.Text == "DIMTEXT" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("dimension block text missing from SheetTexts: %#v", texts)
	}
}

func TestParseDXFDimensionMissingBlockStaysUnlinked(t *testing.T) {
	doc, err := ParseDXF([]byte(`0
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
DIMENSION
5
20
8
0
2
*MISSING
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
`))
	if err != nil {
		t.Fatalf("ParseDXF: %v", err)
	}
	var dim *entity.EntDimension
	for _, ent := range doc.ModelSpace {
		if d, ok := ent.(*entity.EntDimension); ok {
			dim = d
			break
		}
	}
	if dim == nil {
		t.Fatal("no DIMENSION in model space")
	}
	if dim.AnonymousBlock != 0 {
		t.Fatalf("DIMENSION.AnonymousBlock = %d for a missing block, want 0", dim.AnonymousBlock)
	}
}
