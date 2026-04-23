package docker

import "testing"

func TestParseContextPS(t *testing.T) {
	data := []byte(`{"ID":"abc123","Names":"proj-api-1","Image":"api:latest","State":"running","Status":"Up 5 minutes (healthy)","Ports":"127.0.0.1:8080->80/tcp","Labels":"com.docker.compose.project=proj,com.docker.compose.service=api"}
{"ID":"def456","Names":"proj-db-1","Image":"postgres:16","State":"exited","Status":"Exited (0) 2 minutes ago","Ports":"","Labels":"com.docker.compose.project=proj,com.docker.compose.service=db"}
`)

	states := parseContextPS(data, "proj")
	if len(states) != 2 {
		t.Fatalf("expected 2 states, got %d", len(states))
	}
	if states[0].Service != "api" || states[0].Status != "running" || states[0].Health != "healthy" {
		t.Fatalf("api state mismatch: %#v", states[0])
	}
	if len(states[0].Ports) != 1 || states[0].Ports[0] != "127.0.0.1:8080->80/tcp" {
		t.Fatalf("ports not parsed: %#v", states[0].Ports)
	}
	if states[1].Service != "db" || states[1].Status != "stopped" {
		t.Fatalf("db state mismatch: %#v", states[1])
	}
}

func TestParseLabels(t *testing.T) {
	labels := parseLabels("a=1,com.docker.compose.service=api,flag")
	if labels["a"] != "1" {
		t.Fatalf("a label = %q", labels["a"])
	}
	if labels["com.docker.compose.service"] != "api" {
		t.Fatalf("service label = %q", labels["com.docker.compose.service"])
	}
	if _, ok := labels["flag"]; !ok {
		t.Fatal("expected label without value to be retained")
	}
}
