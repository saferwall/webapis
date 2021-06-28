// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package event

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	Conn *kafka.Conn
}

// New creates a new kafka connection to the server.
func New(network, topic, address string) (Producer, error) {
	partition := 0
	conn, err := kafka.DialLeader(context.Background(),
		network, address, topic, partition)
	if err != nil {
		return Producer{}, err
	}

	return Producer{conn}, nil
}

// Produce writes a message to the connection's topic.
func (p Producer) Produce(message []byte) error {
	p.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := p.Conn.WriteMessages(kafka.Message{Value: message})
	if err != nil {
		return err
	}
	return nil
}
