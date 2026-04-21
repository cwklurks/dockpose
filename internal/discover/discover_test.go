package discover

import (
	"os"
	"path/filepath"
	"testing"
)

const simpleYAML = `services:
  web:
    image: nginx
`

const profilesYAML = `services:
  web:
    image: nginx
    profiles: ["frontend"]
  worker:
    image: busybox
    profiles: ["backend"]
`

const depsYAML = `services:
  db:
    image: postgres
  api:
    image: myapi
    depends_on:
      - db
  web:
    image: nginx
    depends_on:
      api:
        condition: service_started
`

const cycleYAML = `services:
  a:
    image: x
    depends_on:
      - b
  b:
    image: x
    depends_on:
      - c
  c:
    image: x
    depends_on:
      - a
`

func writeFixture(t *testing.T, root, name, body string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDiscover(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "simple", simpleYAML)
	writeFixture(t, root, "with_profiles", profilesYAML)
	writeFixture(t, root, "with_deps", depsYAML)
	writeFixture(t, root, "with_cycle", cycleYAML)

	got, err := Discover([]string{root}, 3)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	names := map[string]bool{}
	for _, c := range got {
		names[c.Name] = true
	}

	for _, want := range []string{"simple", "with_profiles", "with_deps"} {
		if !names[want] {
			t.Errorf("expected %q in results, got %+v", want, got)
		}
	}
	if names["with_cycle"] {
		t.Errorf("with_cycle should be skipped (not a valid DAG), got %+v", got)
	}
}
