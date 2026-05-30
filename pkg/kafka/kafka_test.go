package kafka

import (
	"context"
	"encoding/json"
	"testing"
)

func TestMessage_struct(t *testing.T) {
	msg := Message{
		Key:   "key1",
		Topic: "test-topic",
		Payload: map[string]string{
			"foo": "bar",
		},
	}
	if msg.Key != "key1" {
		t.Fatalf("got %q, want %q", msg.Key, "key1")
	}
	if msg.Topic != "test-topic" {
		t.Fatalf("got %q, want %q", msg.Topic, "test-topic")
	}
}

func TestUnmarshalPayload(t *testing.T) {
	type TestPayload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data, _ := json.Marshal(TestPayload{Name: "test", Value: 42})
	got, err := UnmarshalPayload[TestPayload](data)
	if err != nil {
		t.Fatalf("UnmarshalPayload failed: %v", err)
	}
	if got.Name != "test" || got.Value != 42 {
		t.Fatal("unmarshalled data mismatch")
	}
}

func TestUnmarshalPayload_invalid_json(t *testing.T) {
	_, err := UnmarshalPayload[struct{}]([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestNewProducer_env_vars(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	p, err := NewProducer()
	if err != nil {
		t.Fatalf("NewProducer failed: %v", err)
	}
	if p.writer == nil {
		t.Fatal("expected non-nil writer")
	}
	p.Close()
}

func TestNewConsumer_env_vars(t *testing.T) {
	c, err := NewConsumer("test-group", []string{"test-topic"})
	if err != nil {
		t.Fatalf("NewConsumer failed: %v", err)
	}
	if c.reader == nil {
		t.Fatal("expected non-nil reader")
	}
	c.Close()
}

func TestNewConsumer_multiple_topics(t *testing.T) {
	c, err := NewConsumer("test-group", []string{"topic1", "topic2"})
	if err != nil {
		t.Fatalf("NewConsumer failed: %v", err)
	}
	if c.reader == nil {
		t.Fatal("expected non-nil reader")
	}
	c.Close()
}

func TestHandler_type(t *testing.T) {
	var h Handler = func(ctx context.Context, topic string, key string, data []byte) error {
		return nil
	}
	if err := h(context.Background(), "t", "k", []byte("d")); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
}
