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

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the time to live behaviour of a message - it should be deleted
 * automatically by the queue manager after its TTL expires.
 */
func TestMsgTimeToLive(t *testing.T) {

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

	// Check the getter/setter behaviour on a producer.
	tmpProducer := context.CreateProducer()
	assert.Equal(t, 0, tmpProducer.GetTimeToLive())
	tmpProducer.SetTimeToLive(2000)
	assert.Equal(t, 2000, tmpProducer.GetTimeToLive())

	// Send a message that will expire after 1 second, then immediately receive it (will not have expired)
	msgBody := "Get me before I expire!"
	context.CreateProducer().SetTimeToLive(1000).SendString(queue, msgBody)
	time.Sleep(500 * time.Millisecond) // Go to sleep for 0.5s so the message is still present
	rcvBody, rcvErr := consumer.ReceiveStringBodyNoWait()
	assert.Nil(t, rcvErr)
	assert.NotNil(t, rcvBody)
	assert.Equal(t, msgBody, *rcvBody)

	// Send a second message that will expire after 1 second, but wait for 1.5s before trying to consume it.
	// (it should not be received!)
	msgBody2 := "Catch me if you can!"
	context.CreateProducer().SetTimeToLive(1000).SendString(queue, msgBody2)
	time.Sleep(1500 * time.Millisecond) // Go to sleep for 1.5s to allow the message to expire
	rcvMsg2, err2 := consumer.ReceiveNoWait()
	assert.Nil(t, rcvMsg2)
	assert.Nil(t, err2)

	// Print the received message body for debug purposes if the test fails
	if rcvMsg2 != nil {
		switch msg := rcvMsg2.(type) {
		case jms20subset.TextMessage:
			assert.Equal(t, "NO_MESSAGE", msg.GetText())
		default:
			assert.Fail(t, "Got something other than a text message")
		}

	}

}
