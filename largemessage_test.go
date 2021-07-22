/*
 * Copyright (c) IBM Corporation 2019
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0, which is available at
 * http://www.eclipse.org/legal/epl-2.0.
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package main

import (
	"strconv"
	"testing"
	"time"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the send/receive of a large Text message, over the default 32KB size.
 */
func TestLargeTextMessage(t *testing.T) {

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

	// Get a long text string over 32kb in length
	txtOver32kb := getStringOver32kb()

	// Create a TextMessage with it
	msg := context.CreateTextMessageWithString(txtOver32kb)

	// Now send the message and get it back again.
	queue := context.CreateQueue("DEV.QUEUE.1")
	errSend := context.CreateProducer().SetTimeToLive(30000).Send(queue, msg)
	assert.Nil(t, errSend)

	consumer, errCons := context.CreateConsumer(queue)
	assert.Nil(t, errCons)
	if consumer != nil {
		defer consumer.Close()
	}

	// This will fail because the default buffer size when receiving a
	// message is 32kb.
	_, errRcv := consumer.ReceiveNoWait()
	assert.NotNil(t, errRcv)
	assert.Equal(t, "MQRC_TRUNCATED_MSG_FAILED", errRcv.GetReason())
	assert.Equal(t, "2080", errRcv.GetErrorCode())

	// The message is still left on the queue since it failed to be received successfully.

	// Since the buffer size is configured using the ConnectionFactoy we will
	// create a second connection in order to successfully retrieve the message.
	cf.ReceiveBufferSize = len(txtOver32kb) + 50

	context2, ctxErr2 := cf.CreateContext()
	assert.Nil(t, ctxErr2)
	if context2 != nil {
		defer context2.Close()
	}

	consumer2, errCons2 := context2.CreateConsumer(queue)
	assert.Nil(t, errCons2)
	if consumer2 != nil {
		defer consumer2.Close()
	}

	rcvMsg2, errRcv2 := consumer2.ReceiveNoWait()
	assert.Nil(t, errRcv2)
	assert.NotNil(t, rcvMsg2)

	switch msg2 := rcvMsg2.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, &txtOver32kb, msg2.GetText())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

}

/*
 * Test the send/receive of a large Text message, over the default 32KB size.
 */
func TestLargeReceiveStringBodyTextMessage(t *testing.T) {

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

	// Get a long text string over 32kb in length
	txtOver32kb := getStringOver32kb()

	// Create a TextMessage with it
	msg := context.CreateTextMessageWithString(txtOver32kb)

	// Now send the message and get it back again.
	queue := context.CreateQueue("DEV.QUEUE.1")
	errSend := context.CreateProducer().SetTimeToLive(30000).Send(queue, msg)
	assert.Nil(t, errSend)

	consumer, errCons := context.CreateConsumer(queue)
	assert.Nil(t, errCons)
	if consumer != nil {
		defer consumer.Close()
	}

	// This will fail because the default buffer size when receiving a
	// message is 32kb.
	_, errRcv := consumer.ReceiveStringBodyNoWait()
	assert.NotNil(t, errRcv)
	assert.Equal(t, "MQRC_TRUNCATED_MSG_FAILED", errRcv.GetReason())
	assert.Equal(t, "2080", errRcv.GetErrorCode())

	// The message is still left on the queue since it failed to be received successfully.

	// Since the buffer size is configured using the ConnectionFactoy we will
	// create a second connection in order to successfully retrieve the message.
	cf.ReceiveBufferSize = len(txtOver32kb) + 50

	context2, ctxErr2 := cf.CreateContext()
	assert.Nil(t, ctxErr2)
	if context2 != nil {
		defer context2.Close()
	}

	consumer2, errCons2 := context2.CreateConsumer(queue)
	assert.Nil(t, errCons2)
	if consumer2 != nil {
		defer consumer2.Close()
	}

	rcvStr, errRcv2 := consumer2.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv2)
	assert.Equal(t, &txtOver32kb, rcvStr)

}

/*
 * Test the send/receive of a large Bytes message, over the default 32KB size.
 */
