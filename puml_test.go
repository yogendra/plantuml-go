package main

import (
	"testing"
)

func TestEncodeAsText(t *testing.T) {
	expected := "SYWkIImgAStDuNBAJrBGjLDmpCbCJbMmKiX8pSd9vt98pKifpSq11000__y0"
	input := `@startuml
Bob -> Alice : hello
@enduml`
	actual := encodeAsTextFormat([]byte(input))
	if actual != expected {
		t.Fatalf("Expected: %s\nActual:%s", expected, actual)
	}
}
