package mqtt

import (
	"context"
	"io"
	"sync"
	"time"
)

type QoS uint8

const (
	QoS0             QoS = 0x00
	QoS1                 = 0x01
	QoS2                 = 0x02
	SubscribeFailure     = 0x80
)

type Message struct {
	Topic   string
	ID      uint16
	QoS     QoS
	Retain  bool
	Dup     bool
	Payload []byte
}

type Subscription struct {
	Topic string
	QoS   uint8
}

type ConnectOptions struct {
	UserName     string
	Password     string
	CleanSession bool
	KeepAlive    uint16
	Will         *Message
}

type ConnectOption func(*ConnectOptions) error

type CommandClient interface {
	Connect(ctx context.Context, clientID string, opts ...ConnectOption) error
	Disconnect(ctx context.Context) error
	Publish(ctx context.Context, message *Message) error
	Subscribe(ctx context.Context, subs ...*Subscription) error
	Unsubscribe(ctx context.Context, subs ...string) error
}

type HandlerFunc func(*Message)

func (h HandlerFunc) Serve(message *Message) {
	h(message)
}

type Handler interface {
	Serve(*Message)
}

type ConnState int

const (
	StateNew ConnState = iota
	StateActive
	StateClosed
)

type Client struct {
	Transport   io.ReadWriteCloser
	Handler     Handler
	SendTimeout time.Duration
	RecvTimeout time.Duration
	ConnState   func(ConnState)

	sig signaller
	mu  sync.RWMutex
}

type signaller struct {
	chConnAck chan *ConnAck
	chPubAck  map[uint16]chan *PubAck
	chPubRec  map[uint16]chan *PubRec
	chPubComp map[uint16]chan *PubComp
	chSubAck  map[uint16]chan *SubAck
}
