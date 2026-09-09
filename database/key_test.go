package database

import (
	"slices"
	"strings"
	"testing"
)

type alpha struct {
	key string
}

func (a *alpha) SetKey(k string) { a.key = k }
func (a *alpha) GetKey() string  { return a.key }

type Beta struct {
	key string
}

func (b *Beta) SetKey(k string) { b.key = k }
func (b *Beta) GetKey() string  { return b.key }

func TestEntityNameForModel(t *testing.T) {
	if got := EntityNameForModel(&alpha{}); got != "alpha" {
		t.Errorf("EntityNameForModel() = %q, want %q", got, "alpha")
	}

	if got := EntityNameForModel(&Beta{}); got != "beta" {
		t.Errorf("EntityNameForModel() = %q, want %q", got, "beta")
	}
}

func TestPrefixForModel(t *testing.T) {
	if got := PrefixForModel(&Beta{}); got != "beta:" {
		t.Errorf("PrefixForModel() = %q, want %q", got, "beta:")
	}
}

func TestNewKeyIsPrefixedAndUnique(t *testing.T) {
	seen := make(map[string]bool)

	for range 128 {
		key := NewKey(&Beta{})

		if strings.HasPrefix(key, "beta:") == false {
			t.Fatalf("NewKey() = %q, want a beta: prefix", key)
		}
		if seen[key] == true {
			t.Fatalf("NewKey() returned %q twice", key)
		}

		seen[key] = true
	}
}

func TestNewKeysSortIntoCreationOrder(t *testing.T) {
	var keys []string
	for range 32 {
		keys = append(keys, NewKey(&Beta{}))
	}

	shuffled := slices.Clone(keys)
	slices.Reverse(shuffled)

	SortKeys(shuffled)

	if slices.Equal(shuffled, keys) == false {
		t.Errorf("SortKeys() did not restore creation order")
	}
}

func TestSortKeysHandlesKeysWithoutUUIDs(t *testing.T) {
	keys := []string{"config", "activeblock", NewKey(&Beta{})}

	SortKeys(keys)

	if len(keys) != 3 {
		t.Errorf("SortKeys() changed the number of keys")
	}
	if slices.Contains(keys, "config") == false ||
		slices.Contains(keys, "activeblock") == false {
		t.Errorf("SortKeys() dropped a key")
	}
}

func TestSortKeysIsStableAcrossCalls(t *testing.T) {
	keys := []string{"config", "activeblock"}
	for range 8 {
		keys = append(keys, NewKey(&Beta{}))
	}

	first := slices.Clone(keys)
	SortKeys(first)

	second := slices.Clone(keys)
	slices.Reverse(second)
	SortKeys(second)

	if slices.Equal(first, second) == false {
		t.Errorf("SortKeys() = %v and %v for the same input set", first, second)
	}
}

func TestGetOrderedKeys(t *testing.T) {
	rows := make(map[string]*Beta)

	var keys []string
	for range 16 {
		key := NewKey(&Beta{})
		keys = append(keys, key)
		rows[key] = &Beta{key: key}
	}

	got := GetOrderedKeys(rows)

	if slices.Equal(got, keys) == false {
		t.Errorf("GetOrderedKeys() = %v, want creation order %v", got, keys)
	}
}

func TestGetOrderedKeysOnAnEmptyMap(t *testing.T) {
	if got := GetOrderedKeys(map[string]*Beta{}); len(got) != 0 {
		t.Errorf("GetOrderedKeys() = %v, want nothing", got)
	}
}
