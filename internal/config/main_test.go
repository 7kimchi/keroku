package config

import (
	"errors"
	"io/fs"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

const fakeToken = "fake.test.token"

// env builds a Source from a map plus optional files.
func env(vars map[string]string, files map[string]string) Source {
	return Source{
		Getenv: func(k string) string { return vars[k] },
		ReadFile: func(p string) ([]byte, error) {
			if v, ok := files[p]; ok {
				return []byte(v), nil
			}
			return nil, &fs.PathError{Op: "open", Path: p, Err: fs.ErrNotExist}
		},
	}
}

func base() map[string]string {
	return map[string]string{"DISCORD_TOKEN": fakeToken, "DATABASE_URL": "postgres://u:p@db/keroku"}
}

func mustFail(t *testing.T, vars map[string]string, files map[string]string) error {
	t.Helper()
	_, err := Load(env(vars, files))
	if err == nil {
		t.Fatalf("want error for %v", vars)
	}
	return err
}

func with(k, v string) map[string]string {
	m := base()
	m[k] = v
	return m
}

func TestLoadNilSource(t *testing.T) {
	if _, err := Load(Source{}); err == nil {
		t.Fatal("nil source accepted")
	}
}

func TestReadFileErrorIsWrapped(t *testing.T) {
	m := base()
	delete(m, "DISCORD_TOKEN")
	m["DISCORD_TOKEN_FILE"] = "/missing"
	if err := mustFail(t, m, nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("got %v", err)
	}
}
