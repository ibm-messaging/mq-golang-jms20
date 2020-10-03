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

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the typical request/reply messaging pattern where the message ID of the
 * request message is used as the correlation ID of the response.
 */
func TestRequestReply(t *testing.T) {

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

	// First, check the Send queue is initially empty
	requestQueue := context.CreateQueue("DEV.QUEUE.1")
	requestConsumer, rConErr := context.CreateConsumer(requestQueue)
	assert.Nil(t, rConErr)
	if requestConsumer != nil {
		defer requestConsumer.Close()
	}
	reqMsgTest, err2 := requestConsumer.ReceiveNoWait()
	assert.Nil(t, err2)
	assert.Nil(t, reqMsgTest)
	requestConsumer.Close()

	// Check the Reply queue is initially empty, then close the consumer.
	replyQueue := context.CreateQueue("DEV.QUEUE.2")
	replyConsumer, rConErr2 := context.CreateConsumer(replyQueue)
	assert.Nil(t, rConErr2)
	replyMsgTest, err2 := replyConsumer.ReceiveNoWait()
	replyConsumer.Close()
	assert.Nil(t, err2)
	assert.Nil(t, replyMsgTest)

	// Create a message and send it to the Send queue
	msgBody := "RequestMsg"
	sentMsg := context.CreateTextMessageWithString(msgBody)
	assert.Equal(t, msgBody, *sentMsg.GetText())

	// Set the replyTo queue
	assert.Nil(t, sentMsg.GetJMSReplyTo())
	sentMsg.SetJMSReplyTo(replyQueue)
	assert.NotNil(t, sentMsg.GetJMSReplyTo())

	err := context.CreateProducer().Send(requestQueue, sentMsg)
	assert.Nil(t, err)

	// Get the generated MessageID from the sent message, so that we
	// can receive the response message using it later
	msgID := sentMsg.GetJMSMessageID()
	assert.NotEqual(t, "", msgID)
	assert.Equal(t, 48, len(msgID))

	// "Another application" will consume the request message and sent
	// the reply message
	replyToMessage(t, cf, requestQueue)

	// Receive the reply message, selecting by CorrelID
	replyConsumer, rConErr3 := context.CreateConsumerWithSelector(replyQueue, "JMSCorrelationID = '"+msgID+"'")
	assert.Nil(t, rConErr3)
	if replyConsumer != nil {
		defer replyConsumer.Close()
	}
	respMsg, err2 := replyConsumer.ReceiveNoWait()
	assert.Nil(t, err2)
	assert.NotNil(t, respMsg)

	// We have a "Message" object and can use a switch to safely test it is our expected TextMessage type
	switch msg := respMsg.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, "ReplyMsg", *msg.GetText())
		assert.Equal(t, msgID, msg.GetJMSCorrelationID())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

}

/*
 * Simulate another application replying to a message
 */
func replyToMessage(t *testing.T, cf jms20subset.ConnectionFactory, requestQueue jms20subset.Destination) {

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	rrContext, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if rrContext != nil {
		defer rrContext.Close()
	}

	// Receive the request message.
	requestConsumer, rConErr := rrContext.CreateConsumer(requestQueue)
	assert.Nil(t, rConErr)
	if requestConsumer != nil {
		defer requestConsumer.Close()
	}
	reqMsg, err := requestConsumer.ReceiveNoWait()
	assert.Nil(t, err)
	reqMsgID := reqMsg.GetJMSMessageID()
	assert.NotEqual(t, "", reqMsgID)

	// Get the replyTo queue and request MsgID
	replyDest := reqMsg.GetJMSReplyTo()
	assert.NotNil(t, replyDest)
	assert.Equal(t, "DEV.QUEUE.2", replyDest.GetDestinationName())

	replyMsgBody := "ReplyMsg"
	replyMsg := rrContext.CreateTextMessageWithString(replyMsgBody)
	assert.Equal(t, "", replyMsg.GetJMSCorrelationID())
	replyMsg.SetJMSCorrelationID(reqMsgID)
	assert.NotEqual(t, "", replyMsg.GetJMSCorrelationID())

	// Send the reply message back to the original application
	err2 := rrContext.CreateProducer().Send(replyDest, replyMsg)
	assert.Nil(t, err2)

}
