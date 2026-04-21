package stack

import (
	"reflect"
	"testing"
)

func TestBuildGraphLinearChain(t *testing.T) {
	// A depends on B, B depends on C
	services := map[string]ServiceConfig{
		"A": {DependsOn: []string{"B"}},
		"B": {DependsOn: []string{"C"}},
		"C": {},
	}
	adj, err := BuildGraph(services)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(adj["C"], []string{"B"}) {
		t.Errorf("C dependents = %v, want [B]", adj["C"])
	}
	if !reflect.DeepEqual(adj["B"], []string{"A"}) {
		t.Errorf("B dependents = %v, want [A]", adj["B"])
	}
	if len(adj["A"]) != 0 {
		t.Errorf("A dependents = %v, want []", adj["A"])
	}
}

func TestBuildGraphForkAndDiamond(t *testing.T) {
	// Fork: B and C depend on A
	fork := map[string]ServiceConfig{
		"A": {},
		"B": {DependsOn: []string{"A"}},
		"C": {DependsOn: []string{"A"}},
	}
	adj, err := BuildGraph(fork)
	if err != nil {
		t.Fatalf("fork: %v", err)
	}
	if !reflect.DeepEqual(adj["A"], []string{"B", "C"}) {
		t.Errorf("fork A dependents = %v, want [B C]", adj["A"])
	}

	// Diamond: B,C depend on A; D depends on B,C
	diamond := map[string]ServiceConfig{
		"A": {},
		"B": {DependsOn: []string{"A"}},
		"C": {DependsOn: []string{"A"}},
		"D": {DependsOn: []string{"B", "C"}},
	}
	adj2, err := BuildGraph(diamond)
	if err != nil {
		t.Fatalf("diamond: %v", err)
	}
	if !reflect.DeepEqual(adj2["A"], []string{"B", "C"}) {
		t.Errorf("diamond A = %v", adj2["A"])
	}
	if !reflect.DeepEqual(adj2["B"], []string{"D"}) {
		t.Errorf("diamond B = %v", adj2["B"])
	}
	if !reflect.DeepEqual(adj2["C"], []string{"D"}) {
		t.Errorf("diamond C = %v", adj2["C"])
	}
}

func TestTopologicalOrder(t *testing.T) {
	services := map[string]ServiceConfig{
		"db":     {},
		"redis":  {},
		"app":    {DependsOn: []string{"db", "redis"}},
		"worker": {DependsOn: []string{"app"}},
	}
	adj, err := BuildGraph(services)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	order, err := TopologicalOrder(adj)
	if err != nil {
		t.Fatalf("topo: %v", err)
	}
	want := []string{"db", "redis", "app", "worker"}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("order = %v, want %v", order, want)
	}
}

func TestDetectCycles(t *testing.T) {
	linear := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": nil,
	}
	if DetectCycles(linear) {
		t.Error("linear graph should not have cycle")
	}
	cyclic := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}
	if !DetectCycles(cyclic) {
		t.Error("cyclic graph should have cycle")
	}
}

func TestDetectCyclesReturnsError(t *testing.T) {
	services := map[string]ServiceConfig{
		"a": {DependsOn: []string{"b"}},
		"b": {DependsOn: []string{"a"}},
	}
	_, err := BuildGraph(services)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}
