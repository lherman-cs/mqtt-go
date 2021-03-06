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
	"time"
)

// ErrKeepAliveDisabled is returned if Runned on keep alive disabled connection.
var ErrKeepAliveDisabled = errors.New("keep alive disabled")

// ErrPingTimeout is returned on ping response timeout.
var ErrPingTimeout = errors.New("ping timeout")

// KeepAlive runs keep alive loop.
// It must be called after Connect and interval must be smaller than the value
// specified by WithKeepAlive option passed to Connect.
func KeepAlive(ctx context.Context, cli ClientCloser, interval, timeout time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C

		ctxTo, cancel := context.WithTimeout(ctx, timeout)
		if err := cli.Ping(ctxTo); err != nil {
			defer cancel()
			// The client should close the connection if PINGRESP is not returned.
			// MQTT 3.1.1 spec. 3.1.2.10
			cli.Close()

			select {
			case <-ctx.Done():
				// Parent context cancelled.
				return ctx.Err()
			default:
			}
			select {
			case <-ctxTo.Done():
				return ErrPingTimeout
			default:
			}
			return err
		}
		cancel()
	}
}
