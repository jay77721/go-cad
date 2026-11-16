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
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCorpusRegression runs the public CAD corpus when CAD_CORPUS_DIR points
// at a directory with dxf/, dwg/ and dxfb/ subdirectories. It pins the three
// DXF fixes (dimension texts, codepage, type names) and the text-aware
// viewport against real drawings; without the env var it is a no-op so the
// unit tier stays self-contained.
func TestCorpusRegression(t *testing.T) {
	root := os.Getenv("CAD_CORPUS_DIR")
	if root == "" {
		t.Skip("CAD_CORPUS_DIR not set; skipping real-corpus regression")
	}

	type result struct {
		version string
		texts   int
		inside  int
		rows    int
		unknown int
		hasM2   bool
	}
	results := map[string]result{}

	for _, dir := range []string{"dxf", "dwg", "dxfb"} {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			continue
		}
		for _, entry := range entries {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != ".dxf" && ext != ".dxfb" && ext != ".dwg" {
				continue
			}
			name := dir + "/" + entry.Name()
			data, err := os.ReadFile(filepath.Join(root, dir, entry.Name()))
			if err != nil {
				t.Fatalf("%s: read: %v", name, err)
			}
			var doc *Document
			if ext == ".dwg" {
				doc, err = Parse(data)
			} else {
				doc, err = ParseDXF(data)
			}
			if err != nil {
				t.Fatalf("%s: parse: %v", name, err)
			}

			texts := SheetTexts(doc)
			sheets := DetectSheets(doc)
			inside := -1
			if len(sheets) == 1 && sheets[0].Name == "整图" {
				vp := SheetViewport(sheets[0].Box)
				inside = 0
				for _, text := range texts {
					if text.X >= vp[0] && text.X <= vp[2] && text.Y >= vp[1] && text.Y <= vp[3] {
						inside++
					}
				}
			}

			unknown, rows := 0, 0
			if raw, dumpErr := DumpEntitiesDoc(doc); dumpErr == nil {
				var dumped []map[string]any
				if json.Unmarshal([]byte(raw), &dumped) == nil {
					rows = len(dumped)
					for _, row := range dumped {
						if typ, _ := row["type"].(string); typ == "UNKNOWN" {
							unknown++
						}
					}
				}
			}
			hasM2 := false
			for _, text := range texts {
				if strings.Contains(text.Text, "m²") {
					hasM2 = true
				}
			}
			results[name] = result{version: doc.Version(), texts: len(texts), inside: inside, rows: rows, unknown: unknown, hasM2: hasM2}
			t.Logf("%-30s ver=%-6s texts=%3d inside=%3d rows=%3d unknown=%2d m2=%v sheets=%d",
				name, doc.Version(), len(texts), inside, rows, unknown, hasM2, len(sheets))

			if rows > 0 && unknown > 0 {
				t.Errorf("%s: %d/%d dump rows are UNKNOWN", name, unknown, rows)
			}
			if inside >= 0 && len(texts) > 0 && inside == 0 {
				t.Errorf("%s: text-aware fallback put 0/%d texts inside the viewport", name, len(texts))
			}
			if strings.Contains(entry.Name(), "apartment_floor_plan") && inside >= 0 && inside != len(texts) {
				t.Errorf("%s: apartment regression: %d/%d texts inside the viewport", name, inside, len(texts))
			}
			if strings.Contains(entry.Name(), "apartment_floor_plan") && !hasM2 {
				t.Errorf("%s: codepage regression: no text contains m²", name)
			}
			if len(texts) > 0 {
				pngData, renderErr := RenderPNG(doc, RenderOptions{Width: 512})
				if renderErr != nil {
					t.Errorf("%s: RenderPNG: %v", name, renderErr)
				} else if !bytes.HasPrefix(pngData, []byte("\x89PNG")) {
					t.Errorf("%s: RenderPNG returned a non-PNG payload", name)
				}
				sheetsOut, renderErr := RenderAllSheets(doc, RenderOptions{Width: 512}, "png")
				if renderErr != nil {
					t.Errorf("%s: RenderAllSheets: %v", name, renderErr)
				} else {
					for _, sheet := range sheetsOut {
						vp := sheet.Viewport
						if vp[2] <= vp[0] || vp[3] <= vp[1] {
							t.Errorf("%s: sheet %q has an empty viewport %v", name, sheet.Name, vp)
						}
					}
				}
			}
		}
	}

	for name, dwg := range results {
		if !strings.HasPrefix(name, "dwg/") {
			continue
		}
		base := strings.TrimPrefix(name, "dwg/")
		dxfb, ok := results["dxfb/"+strings.TrimSuffix(base, ".dwg")+".dxfb"]
		if !ok {
			continue
		}
		t.Logf("PAIR %-28s dwg texts=%d dxfb texts=%d", base, dwg.texts, dxfb.texts)
		if dwg.texts != dxfb.texts {
			t.Errorf("PAIR %s: dwg texts=%d, dxfb texts=%d", base, dwg.texts, dxfb.texts)
		}
	}
}
