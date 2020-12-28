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
	"testing"
	"time"

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the behaviour of the receive-with-wait methods.
 */
func TestReceiveWait(t *testing.T) {

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

	// Equivalent to a JNDI lookup or other declarative definition
	queue := context.CreateQueue("DEV.QUEUE.1")

	// Set up the consumer ready to receive messages.
	consumer, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	if consumer != nil {
		defer consumer.Close()
	}

	// Check no message on the queue to start with
	testMsg, err1 := consumer.ReceiveNoWait()
	assert.Nil(t, err1)
	assert.Nil(t, testMsg)

	waitTime := int32(500)

	// Since there is no message on the queue, the receiveWait call should
	// wait for the requested period of time, before returning nil.
	startTime := currentTimeMillis()
	testMsg2, err2 := consumer.Receive(waitTime)
	endTime := currentTimeMillis()
	assert.Nil(t, testMsg2)
	assert.Nil(t, err2)

	// Within a reasonable margin of the expected wait time.
	assert.True(t, (endTime-startTime-int64(waitTime)) > -100)

	// Send a message.
	err := context.CreateProducer().SendString(queue, "MyMessage")
	assert.Nil(t, err)

	// With a message on the queue the receiveWait call should return immediately
	// with the message.
	startTime2 := currentTimeMillis()
	testMsg3, err3 := consumer.Receive(waitTime)
	endTime2 := currentTimeMillis()
	assert.NotNil(t, testMsg3)
	assert.Nil(t, err3)

	// Should not wait the entire waitTime duration
	assert.True(t, (endTime2-startTime2) < 300)

}

func TestReceiveStringBodyWait(t *testing.T) {

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

	// Equivalent to a JNDI lookup or other declarative definition
	queue := context.CreateQueue("DEV.QUEUE.1")

	// Set up the consumer ready to receive messages.
	consumer, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	if consumer != nil {
		defer consumer.Close()
	}

	// Check no message on the queue to start with
	msg, err1 := consumer.ReceiveStringBodyNoWait()
	assert.Nil(t, err1)
	assert.Nil(t, msg)

	waitTime := int32(500)

	// Since there is no message on the queue, the receiveWait call should
	// wait for the requested period of time, before returning nil.
	startTime := currentTimeMillis()
	msg2, err2 := consumer.ReceiveStringBody(waitTime)
	endTime := currentTimeMillis()
	assert.Nil(t, msg2)
	assert.Nil(t, err2)

	// Within a reasonable margin of the expected wait time.
	assert.True(t, (endTime-startTime-int64(waitTime)) > -100)

	// Send a message.
	inputMsg := "MyMessageTest"
	err := context.CreateProducer().SendString(queue, inputMsg)
	assert.Nil(t, err)

	// With a message on the queue the receiveWait call should return immediately
	// with the message.
	startTime2 := currentTimeMillis()
	msg3, err3 := consumer.ReceiveStringBody(waitTime)
	endTime2 := currentTimeMillis()
	assert.Nil(t, err3)
	assert.Equal(t, inputMsg, *msg3)

	// Should not wait the entire waitTime duration
	assert.True(t, (endTime2-startTime2) < 300)

}

// Return the current time in milliseconds
func currentTimeMillis() int64 {
	return int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond)
}
