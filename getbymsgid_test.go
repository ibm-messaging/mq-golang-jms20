/*
 * Copyright (c) IBM Corporation 2022
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
	"reflect"
	"testing"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test receiving a specific message from a queue using its JMSMessageID
 */
func TestGetByMsgID(t *testing.T) {

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

	// First, check the queue is empty
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	if consumer != nil {
		defer consumer.Close()
	}
	reqMsgTest, err := consumer.ReceiveNoWait()
	assert.Nil(t, err)
	assert.Nil(t, reqMsgTest)

	// Put a couple of messages before the one we're aiming to get back
	producer := context.CreateProducer().SetTimeToLive(10000)
	producer.SendString(queue, "One")
	producer.SendString(queue, "Two")

	myMsgThreeStr := "Three"
	sentMsg := context.CreateTextMessageWithString(myMsgThreeStr)
	err = producer.Send(queue, sentMsg)
	assert.Nil(t, err)
	sentMsgID := sentMsg.GetJMSMessageID()

	// Put a couple of messages after the one we're aiming to get back
	producer.SendString(queue, "Four")
	producer.SendString(queue, "Five")

	// Create the consumer to read by CorrelID
	correlIDConsumer, correlErr := context.CreateConsumerWithSelector(queue, "JMSMessageID = '"+sentMsgID+"'")
	assert.Nil(t, correlErr)
	gotCorrelMsg, correlGetErr := correlIDConsumer.ReceiveNoWait()

	// Clean up the remaining messages from the queue before we start checking if
	// we got the right one back.
	var cleanupMsg jms20subset.Message
	for ok := true; ok; ok = (cleanupMsg != nil) {
		cleanupMsg, err = consumer.ReceiveNoWait()
	}

	// Now do the comparisons
	assert.Nil(t, correlGetErr)
	assert.NotNil(t, gotCorrelMsg)
	gotMsgID := gotCorrelMsg.GetJMSMessageID()
	assert.Equal(t, sentMsgID, gotMsgID)

	switch msg := gotCorrelMsg.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, myMsgThreeStr, *msg.GetText())
	default:
		fmt.Println(reflect.TypeOf(gotCorrelMsg))
		assert.Fail(t, "Got something other than a text message")
	}

}
