// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package queue

import (
	"github.com/nsqio/go-nsq"
)

// Producer wraps the NSQ producer object.
type Producer struct {
	Producer *nsq.Producer
}

// New creates a new NSQ producer.
func New(address, topic string) (Producer, error) {

	config := nsq.NewConfig()
	p, err := nsq.NewProducer(address, config)
	if err != nil {
		return Producer{}, err
	}
	if p.Ping() != nil {
		return Producer{}, err
	}

	return Producer{p}, nil
}

// Produce writes a message to the connection's topic.
func (p Producer) Produce(topic string, message []byte) error {

	err := p.Producer.Publish(topic, message)
	if err != nil {
		return err
	}
	return nil
}
