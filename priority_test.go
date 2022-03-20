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
	"testing"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the retrieval of special header properties
 */
func TestPrioritySetGet(t *testing.T) {

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

	// Create a TextMessage and check that we can populate it
	msgBody := "PriorityMsg"
	txtMsg := context.CreateTextMessage()
	txtMsg.SetText(msgBody)

	// Set up objects for send/receive
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Now send the message and get it back again, to check that it roundtripped.
	ttlMillis := 20000
	producer := context.CreateProducer().SetTimeToLive(ttlMillis)

	// Check the default priority.
	assert.Equal(t, jms20subset.Priority_DEFAULT, producer.GetPriority())

	errSend := producer.Send(queue, txtMsg)
	assert.Nil(t, errSend)

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, msgBody, *msg.GetText())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

	// Check the Priority
	gotPropValue := rcvMsg.GetJMSPriority()
	assert.Equal(t, jms20subset.Priority_DEFAULT, gotPropValue)

	// -------
	// Go again with a different priority.
	newMsgPriority := 2
	producer = producer.SetPriority(newMsgPriority)
	assert.Equal(t, newMsgPriority, producer.GetPriority())

	errSend = producer.Send(queue, txtMsg)
	assert.Nil(t, errSend)

	rcvMsg, errRvc = consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, msgBody, *msg.GetText())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

	// Check the Priority
	gotPropValue = rcvMsg.GetJMSPriority()
	assert.Equal(t, newMsgPriority, gotPropValue)

	// -------
	// Go again with a different priority.
	newMsgPriority2 := 7
	producer = producer.SetPriority(newMsgPriority2)
	assert.Equal(t, newMsgPriority2, producer.GetPriority())

	errSend = producer.Send(queue, txtMsg)
	assert.Nil(t, errSend)

	rcvMsg, errRvc = consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, msgBody, *msg.GetText())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

	// Check the Priority
	gotPropValue = rcvMsg.GetJMSPriority()
	assert.Equal(t, newMsgPriority2, gotPropValue)

}

/*
 * Test the functional behaviour of messages sent with different priorities,
 * where the expectation is that they should be returned from the queue in priority
 * order (with FIFO within a priority group), and not FIFO across the entire queue.
 */
func TestPriorityOrdering(t *testing.T) {

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

	// Set up objects for send/receive
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Send a sequence of messages with different priorities.
	id_msg1_pri5 := sendMessageWithPriority(t, context, queue, 5)
	id_msg2_pri2 := sendMessageWithPriority(t, context, queue, 2)
	id_msg3_pri8 := sendMessageWithPriority(t, context, queue, 8)
	id_msg4_pri2 := sendMessageWithPriority(t, context, queue, 2)
	id_msg5_pri4 := sendMessageWithPriority(t, context, queue, 4)
	id_msg6_pri5 := sendMessageWithPriority(t, context, queue, 5)
	id_msg7_pri2 := sendMessageWithPriority(t, context, queue, 2)
	id_msg8_pri2 := sendMessageWithPriority(t, context, queue, 2)
	id_msg9_pri4 := sendMessageWithPriority(t, context, queue, 4)
	id_msg10_pri1 := sendMessageWithPriority(t, context, queue, 1)

	// We expect them to be returned in priority group order (highest first),
	// with FIFO within the priority group.
	//
	// This behaviour relies on MsgDeliverySequence for the queue
	// being set to MQMDS_PRIORITY (which is the default).
	receiveMessageAndCheck(t, consumer, id_msg3_pri8, 8)
	receiveMessageAndCheck(t, consumer, id_msg1_pri5, 5)
	receiveMessageAndCheck(t, consumer, id_msg6_pri5, 5)
	receiveMessageAndCheck(t, consumer, id_msg5_pri4, 4)
	receiveMessageAndCheck(t, consumer, id_msg9_pri4, 4)
	receiveMessageAndCheck(t, consumer, id_msg2_pri2, 2)
	receiveMessageAndCheck(t, consumer, id_msg4_pri2, 2)
	receiveMessageAndCheck(t, consumer, id_msg7_pri2, 2)
	receiveMessageAndCheck(t, consumer, id_msg8_pri2, 2)
	receiveMessageAndCheck(t, consumer, id_msg10_pri1, 1)

}

// Send a message with the specified priority.
// Return the generated MessageID.
func sendMessageWithPriority(t *testing.T, context jms20subset.JMSContext, queue jms20subset.Queue, priority int) string {

	ttlMillis := 20000
	producer := context.CreateProducer().SetTimeToLive(ttlMillis)
	producer = producer.SetPriority(priority)

	// Create a TextMessage and check that we can populate it
	msgBody := "PriorityOrderingMsg"
	txtMsg := context.CreateTextMessage()
	txtMsg.SetText(msgBody)

	errSend := producer.Send(queue, txtMsg)
	assert.Nil(t, errSend)

	return txtMsg.GetJMSMessageID()

}

// Try to receive a message from the queue and check that it matches
// the expected attributes
func receiveMessageAndCheck(t *testing.T, consumer jms20subset.JMSConsumer, expectedMsgID string, expectedPriority int) {

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	assert.Equal(t, expectedPriority, rcvMsg.GetJMSPriority())
	assert.Equal(t, expectedMsgID, rcvMsg.GetJMSMessageID())

}
