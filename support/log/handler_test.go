package log

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestUniqueHandler(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, nil)
	uh := NewUniqueHandler(h)
	logger := slog.New(uh)

	// Case 1: Multiple With calls
	logger.With("comp", "A").With("comp", "B").Info("test 1")

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatal(err)
	}

	if m["comp"] != "B" {
		t.Errorf("expected comp=B, got %v", m["comp"])
	}
	buf.Reset()

	// Case 2: With followed by Record attributes
	logger.With("mod", "repo-go").Info("test 2", "mod", "repo-kafka-go")
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["mod"] != "repo-kafka-go" {
		t.Errorf("expected mod=repo-kafka-go, got %v", m["mod"])
	}
	buf.Reset()

	// Case 3: Multiple unique keys (including app)
	logger.With("app", "kapbit", "comp", "C", "mod", "M1").With("mod", "M2").Info("test 3", "node-id", "N1")
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["app"] != "kapbit" || m["comp"] != "C" || m["mod"] != "M2" || m["node-id"] != "N1" {
		t.Errorf("unexpected values: app=%v, comp=%v, mod=%v, node-id=%v", m["app"], m["comp"], m["mod"], m["node-id"])
	}
	buf.Reset()

	// Case 4: exploded properties
	logger.With("partition", 1, "txn-id", "T1").Info("test 4", "partition", 2)
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["partition"].(float64) != 2 || m["txn-id"] != "T1" {
		t.Errorf("expected partition=2, txn-id=T1, got partition=%v, txn-id=%v", m["partition"], m["txn-id"])
	}
}
