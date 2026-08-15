package core

import "testing"

func TestBotPeerFlagIsAlwaysVeryThick(t *testing.T) {
	if botPeerFlag != "very_thick" {
		t.Fatalf("botPeerFlag = %q, want very_thick", botPeerFlag)
	}
}
