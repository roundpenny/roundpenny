// Copyright (c) 2026 RoundPenny. All rights reserved.

package cache

import (
	"context"
	"testing"
	"time"
)

func newMockClient() *Client {
	return &Client{mock: true, data: make(map[string]string)}
}

func TestClient_GetSet(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	err := c.Set(ctx, "key1", "val1", time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	v, err := c.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if v != "val1" {
		t.Fatalf("got %q, want %q", v, "val1")
	}

	_, err = c.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestClient_Delete(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	c.Set(ctx, "key1", "val1", time.Minute)
	c.Del(ctx, "key1")

	_, err := c.Get(ctx, "key1")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestClient_Exists(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	c.Set(ctx, "key1", "val1", time.Minute)

	ok, err := c.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !ok {
		t.Fatal("expected key to exist")
	}

	c.Del(ctx, "key1")
	ok, err = c.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if ok {
		t.Fatal("expected key to not exist")
	}
}

func TestClient_Expire(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()
	if err := c.Expire(ctx, "key1", time.Minute); err != nil {
		t.Fatalf("Expire failed: %v", err)
	}
}

func TestClient_Incr(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	n, err := c.Incr(ctx, "counter")
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("got %d, want 1", n)
	}

	n, err = c.Incr(ctx, "counter")
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if n != 2 {
		t.Fatalf("got %d, want 2", n)
	}
}

func TestClient_IncrBy(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	n, err := c.IncrBy(ctx, "counter", 5)
	if err != nil {
		t.Fatalf("IncrBy failed: %v", err)
	}
	if n != 5 {
		t.Fatalf("got %d, want 5", n)
	}
}

func TestClient_TTL(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	d, err := c.TTL(ctx, "key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if d != -1 {
		t.Fatalf("got %v, want -1", d)
	}
}

func TestClient_Incr_zero(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	c.data["zero"] = "0"
	n, err := c.Incr(ctx, "zero")
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("got %d, want 1", n)
	}
}
