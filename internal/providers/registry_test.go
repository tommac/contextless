package providers

import "testing"

// scriptedFake stands for an adapter defined outside the package, such as
// the orchestrator tests' scripted fakes registered under live names.
type scriptedFake struct{ *FakeAdapter }

// IsLive covers every registered provider but the fake, by value and by
// pointer: a provider added to NewAdapter but not to IsLive fails here.
func TestIsLive(t *testing.T) {
	for _, p := range Providers {
		a, err := NewAdapter(p)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := IsLive(a), p != "fake"; got != want {
			t.Errorf("IsLive(%s adapter %T) = %v, want %v", p, a, got, want)
		}
	}
	for _, a := range []Adapter{&ClaudeAdapter{}, &CodexAdapter{}, &GeminiAdapter{}, &KimiAdapter{}, &GrokAdapter{}, &MuseAdapter{}} {
		if !IsLive(a) {
			t.Errorf("IsLive(%T) = false", a)
		}
	}
	if IsLive(scriptedFake{&FakeAdapter{}}) {
		t.Error("an adapter type from outside the package counts as live")
	}
	if IsLive(nil) {
		t.Error("IsLive(nil) = true")
	}
}
