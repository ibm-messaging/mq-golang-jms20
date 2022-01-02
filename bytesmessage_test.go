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

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
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

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
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

	rcvMsg, errRcv := consumer.ReceiveNoWait()
	assert.Nil(t, errRcv)
	assert.NotNil(t, rcvMsg)

	switch msg2 := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, 0, msg2.GetBodyLength())
		assert.Equal(t, []byte{}, *msg2.ReadBytes())
	default:
		assert.Fail(t, "Got something other than a bytes message")
	}

}

/*
 * Test send and receive of a bytes message with some basic content.
 */
func TestBytesMessageWithBody(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	// Create a BytesMessage
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

	rcvMsg, errRcv := consumer.ReceiveNoWait()
	assert.Nil(t, errRcv)
	assert.NotNil(t, rcvMsg)

	switch msg2 := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, len(msgBody), msg2.GetBodyLength())
		assert.Equal(t, msgBody, *msg2.ReadBytes())
	default:
		assert.Fail(t, "Got something other than a bytes message")
	}

}

/*
 * Test send and receive of a bytes message with init'd at create time.
 */
func TestBytesMessageInitWithBytes(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
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
	msgBody := []byte{'b', 'y', 't', 'e', 's', 'm', 'e', 's', 's', 'a', 'g', 'e', 'A', 'B', 'C'}
	msg := context.CreateBytesMessageWithBytes(msgBody)
	assert.Equal(t, 15, msg.GetBodyLength())
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

	rcvMsg, errRcv := consumer.Receive(2000) // also try receive with wait
	assert.Nil(t, errRcv)
	assert.NotNil(t, rcvMsg)

	switch msg2 := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, len(msgBody), msg2.GetBodyLength())
		assert.Equal(t, msgBody, *msg2.ReadBytes())
	default:
		assert.Fail(t, "Got something other than a bytes message")
	}

}

/*
 * Test Producer SendBytes shortcut
 */
func TestBytesMessageProducerSendBytes(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
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
	msgBody := []byte{'b', 'y', 't', 'e', 's', 'p', 'r', 'o', 'd', 'u', 'c', 'e', 'r'}

	// Now send the message and get it back again, to check that it roundtripped.
	queue := context.CreateQueue("DEV.QUEUE.1")
	errSend := context.CreateProducer().SetTimeToLive(5000).SendBytes(queue, msgBody)
	assert.Nil(t, errSend)

	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	rcvMsg, errRcv := consumer.ReceiveNoWait()
	assert.Nil(t, errRcv)
	assert.NotNil(t, rcvMsg)

	switch msg2 := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, 13, msg2.GetBodyLength())
		assert.Equal(t, msgBody, *msg2.ReadBytes())
	default:
		assert.Fail(t, "Got something other than a bytes message")
	}

}

/*
 * Test Consumer ReceiveBytesBodyNoWait shortcut
 */
func TestBytesMessageConsumerReceiveBytesBodyNoWait(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Check the behaviour if there is no message to receive.
	var expectedNil *[]byte // uninitialized
	rcvBytes, errRcv := consumer.ReceiveBytesBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, expectedNil, rcvBytes)

	// Create a BytesMessage, and check it has nil content.
	msgBody := []byte{'b', 'y', 't', 'e', 's', 'n', 'o', 'w', 'a', 'i', 't'}

	// Now send the message and get it back again, to check that it roundtripped.
	errSend := context.CreateProducer().SetTimeToLive(5000).SendBytes(queue, msgBody)
	assert.Nil(t, errSend)

	rcvBytes, errRcv = consumer.ReceiveBytesBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, msgBody, *rcvBytes)

}

/*
 * Test Consumer ReceiveBytesBody shortcut
 */
func TestBytesMessageConsumerReceiveBytesBody(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Check the behaviour if there is no message to receive.
	var expectedNil *[]byte // uninitialized
	rcvBytes, errRcv := consumer.ReceiveBytesBody(250)
	assert.Nil(t, errRcv)
	assert.Equal(t, expectedNil, rcvBytes)

	// Create a BytesMessage, and check it has nil content.
	msgBody := []byte{'b', 'y', 't', 'e', 's', 'w', 'a', 'i', 't'}

	// Now send the message and get it back again, to check that it roundtripped.
	errSend := context.CreateProducer().SetTimeToLive(5000).SendBytes(queue, msgBody)
	assert.Nil(t, errSend)

	rcvBytes, errRcv = consumer.ReceiveBytesBody(500)
	assert.Nil(t, errRcv)
	assert.Equal(t, msgBody, *rcvBytes)

}

/*
 * Test Consumer ReceiveBytesBody/ReceiveStringBody shortcuts when unexpected
 * types of messages are received.
 */
func TestBytesMessageConsumerMixedMessageErrors(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Send a BytesMessage, try to receive as text with no wait.
	msgBodyBytes := []byte{'b', 'y', 't', 'e', 's', '1', '2', '3', '4'}
	errSend := context.CreateProducer().SetTimeToLive(5000).SendBytes(queue, msgBodyBytes)
	assert.Nil(t, errSend)
	rcvStr, errRcv := consumer.ReceiveStringBodyNoWait()
	assert.Nil(t, rcvStr)
	assert.Equal(t, "MQJMS6068", errRcv.GetErrorCode())
	assert.Equal(t, "MQJMS_DIR_MIN_NOTTEXT", errRcv.GetReason())

	// Send a BytesMessage, try to receive as text with wait.
	errSend = context.CreateProducer().SetTimeToLive(5000).SendBytes(queue, msgBodyBytes)
	assert.Nil(t, errSend)
	rcvStr, errRcv = consumer.ReceiveStringBody(200)
	assert.Nil(t, rcvStr)
	assert.Equal(t, "MQJMS6068", errRcv.GetErrorCode())
	assert.Equal(t, "MQJMS_DIR_MIN_NOTTEXT", errRcv.GetReason())

	// Send a TextMessage, try to receive as bytes with no wait.
	msgBodyStr := "TextMessage is not Bytes"
	errSend = context.CreateProducer().SetTimeToLive(5000).SendString(queue, msgBodyStr)
	assert.Nil(t, errSend)
	rcvBytes, errRcv := consumer.ReceiveBytesBodyNoWait()
	var expectedNil *[]byte // uninitialized
	assert.Equal(t, expectedNil, rcvBytes)
	assert.Equal(t, "MQJMS6068", errRcv.GetErrorCode())
	assert.Equal(t, "MQJMS_DIR_MIN_NOTBYTES", errRcv.GetReason())

	// Send a TextMessage, try to receive as bytes with wait.
	errSend = context.CreateProducer().SetTimeToLive(5000).SendString(queue, msgBodyStr)
	assert.Nil(t, errSend)
	rcvBytes, errRcv = consumer.ReceiveBytesBody(200)
	assert.Equal(t, expectedNil, rcvBytes)
	assert.Equal(t, "MQJMS6068", errRcv.GetErrorCode())
	assert.Equal(t, "MQJMS_DIR_MIN_NOTBYTES", errRcv.GetReason())

}
