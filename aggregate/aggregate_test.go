package aggregate

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestAggregateZero(t *testing.T) {
	var a ID

	// Ensure that IsZero correctly identifies a zero value
	if !a.IsZero() {
		t.Fatalf("ID '%s' is not zero", a)
	}
	// Generate a zero ID
	a = NewID(ResourceType_INVALID_RESOURCE, "")
	if !a.IsZero() {
		t.Fatalf("Zero ID not identified as such")
	}

	// Aggregate root must not be empty
	a = NewID(ResourceType_INVALID_RESOURCE, "ID")
	if !a.IsZero() {
		t.Fatalf("Zero AggregateRoot in ID not identified as such")
	}

	// Non-zero aggregateID
	a = NewID(42, "")
	if a.IsZero() {
		t.Fatalf("ID '%s' is zero", a)
	}
}

func TestEncodingAggregate(t *testing.T) {
	var i ID
	var s = "10.55.220.225.MEMORY.1"

	// 1. Encode/decode chain
	if err := json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, s)), &i); err != nil {
		t.Fatalf("failed to unmarshal legitimate ID %q: %s", s, err)
	} else if i.String() != s {
		t.Fatalf("unmarshal changed the ID value: %s <> %s", i, s)
	} else if b, err := json.Marshal(i); err != nil {
		t.Fatalf("failed to marshal %s: %s", i, err)
	} else if string(b) != fmt.Sprintf(`"%s"`, s) {
		t.Fatalf("encoding did not produce original form: %q", string(b))
	}

	// 2. Ensure it detects invalid forms
	if err := json.Unmarshal([]byte(`""`), &i); err == nil {
		t.Fatalf("expected error unmarshalling empty string, but got nil")
	} else if err := json.Unmarshal([]byte(`"THIS RESOURCE TYPE IS INVALID"`), &i); err == nil {
		t.Fatalf("expected error unmarshalling invalid resource type, but got nil")
	} else if err := json.Unmarshal([]byte(`"1.2.3.MEMORY.test"`), &i); err == nil { // invalid IP
		t.Fatalf("expected error unmarshalling invalid IP address, but got nil")
	} else if err := json.Unmarshal([]byte(`"INVALID_RESOURCE"`), &i); err != nil {
		t.Fatalf("failed to unmarshal invalid resource specifier: %s", err)
	} else if i.String() != "INVALID_RESOURCE" {
		t.Fatalf("unable to correctly deserialize invalid resource: got %q", i)
	}

	// 3. IP address variations (which we have to test until we have node IDs)
	s = "10.55.220.27.CPU"
	if err := json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, s)), &i); err != nil {
		t.Fatalf("failed to unmarshal %s resource specifier: %s", s, err)
	} else if i.String() != s {
		t.Fatalf("unmarshaling garbled input: expected %q, got %q", s, i)
	}
	s = "10.55.220.27"
	if err := json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, s)), &i); err != nil {
		t.Fatalf("failed to unmarshal %s resource specifier: %s", s, err)
	} else if i.String() != s+".INVALID_RESOURCE" {
		t.Fatalf("unmarshaling garbled input: expected %q, got %q", s, i)
	}
}
