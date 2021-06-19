/*
 * Copyright (c) IBM Corporation 2021
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

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the ability to send a message asynchronously, which can give a higher
 * rate of sending non-persistent messages, in exchange for less/no checking for errors.
 */
func TestAsyncPut(t *testing.T) {

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

	// Set up the producer and consumer with the SYNCHRONOUS (not async yet) queue
	syncQueue := context.CreateQueue("DEV.QUEUE.1")
	producer := context.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_NON_PERSISTENT).SetTimeToLive(60000)

	consumer, errCons := context.CreateConsumer(syncQueue)
	assert.Nil(t, errCons)
	if consumer != nil {
		defer consumer.Close()
	}

	// Create a unique message prefix representing this execution of the test case.
	testcasePrefix := strconv.FormatInt(currentTimeMillis(), 10)
	syncMsgPrefix := "syncput_" + testcasePrefix + "_"
	asyncMsgPrefix := "asyncput_" + testcasePrefix + "_"
	numberMessages := 50

	// First get a baseline for how long it takes us to send the batch of messages
	// WITHOUT async put.
	syncStartTime := currentTimeMillis()
	for i := 0; i < numberMessages; i++ {

		// Create a TextMessage and send it.
		msg := context.CreateTextMessageWithString(syncMsgPrefix + strconv.Itoa(i))

		errSend := producer.Send(syncQueue, msg)
		assert.Nil(t, errSend)
	}
	syncEndTime := currentTimeMillis()

	syncSendTime := syncEndTime - syncStartTime
	//fmt.Println("Took " + strconv.FormatInt(syncSendTime, 10) + "ms to send " + strconv.Itoa(numberMessages) + " synchronous messages.")

	// Receive the messages back again
	finishedReceiving := false
	rcvCount := 0

	for !finishedReceiving {
		rcvTxt, errRvc := consumer.ReceiveStringBodyNoWait()
		assert.Nil(t, errRvc)

		if rcvTxt != nil {
			// Check the message bod matches what we expect
			assert.Equal(t, syncMsgPrefix+strconv.Itoa(rcvCount), *rcvTxt)
			rcvCount++
		} else {
			finishedReceiving = true
		}
	}

	// --------------------------------------------------------
	// Now repeat the experiment but with ASYNC message put
	asyncQueue := context.CreateQueue("DEV.QUEUE.1").SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED)

	asyncStartTime := currentTimeMillis()
	for i := 0; i < numberMessages; i++ {

		// Create a TextMessage and send it.
		msg := context.CreateTextMessageWithString(asyncMsgPrefix + strconv.Itoa(i))

		errSend := producer.Send(asyncQueue, msg)
		assert.Nil(t, errSend)
	}
	asyncEndTime := currentTimeMillis()

	asyncSendTime := asyncEndTime - asyncStartTime
	//fmt.Println("Took " + strconv.FormatInt(asyncSendTime, 10) + "ms to send " + strconv.Itoa(numberMessages) + " ASYNC messages.")

	// Receive the messages back again
	finishedReceiving = false
	rcvCount = 0

	for !finishedReceiving {
		rcvTxt, errRvc := consumer.ReceiveStringBodyNoWait()
		assert.Nil(t, errRvc)

		if rcvTxt != nil {
			// Check the message bod matches what we expect
			assert.Equal(t, asyncMsgPrefix+strconv.Itoa(rcvCount), *rcvTxt)
			rcvCount++
		} else {
			finishedReceiving = true
		}
	}

	assert.Equal(t, numberMessages, rcvCount)

	// Expect that async put is at least 10% faster than sync put for non-persistent messages
	// (in testing against a remote queue manager it was actually 30% faster)
	assert.True(t, 100*asyncSendTime < 90*syncSendTime)

}

/*
 * Test the getter/setter functions for controlling async put.
 */
func TestAsyncPutGetterSetter(t *testing.T) {

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

	// Set up the producer and consumer
	queue := context.CreateQueue("DEV.QUEUE.1")

	// Check the default
	assert.Equal(t, jms20subset.Destination_PUT_ASYNC_ALLOWED_AS_DEST, queue.GetPutAsyncAllowed())

	// Check enabled
	queue = queue.SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED)
	assert.Equal(t, jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED, queue.GetPutAsyncAllowed())

	// Check disabled
	queue = queue.SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_DISABLED)
	assert.Equal(t, jms20subset.Destination_PUT_ASYNC_ALLOWED_DISABLED, queue.GetPutAsyncAllowed())

	// Check as-dest
	queue = queue.SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_AS_DEST)
	assert.Equal(t, jms20subset.Destination_PUT_ASYNC_ALLOWED_AS_DEST, queue.GetPutAsyncAllowed())

}
