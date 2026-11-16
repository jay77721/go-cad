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

package render

import (
	"bytes"
	"testing"

	"github.com/unitedrhino/go-cad/internal/drawing"
	"github.com/unitedrhino/go-cad/internal/entity"
	"github.com/unitedrhino/go-cad/internal/object"
)

// buildFarGeometryDoc puts a dense geometry cluster far away from a single
// text label. The density-based quantile bounds pick the geometry cluster;
// the text-aware fallback must pick the label's cluster instead.
func buildFarGeometryDoc(textX, textY float64) *drawing.Document {
	doc := &drawing.Document{
		Blocks:       map[uint64][]any{},
		Attribs:      map[uint64]*entity.EntAttrib{},
		LayerColors:  map[uint64]drawing.LayerColor{},
		InternalObjs: map[uint64]*object.ObjGeneric{},
	}
	for i := 0; i < 30; i++ {
		y := -1000.0 + float64(i)*5
		doc.ModelSpace = append(doc.ModelSpace, &entity.EntLine{
			Start: entity.Point3{X: -1000, Y: y, Z: 0},
			End:   entity.Point3{X: -900, Y: y + 2, Z: 0},
		})
	}
	doc.ModelSpace = append(doc.ModelSpace, &entity.EntText{
		Text:      "label",
		Insertion: entity.Point3{X: textX, Y: textY, Z: 0},
		Height:    10,
	})
	return doc
}

func TestSheetContentBoundsPicksTextCluster(t *testing.T) {
	doc := buildFarGeometryDoc(0, 0)
	sp := prepareSheets(doc)
	b := sheetContentBounds(sp.prims)
	if b.Invalid() {
		t.Fatal("sheetContentBounds returned an invalid box")
	}
	if !pointInBox(0, 0, b) {
		t.Fatalf("text-aware bounds %+v do not contain the label at (0,0)", b)
	}
	if b.MinX < -100 {
		t.Fatalf("text-aware bounds %+v dragged in the far geometry cluster", b)
	}
}

func TestFallbackWholeContainsText(t *testing.T) {
	doc := buildFarGeometryDoc(0, 0)
	sheets := DetectSheets(doc)
	if len(sheets) != 1 || sheets[0].Name != "整图" {
		t.Skipf("frame detection produced %d sheets; fallback path not exercised", len(sheets))
	}
	b := sheets[0].Box
	if !pointInBox(0, 0, drawing.Box2{MinX: b[0], MinY: b[1], MaxX: b[2], MaxY: b[3]}) {
		t.Fatalf("fallback sheet box %v does not contain the label", b)
	}
}

func TestRenderPNGOptionsBox(t *testing.T) {
	doc := buildFarGeometryDoc(0, 0)
	box := [4]float64{-20, -20, 20, 20}
	data, err := RenderPNG(doc, RenderOptions{Width: 200, Box: &box})
	if err != nil {
		t.Fatalf("RenderPNG with Box: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("\x89PNG")) {
		t.Fatal("RenderPNG did not return a PNG")
	}
}

func TestRenderAllSheetsViewport(t *testing.T) {
	doc := buildFarGeometryDoc(0, 0)
	results, err := RenderAllSheets(doc, RenderOptions{Width: 200}, "png")
	if err != nil {
		t.Fatalf("RenderAllSheets: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("RenderAllSheets returned no results")
	}
	for _, result := range results {
		vp := result.Viewport
		if vp[2] <= vp[0] || vp[3] <= vp[1] {
			t.Fatalf("sheet %q has an empty viewport %v", result.Name, vp)
		}
	}
}

func TestSheetViewportMargin(t *testing.T) {
	vp := SheetViewport([4]float64{0, 0, 100, 100})
	want := [4]float64{-2, -2, 102, 102}
	for i := range want {
		if vp[i] != want[i] {
			t.Fatalf("SheetViewport = %v, want %v", vp, want)
		}
	}
}
