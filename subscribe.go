// Copyright 2019 The mqtt-go authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mqtt

import (
	"context"
	"errors"
)

// ErrInvalidSubAck means that the incomming SUBACK packet is inconsistent with the request.
var ErrInvalidSubAck = errors.New("invalid SUBACK")

type subscribeFlag byte

const (
	subscribeFlagQoS0 subscribeFlag = 0x00
	subscribeFlagQoS1 subscribeFlag = 0x01
	subscribeFlagQoS2 subscribeFlag = 0x02
)

type pktSubscribe struct {
	ID            uint16
	Subscriptions []Subscription
}

func (p *pktSubscribe) Pack() []byte {
	payload := make([]byte, 0, packetBufferCap)
	for _, sub := range p.Subscriptions {
		payload = appendString(payload, sub.Topic)

		var flag byte
		switch sub.QoS {
		case QoS0:
			flag |= byte(subscribeFlagQoS0)
		case QoS1:
			flag |= byte(subscribeFlagQoS1)
		case QoS2:
			flag |= byte(subscribeFlagQoS2)
		default:
			panic("invalid QoS")
		}
		payload = append(payload, flag)
	}
	return pack(
		byte(packetSubscribe|packetFromClient),
		packUint16(p.ID),
		payload,
	)
}

// Subscribe topics.
func (c *BaseClient) Subscribe(ctx context.Context, subs ...Subscription) error {
	c.muConnecting.RLock()
	defer c.muConnecting.RUnlock()

	id := c.newID()

	chSubAck := make(chan *pktSubAck, 1)
	c.sig.mu.Lock()
	if c.sig.chSubAck == nil {
		c.sig.chSubAck = make(map[uint16]chan *pktSubAck)
	}
	c.sig.chSubAck[id] = chSubAck
	c.sig.mu.Unlock()

	pkt := (&pktSubscribe{ID: id, Subscriptions: subs}).Pack()
	if err := c.write(pkt); err != nil {
		return wrapError(err, "sending SUBSCRIBE")
	}
	select {
	case <-c.connClosed:
		return ErrClosedTransport
	case <-ctx.Done():
		return ctx.Err()
	case subAck := <-chSubAck:
		if len(subAck.Codes) != len(subs) {
			c.Transport.Close()
			return wrapErrorf(ErrInvalidSubAck, "subscribing %d topics: %d topics in SUBACK", len(subs), len(subAck.Codes))
		}
		for i := 0; i < len(subAck.Codes); i++ {
			subs[i].QoS = QoS(subAck.Codes[i])
		}
	}
	return nil
}