func TestLargeBytesMessage(t *testing.T) {

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

	// Get a long text string over 32kb in length
	txtOver32kb := getStringOver32kb()
	bytesOver32kb := []byte(txtOver32kb)

	// Create a TextMessage with it
	msg := context.CreateBytesMessageWithBytes(bytesOver32kb)

	// Now send the message and get it back again.
	queue := context.CreateQueue("DEV.QUEUE.1")
	errSend := context.CreateProducer().SetTimeToLive(30000).Send(queue, msg)
	assert.Nil(t, errSend)

	consumer, errCons := context.CreateConsumer(queue)
	assert.Nil(t, errCons)
	if consumer != nil {
		defer consumer.Close()
	}

	// This will fail because the default buffer size when receiving a
	// message is 32kb.
	_, errRcv := consumer.ReceiveNoWait()
	assert.NotNil(t, errRcv)
	assert.Equal(t, "MQRC_TRUNCATED_MSG_FAILED", errRcv.GetReason())
	assert.Equal(t, "2080", errRcv.GetErrorCode())

	// The message is still left on the queue since it failed to be received successfully.

	// Since the buffer size is configured using the ConnectionFactoy we will
	// create a second connection in order to successfully retrieve the message.
	cf.ReceiveBufferSize = len(bytesOver32kb) + 50

	context2, ctxErr2 := cf.CreateContext()
	assert.Nil(t, ctxErr2)
	if context2 != nil {
		defer context2.Close()
	}

	consumer2, errCons2 := context2.CreateConsumer(queue)
	assert.Nil(t, errCons2)
	if consumer2 != nil {
		defer consumer2.Close()
	}

	rcvMsg2, errRcv2 := consumer2.ReceiveNoWait()
	assert.Nil(t, errRcv2)
	assert.NotNil(t, rcvMsg2)

	switch msg2 := rcvMsg2.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, &bytesOver32kb, msg2.ReadBytes())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

}

/*
 * Test the send/receive of a large Bytes message, over the default 32KB size.
 */
func TestLargeReceiveBytesBodyBytesMessage(t *testing.T) {

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

	// Get a long text string over 32kb in length
	txtOver32kb := getStringOver32kb()
	bytesOver32kb := []byte(txtOver32kb)

	// Create a TextMessage with it
	msg := context.CreateBytesMessageWithBytes(bytesOver32kb)

	// Now send the message and get it back again.
	queue := context.CreateQueue("DEV.QUEUE.1")
	errSend := context.CreateProducer().SetTimeToLive(30000).Send(queue, msg)
	assert.Nil(t, errSend)

	consumer, errCons := context.CreateConsumer(queue)
	assert.Nil(t, errCons)
	if consumer != nil {
		defer consumer.Close()
	}

	// This will fail because the default buffer size when receiving a
	// message is 32kb.
	_, errRcv := consumer.ReceiveBytesBodyNoWait()
	assert.NotNil(t, errRcv)
	assert.Equal(t, "MQRC_TRUNCATED_MSG_FAILED", errRcv.GetReason())
	assert.Equal(t, "2080", errRcv.GetErrorCode())

	// The message is still left on the queue since it failed to be received successfully.

	// Since the buffer size is configured using the ConnectionFactoy we will
	// create a second connection in order to successfully retrieve the message.
	cf.ReceiveBufferSize = len(bytesOver32kb) + 50

	context2, ctxErr2 := cf.CreateContext()
	assert.Nil(t, ctxErr2)
	if context2 != nil {
		defer context2.Close()
	}

	consumer2, errCons2 := context2.CreateConsumer(queue)
	assert.Nil(t, errCons2)
	if consumer2 != nil {
		defer consumer2.Close()
	}

	rcvBytes2, errRcv2 := consumer2.ReceiveBytesBodyNoWait()
	assert.Nil(t, errRcv2)
	assert.Equal(t, &bytesOver32kb, rcvBytes2)

}

func getStringOver32kb() string {

	// Build a text string which is over 32KB (in a not very efficient way!)
	// First snippet is 100 chars long.
	txt100chars := "The quick brown fox jumped over the lazy dog into the flowing river that rushed over the cliff edge."
	txt1300chars := strconv.FormatInt(time.Now().UnixNano(), 10) + txt100chars + txt100chars + txt100chars + txt100chars + txt100chars +
		txt100chars + txt100chars + txt100chars + txt100chars + txt100chars +
		txt100chars + txt100chars + txt100chars

	txtOver32kb := txt1300chars

	for len(txtOver32kb) < 32*1024 {
		txtOver32kb += txt1300chars
	}

	return txtOver32kb

}
