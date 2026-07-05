package main

import "testing"

func TestRemoveProfane(t *testing.T) {
	text := "some kerfuFFle and ShArBert but no FORNAX"
	expectedText := "some **** and **** but no ****"
	actualText := removeProfane(text)

	if actualText != expectedText {
		t.Fatalf("expected %v, got %v", expectedText, actualText)
	}
}
