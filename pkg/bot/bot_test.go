package bot

import "testing"

func TestStopBeforeStartIsSafeAndPreventsReopen(t *testing.T) {
	b, err := New("test-token")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := b.Stop(); err != nil {
		t.Fatalf("first Stop() error = %v", err)
	}
	if err := b.Stop(); err != nil {
		t.Fatalf("second Stop() error = %v", err)
	}
	if err := b.Start(); err == nil {
		t.Fatal("Start() after Stop() returned nil, want lifecycle error")
	}
}
