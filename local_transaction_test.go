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
	transactedContext.Commit()
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
	transactedContext.Rollback() // Undo the messages
	rcvBody, errRcv = untransactedConsumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRcv)
	assert.Nil(t, rcvBody)
	transactedContext.Commit() // Should no longer be under the transaction
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

// get-transaction
//   - get without transaction, immediately disappears
//   - get multiple with transaction, immedatiately disappears
//   - rollback, multiple reappear
//   - get multiple with transaction, commit, not available

// get-put transaction
//   - place initial message on queue
//   - get message under transaction, put reply message to different queue
//   - neither request nor reply message is available
//   - rollback; request message is available, reply message is not
//   - (again) get message under transaction, put reply message to different queue
//   - commit; request message is gone, and reply message is available
