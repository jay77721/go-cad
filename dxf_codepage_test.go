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

// buildDXFWithText writes a minimal ASCII DXF with one TEXT entity. The
// codepage header comes before $ACADVER on purpose: the final version is only
// known at ENDSEC, so the reader must re-apply the codepage once the header is
// complete.
func buildDXFWithText(version, codepage string, text []byte) []byte {
	head := "0\nSECTION\n2\nHEADER\n9\n$DWGCODEPAGE\n3\n" + codepage +
		"\n9\n$ACADVER\n1\n" + version +
		"\n0\nENDSEC\n0\nSECTION\n2\nENTITIES\n0\nTEXT\n5\n20\n8\n0\n" +
		"10\n0\n20\n0\n30\n0\n40\n2\n1\n"
	tail := "\n0\nENDSEC\n0\nEOF\n"
	return append(append([]byte(head), text...), []byte(tail)...)
}

func firstDXFText(t *testing.T, doc *Document) string {
	t.Helper()
	for _, ent := range doc.ModelSpace {
		if text, ok := ent.(*entity.EntText); ok {
			return text.Text
		}
	}
	t.Fatal("no TEXT entity in model space")
	return ""
}

func TestParseDXFCodepageR2000Windows1252(t *testing.T) {
	doc, err := ParseDXF(buildDXFWithText("AC1015", "ANSI_1252", []byte("S = 27,72 m\xb2")))
	if err != nil {
		t.Fatalf("ParseDXF: %v", err)
	}
	if got, want := firstDXFText(t, doc), "S = 27,72 m²"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
}

func TestParseDXFCodepageR2004GBK(t *testing.T) {
	doc, err := ParseDXF(buildDXFWithText("AC1018", "ANSI_936", []byte{0xd6, 0xd0}))
	if err != nil {
		t.Fatalf("ParseDXF: %v", err)
	}
	if got, want := firstDXFText(t, doc), "中"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
}

func TestParseDXFCodepageR2007StaysUTF8(t *testing.T) {
	doc, err := ParseDXF(buildDXFWithText("AC1021", "ANSI_1252", []byte("S = 27,72 m²")))
	if err != nil {
		t.Fatalf("ParseDXF: %v", err)
	}
	if got, want := firstDXFText(t, doc), "S = 27,72 m²"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
}

func TestParseDXFCodepagePreR13Windows1252(t *testing.T) {
	doc, err := ParseDXF(buildDXFWithText("AC1009", "ANSI_1252", []byte("S = 27,72 m\xb2")))
	if err != nil {
		t.Fatalf("ParseDXF: %v", err)
	}
	if got, want := firstDXFText(t, doc), "S = 27,72 m²"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
}
