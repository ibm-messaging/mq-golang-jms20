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
 * Test the behaviour of sending a message under a transaction.
 *
 * - put without transaction, immediately available
 * - put multiple msgs with transaction, not immediately available
 * - available after commit
 * - put multiple msgs under transaction then rollback - not available
 * - put message under transaction, close connection - should not be available
 */
func TestPutTransaction(t *testing.T) {

	// Create a ConnectionFactory using some property files
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager.
	untransactedContext, errCtx := cf.CreateContext()
	assert.Nil(t, errCtx)
	if untransactedContext != nil {
		defer untransactedContext.Close()
	}

	transactedContext, errCtx := cf.CreateContextWithSessionMode(jms20subset.JMSContextSESSIONTRANSACTED)
	assert.Nil(t, errCtx)
	if transactedContext != nil {
		defer transactedContext.Close()
	}

	// Create queue objects that points at an IBM MQ queue
	queueName := "DEV.QUEUE.1"
	unQueue := untransactedContext.CreateQueue(queueName)
	trQueue := transactedContext.CreateQueue(queueName)

	// Create an untransacted consumer
	untransactedConsumer, errCons := untransactedContext.CreateConsumer(unQueue)
	assert.Nil(t, errCons)
	if untransactedConsumer != nil {
		defer untransactedConsumer.Close()
	}

	// Send an untransacted message and check it is immediately available
	untransactedProducer := untransactedContext.CreateProducer().SetTimeToLive(20000)
	bodyTxt := "untransacted-put"
	errSend := untransactedProducer.SendString(unQueue, bodyTxt)
	assert.Nil(t, errSend)
	rcvBody, errRcv := untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt, *rcvBody)

	// put multiple msgs with transaction, not immediately available
	transactedProducer := transactedContext.CreateProducer().SetTimeToLive(20000)
	bodyTxt1 := "transacted-put-1"
	bodyTxt2 := "transacted-put-2"
	errSend = transactedProducer.SendString(trQueue, bodyTxt1)
	assert.Nil(t, errSend)
	errSend = transactedProducer.SendString(trQueue, bodyTxt2)
	assert.Nil(t, errSend)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody) // Expect nil here - message should not have been received.

	// Commit and messages should now be available (in order)
	cxErr := transactedContext.Commit()
	assert.Nil(t, cxErr)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt1, *rcvBody)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt2, *rcvBody)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody) // Only expected two messages

	// put multiple msgs under transaction then rollback - not available
	bodyTxt1 = "transacted-put-rollback-1"
	bodyTxt2 = "transacted-put-rollback-2"
	errSend = transactedProducer.SendString(trQueue, bodyTxt1)
	assert.Nil(t, errSend)
	errSend = transactedProducer.SendString(trQueue, bodyTxt2)
	assert.Nil(t, errSend)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)
	rbErr := transactedContext.Rollback() // Undo the messages
	assert.Nil(t, rbErr)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)
	cxErr = transactedContext.Commit() // Should no longer be under the transaction
	assert.Nil(t, cxErr)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)

	// put message under transaction, close connection - should not be available
	errSend = transactedProducer.SendString(trQueue, "orphan1")
	assert.Nil(t, errSend)
	errSend = transactedProducer.SendString(trQueue, "orphan2")
	assert.Nil(t, errSend)
	transactedContext.Close()
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)

}

/**
 * Test the behaviour of receiving a message under a transaction.
 *
 * - get without transaction, immediately disappears
 * - get multiple with transaction, immediately disappears
 * - rollback, multiple reappear
 * - get with transaction, commit, not available
 * - receive under transaction, close connection - should be available again (rollback)
 */
