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
	"fmt"
	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
 * Test the ability to send both Persistent and Non-Persistent messages.
 */
func TestDeliveryMode(t *testing.T) {

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

	// Equivalent to a JNDI lookup or other declarative definition
	queue := context.CreateQueue("DEV.QUEUE.1")

	// First, check the queue is empty
	consumer, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	if consumer != nil {
		defer consumer.Close()
	}

	expectedNilMsg, rcvErr := consumer.ReceiveStringBodyNoWait()
	assert.Nil(t, rcvErr)
	assert.Nil(t, expectedNilMsg)
	if expectedNilMsg != nil {
		fmt.Println("Unexpected message with body: " + *expectedNilMsg)
	}

	// Check the default behaviour if you don't set anything, which is that
	// messages are send as Persistent.
	tmpProducer := context.CreateProducer()
	assert.Equal(t, jms20subset.DeliveryMode_PERSISTENT, tmpProducer.GetDeliveryMode())
	msgBody := "DeliveryMode test default"
	errDef := tmpProducer.SendString(queue, msgBody)
	assert.Nil(t, errDef)
	defMsg, errDef2 := consumer.ReceiveNoWait()
	assert.Nil(t, errDef2)
	assert.NotNil(t, defMsg)
	assert.Equal(t, jms20subset.DeliveryMode_PERSISTENT, defMsg.GetJMSDeliveryMode())

	// We have a "Message" object and can use a switch to safely test it is our expected TextMessage
	switch msg := defMsg.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, msgBody, *msg.GetText())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

	// Check the getter/setter behaviour on the Producer
	tmpProducer = context.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_PERSISTENT)
	assert.Equal(t, jms20subset.DeliveryMode_PERSISTENT, tmpProducer.GetDeliveryMode())
	tmpProducer.SetDeliveryMode(jms20subset.DeliveryMode_NON_PERSISTENT)
	assert.Equal(t, jms20subset.DeliveryMode_NON_PERSISTENT, tmpProducer.GetDeliveryMode())

	// Create a Producer and set the delivery mode to Persistent
	msgBody = "DeliveryMode test P"
	err := context.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_PERSISTENT).SendString(queue, msgBody)
	assert.Nil(t, err)

	pMsg, err2 := consumer.ReceiveNoWait()
	assert.Nil(t, err2)
	assert.NotNil(t, pMsg)

	assert.Equal(t, jms20subset.DeliveryMode_PERSISTENT, pMsg.GetJMSDeliveryMode())

	// Create a Producer and set the delivery mode to NonPersistent
	msgBody = "DeliveryMode test NP"
	err3 := context.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_NON_PERSISTENT).SendString(queue, msgBody)
	assert.Nil(t, err3)

	npMsg, err4 := consumer.ReceiveNoWait()
	assert.Nil(t, err4)
	assert.NotNil(t, npMsg)

	assert.Equal(t, jms20subset.DeliveryMode_NON_PERSISTENT, npMsg.GetJMSDeliveryMode())

	// We have a "Message" object and can use a switch to safely test it is our expected TextMessage
	switch msg := npMsg.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, msgBody, *msg.GetText())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

}
