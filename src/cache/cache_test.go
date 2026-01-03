package cache_test

import (
	"testing"
	"time"

	"github.com/howmanysmall/npm-registry-mcp/src/cache"
)

func TestCache_GetSet(t *testing.T) {
	t.Parallel()

	c := cache.New(5*time.Minute, 10*time.Minute)

	// Set a value
	c.Set("key1", "value1")

	// Get it back
	val, found := cache.Get[string](c, "key1")
	if !found {
		t.Fatal("expected to find key1")
	}

	if val != "value1" {
		t.Errorf("expected value1, got %s", val)
	}

	// Get missing key
	_, found = cache.Get[string](c, "missing")
	if found {
		t.Error("expected not to find missing key")
	}
}

func TestCache_TypedGet(t *testing.T) {
	t.Parallel()

	type myStruct struct {
		Name string
		Age  int
	}

	c := cache.New(5*time.Minute, 10*time.Minute)

	c.Set("person", myStruct{Name: "Alice", Age: 30})

	val, found := cache.Get[myStruct](c, "person")
	if !found {
		t.Fatal("expected to find person")
	}

	if val.Name != "Alice" || val.Age != 30 {
		t.Errorf("unexpected value: %+v", val)
	}
}

func TestCache_TypeMismatch(t *testing.T) {
	t.Parallel()

	c := cache.New(5*time.Minute, 10*time.Minute)
	c.Set("key", "string value")

	// Attempt to get as int should fail
	_, found := cache.Get[int](c, "key")
	if found {
		t.Error("expected type mismatch to return not found")
	}
}

func TestCache_SetWithExpiration(t *testing.T) {
	t.Parallel()

	c := cache.New(5*time.Minute, 10*time.Minute)
	c.SetWithExpiration("key", "value", 1*time.Hour)

	val, found := cache.Get[string](c, "key")
	if !found {
		t.Fatal("expected to find key")
	}

	if val != "value" {
		t.Errorf("expected value, got %s", val)
	}
}

func TestCache_Delete(t *testing.T) {
	t.Parallel()

	c := cache.New(5*time.Minute, 10*time.Minute)
	c.Set("key", "value")
	c.Delete("key")

	_, found := cache.Get[string](c, "key")
	if found {
		t.Error("expected key to be deleted")
	}
}

func TestCache_Flush(t *testing.T) {
	t.Parallel()

	c := cache.New(5*time.Minute, 10*time.Minute)
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Flush()

	_, found1 := cache.Get[string](c, "key1")
	_, found2 := cache.Get[string](c, "key2")

	if found1 || found2 {
		t.Error("expected all keys to be flushed")
	}
}