func TestGetTransaction(t *testing.T) {

	// Create a ConnectionFactory using some property files
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager.
	untransactedContext, errCtx := cf.CreateContext()
	assert.Nil(t, errCtx)
	if untransactedContext != nil {
		defer untransactedContext.Close()
	}

	transactedContext, errCtx := cf.CreateContextWithSessionMode(jms20subset.JMSContextSESSIONTRANSACTED)
	assert.Nil(t, errCtx)
	if transactedContext != nil {
		defer transactedContext.Close()
	}

	// Create queue objects that points at an IBM MQ queue
	queueName := "DEV.QUEUE.1"
	unQueue := untransactedContext.CreateQueue(queueName)
	trQueue := transactedContext.CreateQueue(queueName)

	// Create an transacted consumer
	transactedConsumer, errCons := transactedContext.CreateConsumer(trQueue)
	assert.Nil(t, errCons)
	if transactedConsumer != nil {
		defer transactedConsumer.Close()
	}

	// Create an untransacted consumer
	untransactedConsumer, errCons := untransactedContext.CreateConsumer(unQueue)
	assert.Nil(t, errCons)
	if untransactedConsumer != nil {
		defer untransactedConsumer.Close()
	}

	// Send an untransacted message and check it is immediately available for untransacted receive
	untransactedProducer := untransactedContext.CreateProducer().SetTimeToLive(20000)
	bodyTxt := "untransacted-get"
	errSend := untransactedProducer.SendString(unQueue, bodyTxt)
	assert.Nil(t, errSend)
	rcvBody, errRcv := untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt, *rcvBody)
	rcvBody, errRcv = transactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody) // Has been consumed by the untransacted consumer

	// get multiple with transaction, immediately disappears
	bodyTxt1 := "transacted-get-1"
	bodyTxt2 := "transacted-get-2"
	errSend = untransactedProducer.SendString(unQueue, bodyTxt1)
	assert.Nil(t, errSend)
	errSend = untransactedProducer.SendString(unQueue, bodyTxt2)
	assert.Nil(t, errSend)

	rcvBody, errRcv = transactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt1, *rcvBody)
	rcvBody, errRcv = transactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt2, *rcvBody)

	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody) // Message is not available (consumed pending transaction)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)

	// rollback, messages reappear
	rbErr := transactedContext.Rollback() // puts the two messages back on the queue
	assert.Nil(t, rbErr)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt1, *rcvBody)

	// get the second reappeared message with transaction, commit, not available
	rcvBody, errRcv = transactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt2, *rcvBody)
	cxErr := transactedContext.Commit() // commit the consumption of the one message.
	assert.Nil(t, cxErr)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody) // No message should be available

	// receive under transaction, close connection - should be available again (rollback)
	bodyTxt3 := "transacted-get-3"
	errSend = untransactedProducer.SendString(unQueue, bodyTxt3)
	assert.Nil(t, errSend)
	rcvBody, errRcv = transactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt3, *rcvBody)
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)    // message not available
	transactedContext.Close() // causes rollback
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, bodyTxt3, *rcvBody) // message now available

}

/**
 * Test the behaviour of receiving and sending a message under the same local transaction.
 *
 * - place initial message on queue
 * - get message under transaction, put reply message to different queue
 * - neither request nor reply message is available
 * - rollback; request message is available, reply message is not
 * - (again) get message under transaction, put reply message to different queue
 * - commit; request message is gone, and reply message is available
 */
