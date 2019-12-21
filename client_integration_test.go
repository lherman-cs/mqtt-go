// +build integration

package mqtt

import (
	"context"
	"testing"
)

func TestIntegration_Connect(t *testing.T) {
	cli := &Client{}
	if err := cli.Dial("mqtt://localhost:1883"); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
	go cli.Serve()

	ctx := context.Background()
	err := cli.Connect(ctx, "Client1")
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if err := cli.Disconnect(ctx); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
}

func TestIntegration_PublishQoS0(t *testing.T) {
	cli := &Client{}
	if err := cli.Dial("mqtt://localhost:1883"); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
	go cli.Serve()

	ctx := context.Background()
	err := cli.Connect(ctx, "Client1")
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	err = cli.Publish(ctx, &Message{
		Topic:   "test",
		Payload: []byte("message"),
	})
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if err := cli.Disconnect(ctx); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
}

func TestIntegration_PublishQoS1(t *testing.T) {
	cli := &Client{}
	if err := cli.Dial("mqtt://localhost:1883"); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
	go cli.Serve()

	ctx := context.Background()
	err := cli.Connect(ctx, "Client1")
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	err = cli.Publish(ctx, &Message{
		Topic:   "test",
		QoS:     QoS1,
		Payload: []byte("message"),
	})
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if err := cli.Disconnect(ctx); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
}

func TestIntegration_PublishQoS2(t *testing.T) {
	cli := &Client{}
	if err := cli.Dial("mqtt://localhost:1883"); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
	go cli.Serve()

	ctx := context.Background()
	err := cli.Connect(ctx, "Client1")
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	err = cli.Subscribe(ctx, []*Message{
		{
			Topic: "test",
			QoS:   QoS1,
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	err = cli.Publish(ctx, &Message{
		Topic:   "test",
		QoS:     QoS2,
		Payload: []byte("message"),
	})
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if err := cli.Disconnect(ctx); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
}
