/*
 * Copyright (c) IBM Corporation 2020
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0, which is available at
 * http://www.eclipse.org/legal/epl-2.0.
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package main

import (
	"testing"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the creation of a bytes message and setting the content.
 */
func TestBytesMessageBody(t *testing.T) {

	// Loads CF parameters from connection_info.json and apiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	// Create a BytesMessage and check that we can populate it
	msgBody := []byte{'a', 'e', 'i', 'o', 'u'}
	msg := context.CreateBytesMessage()
	msg.WriteBytes(msgBody)
	assert.Equal(t, 5, msg.GetBodyLength())
	assert.Equal(t, msgBody, *msg.ReadBytes())

	// Create an empty BytesMessage and check that we query it without errors
	msg = context.CreateBytesMessage()
	assert.Equal(t, 0, msg.GetBodyLength())
	assert.Equal(t, []byte{}, *msg.ReadBytes())

}

/*
 * Test send and receive of a bytes message with no content.
 */
func TestBytesMessageNilBody(t *testing.T) {

	// Loads CF parameters from connection_info.json and apiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	// Create a BytesMessage, and check it has nil content.
	msg := context.CreateBytesMessage()
	assert.Equal(t, []byte{}, *msg.ReadBytes())

	// Now send the message and get it back again, to check that it roundtripped.
	queue := context.CreateQueue("DEV.QUEUE.1")
	errSend := context.CreateProducer().SetTimeToLive(5000).Send(queue, msg)
	assert.Nil(t, errSend)

	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg2 := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, 0, msg2.GetBodyLength())
		assert.Equal(t, []byte{}, *msg2.ReadBytes())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

}

/*
 * Test send and receive of a bytes message with no content.
 */
func TestBytesMessageWithBody(t *testing.T) {

	// Loads CF parameters from connection_info.json and apiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	// Create a BytesMessage, and check it has nil content.
	msgBody := []byte{'b', 'y', 't', 'e', 's', 'm', 'e', 's', 's', 'a', 'g', 'e'}
	msg := context.CreateBytesMessage()
	msg.WriteBytes(msgBody)
	assert.Equal(t, 12, msg.GetBodyLength())
	assert.Equal(t, msgBody, *msg.ReadBytes())

	// Now send the message and get it back again, to check that it roundtripped.
	queue := context.CreateQueue("DEV.QUEUE.1")
	errSend := context.CreateProducer().SetTimeToLive(5000).Send(queue, msg)
	assert.Nil(t, errSend)

	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg2 := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, len(msgBody), msg2.GetBodyLength())
		assert.Equal(t, msgBody, *msg2.ReadBytes())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

}