func TestPutGetTransaction(t *testing.T) {

	// Create a ConnectionFactory using some property files
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager.
	untransactedContext, errCtx := cf.CreateContext()
	assert.Nil(t, errCtx)
	if untransactedContext != nil {
		defer untransactedContext.Close()
	}

	transactedContext, errCtx := cf.CreateContextWithSessionMode(jms20subset.JMSContextSESSIONTRANSACTED)
	assert.Nil(t, errCtx)
	if transactedContext != nil {
		defer transactedContext.Close()
	}

	senderTransactedContext, errCtx := cf.CreateContextWithSessionMode(jms20subset.JMSContextSESSIONTRANSACTED)
	assert.Nil(t, errCtx)
	if senderTransactedContext != nil {
		defer senderTransactedContext.Close()
	}

	// Create Request + Reply queue objects that points at IBM MQ queues
	reqQueueName := "DEV.QUEUE.1"
	unReqQueue := untransactedContext.CreateQueue(reqQueueName)
	trReqQueue := transactedContext.CreateQueue(reqQueueName)

	replyQueueName := "DEV.QUEUE.2"
	unReplyQueue := untransactedContext.CreateQueue(replyQueueName)
	trReplyQueue := transactedContext.CreateQueue(replyQueueName)

	// Create an unrelated transacted producer (different connection)
	transactedReqProducer := senderTransactedContext.CreateProducer().SetTimeToLive(20000)

	// Create an transacted consumer for the request queue
	transactedReqConsumer, errCons := transactedContext.CreateConsumer(trReqQueue)
	assert.Nil(t, errCons)
	if transactedReqConsumer != nil {
		defer transactedReqConsumer.Close()
	}

	// Create an untransacted consumer for the request queue
	untransactedReqConsumer, errCons := untransactedContext.CreateConsumer(unReqQueue)
	assert.Nil(t, errCons)
	if untransactedReqConsumer != nil {
		defer untransactedReqConsumer.Close()
	}

	// Create a transacted producer for the reply
	transactedReplyProducer := transactedContext.CreateProducer().SetTimeToLive(20000)

	// Create an untransacted consumer for the reply queue
	untransactedReplyConsumer, errCons := untransactedContext.CreateConsumer(unReplyQueue)
	assert.Nil(t, errCons)
	if untransactedReplyConsumer != nil {
		defer untransactedReplyConsumer.Close()
	}

	// First check that both queues are empty
	rcvBody, errRcv := untransactedReqConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)
	rcvBody, errRcv = untransactedReplyConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)

	// Use the transacted sender context to send a request message (under a transaction)
	msgBody := "putget-transaction"
	errSend := transactedReqProducer.SendString(trReqQueue, msgBody)
	assert.Nil(t, errSend)
	rcvBody, errRcv = untransactedReqConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)                    // Not yet visible for consumption
	cxErr := senderTransactedContext.Commit() // Make the message visible
	assert.Nil(t, cxErr)

	// get message under transaction, put reply message to different queue
	rcvBody, errRcv = transactedReqConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, msgBody, *rcvBody)
	replyMsgBody := "putget-transaction-reply"
	errSend = transactedReplyProducer.SendString(trReplyQueue, replyMsgBody)
	assert.Nil(t, errSend)

	// neither request nor reply message is available
	rcvBody, errRcv = untransactedReqConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)
	rcvBody, errRcv = untransactedReplyConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)

	// rollback; request message is available, reply message is not
	rbErr := transactedContext.Rollback()
	assert.Nil(t, rbErr)
	rcvBody, errRcv = untransactedReqConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, msgBody, *rcvBody)
	rcvBody, errRcv = untransactedReplyConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)

	// Put a new request message
	msgBody2 := "putget-transaction-2"
	errSend = transactedReqProducer.SendString(trReqQueue, msgBody2)
	assert.Nil(t, errSend)
	cxErr = senderTransactedContext.Commit()
	assert.Nil(t, cxErr)

	// (again) get message under transaction, put reply message to the other queue
	rcvBody, errRcv = transactedReqConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, msgBody2, *rcvBody)
	replyMsgBody2 := "putget-transaction-reply-2"
	errSend = transactedReplyProducer.SendString(trReplyQueue, replyMsgBody2)
	assert.Nil(t, errSend)

	// commit; request message is gone, and reply message is available
	cxErr = transactedContext.Commit()
	assert.Nil(t, cxErr)
	rcvBody, errRcv = untransactedReqConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)
	rcvBody, errRcv = untransactedReplyConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Equal(t, replyMsgBody2, *rcvBody)

}
