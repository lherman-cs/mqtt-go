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

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/at-wat/mqtt-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("   usage: %s server-host.domain\n", os.Args[0])
		fmt.Printf("requires: certificate.crt, private.key, root-CA.crt\n")
		os.Exit(1)
	}
	host := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tlsConfig, err := newTLSConfig(
		host,
		"root-CA.crt",
		"certificate.crt",
		"private.key",
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	println("Connecting to", host)

	cli, err := mqtt.NewReconnectClient(ctx,
		&mqtt.URLDialer{
			URL: fmt.Sprintf("mqtts://%s:8883", host),
			Options: []mqtt.DialOption{
				mqtt.WithTLSConfig(tlsConfig),
			},
		},
		"sample",
		mqtt.WithKeepAlive(30),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	println("Connected")

	cli.Handle(mqtt.HandlerFunc(func(msg *mqtt.Message) {
		fmt.Printf("Received on %s: %s (QoS: %d)\n", msg.Topic, []byte(msg.Payload), int(msg.QoS))
		cancel()
	}))

	if err := cli.Subscribe(ctx, mqtt.Subscription{
		Topic: "test",
		QoS:   mqtt.QoS1,
	}); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	println("Publishing one message to 'test' topic")

	if err := cli.Publish(ctx, &mqtt.Message{
		Topic:   "test",
		QoS:     mqtt.QoS1,
		Payload: []byte("{\"message\": \"Hello\"}"),
	}); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	println("Waiting message on 'test' topic")
	<-ctx.Done()

	println("Disconnecting")

	if err := cli.Disconnect(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	println("Disconnected")
}

func newTLSConfig(host, caFile, crtFile, keyFile string) (*tls.Config, error) {
	certpool := x509.NewCertPool()
	cas, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	certpool.AppendCertsFromPEM(cas)

	cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		ServerName:   host,
		RootCAs:      certpool,
		Certificates: []tls.Certificate{cert},
	}, nil
}
