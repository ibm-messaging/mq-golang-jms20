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
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Minimal example showing how to send a message asynchronously, which can give a higher
 * rate of sending non-persistent messages, in exchange for less/no checking for errors.
 */
func TestAsyncPutSample(t *testing.T) {

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

	// Set up a Producer for NonPersistent messages and Destination the PutAsyncAllowed=true
	producer := context.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_NON_PERSISTENT)
	asyncQueue := context.CreateQueue("DEV.QUEUE.1").SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED)

	// Send a message (asynchronously)
	msg := context.CreateTextMessageWithString("some text")
	errSend := producer.Send(asyncQueue, msg)
	assert.Nil(t, errSend)

	// Tidy up the message to leave the test clean.
	consumer, errCons := context.CreateConsumer(asyncQueue)
	assert.Nil(t, errCons)
	if consumer != nil {
		defer consumer.Close()
	}
	_, errRvc := consumer.ReceiveStringBodyNoWait()
	assert.Nil(t, errRvc)

}

/*
 * Compare the performance benefit of sending messages non-persistent, non-transactional
 * messages asynchronously - which can give a higher message rate, in exchange for
 * less/no checking for errors.
 *
 * The test checks that async put is at least 10% faster than synchronous put.
 * (in testing against a remote queue manager it was actually 30% faster)
 */
func TestAsyncPutComparison(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Check the default value for SendCheckCount, which means never check for errors.
	assert.Equal(t, 0, cf.SendCheckCount)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	// --------------------------------------------------------
	// Start by sending a set of messages using the normal synchronous approach, in
	// order that we can get a baseline timing.

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
	// WITHOUT async put (i.e. using normal synchronous put)
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

	// Receive the messages back again to tidy the queue back to a clean state
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

	// Receive the messages back again to tidy the queue back to a clean state
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
 * Test the ability to successfully send async messages with checking enabled.
 *
 * This test is checking that no failures are reported when the interval checking
 * is enabled.
 */
func TestAsyncPutCheckCount(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Set the CF flag to enable checking for errors after a certain number of messages
	cf.SendCheckCount = 10

	// Check the default value for SendCheckCount
	assert.Equal(t, 10, cf.SendCheckCount)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	// Set up the producer and consumer with the async queue.
	asyncQueue := context.CreateQueue("DEV.QUEUE.1").SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED)
	producer := context.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_NON_PERSISTENT)

	// Create a unique message prefix representing this execution of the test case.
	testcasePrefix := strconv.FormatInt(currentTimeMillis(), 10)
	msgPrefix := "checkCount_" + testcasePrefix + "_"
	numberMessages := 50

	// --------------------------------------------------------
	// Do ASYNC message put
	for i := 0; i < numberMessages; i++ {

		// Create a TextMessage and send it.
		msg := context.CreateTextMessageWithString(msgPrefix + strconv.Itoa(i))

		errSend := producer.Send(asyncQueue, msg)
		assert.Nil(t, errSend)
	}

	// ----------------------------------
	// Receive the messages back again to tidy the queue back to a clean state
	consumer, errCons := context.CreateConsumer(asyncQueue)
	assert.Nil(t, errCons)
	if consumer != nil {
		defer consumer.Close()
	}

	finishedReceiving := false

	for !finishedReceiving {
		rcvMsg, errRvc := consumer.ReceiveNoWait()
		assert.Nil(t, errRvc)

		if rcvMsg == nil {
			finishedReceiving = true
		}
	}
}

/*
 * Validate that errors are reported at the correct interval when a problem occurs.
 *
 * This test case forces a failure to occur by sending 50 messages to a queue that has had its
 * maximum depth set to 25. With SendCheckCount of 10 we will not receive an error until message 30,
 * which is the first time the error check is made after the point at which the queue has filled up.
 */
func TestAsyncPutCheckCountWithFailure(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Set the CF flag to enable checking for errors after a certain number of messages
	cf.SendCheckCount = 10

	// Check the value for SendCheckCount was stored correctly.
	assert.Equal(t, 10, cf.SendCheckCount)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	// Set up the producer and consumer with the async queue.
	QUEUE_25_NAME := "DEV.MAXDEPTH25"
	asyncQueue := context.CreateQueue(QUEUE_25_NAME).SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED)
	producer := context.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_NON_PERSISTENT)

	// Create a unique message prefix representing this execution of the test case.
	testcasePrefix := strconv.FormatInt(currentTimeMillis(), 10)
	msgPrefix := "checkCount_" + testcasePrefix + "_"
	numberMessages := 50

	// Variable to track whether the queue exists or not.
	queueExists := true

	// --------------------------------------------------------
	// Send ASYNC message put
	for i := 0; i < numberMessages; i++ {

		// Create a TextMessage and send it.
		msg := context.CreateTextMessageWithString(msgPrefix + strconv.Itoa(i))

		errSend := producer.Send(asyncQueue, msg)

		// Messages will start to fail at number 25 but we don't get an error until
		// the next check which takes place at 30.
		if i == 0 && isUnknownObjectName(errSend) {

			fmt.Println("Skipping TestAsyncPutCheckCountWithFailure as queue " + QUEUE_25_NAME + " is not defined.")
			queueExists = false
			break // Stop the loop at this point as we know it won't change.

		} else if i < 30 {
			assert.Nil(t, errSend)
		} else if i == 30 {

			assert.NotNil(t, errSend)
			assert.Equal(t, "AsyncPutFailure", errSend.GetErrorCode())

			// Message should be "N failures"
			assert.True(t, strings.Contains(errSend.GetReason(), "6 failures"))
			assert.True(t, strings.Contains(errSend.GetReason(), "0 warnings"))

			// Linked message should have reason of MQRC_Q_FULL
			linkedErr := errSend.GetLinkedError()
			assert.NotNil(t, linkedErr)
			linkedReason := linkedErr.(jms20subset.JMSExceptionImpl).GetReason()
			assert.Equal(t, "MQRC_Q_FULL", linkedReason)

		} else if i == 40 {

			assert.NotNil(t, errSend)
			assert.Equal(t, "AsyncPutFailure", errSend.GetErrorCode())

			// Message should be "N failures"
			assert.True(t, strings.Contains(errSend.GetReason(), "10 failures")) // all of these failed
			assert.True(t, strings.Contains(errSend.GetReason(), "0 warnings"))

			// Linked message should have reason of MQRC_Q_FULL
			linkedErr := errSend.GetLinkedError()
			assert.NotNil(t, linkedErr)
			linkedReason := linkedErr.(jms20subset.JMSExceptionImpl).GetReason()
			assert.Equal(t, "MQRC_Q_FULL", linkedReason)

		} else {
			// Messages 31, 32, ... 39, 41, 42, ...
			// do not give an error because we don't make an error check.
			assert.Nil(t, errSend)
		}
	}

	// If the queue exists then tidy up the messages we sent.
	if queueExists {

		// ----------------------------------
		// Receive the messages back again to tidy the queue back to a clean state
		consumer, errCons := context.CreateConsumer(asyncQueue)
		assert.Nil(t, errCons)
		if consumer != nil {
			defer consumer.Close()
		}

		// Receive the messages back again
		finishedReceiving := false

		for !finishedReceiving {
			rcvMsg, errRvc := consumer.ReceiveNoWait()
			assert.Nil(t, errRvc)

			if rcvMsg == nil {
				finishedReceiving = true
			}
		}
	}
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

/*
 * Compare the performance benefit of sending messages persistent, transactional
 * messages asynchronously - which can give a higher message rate, in exchange for
 * less/no checking for errors.
 *
 * The test checks that async put is at least 10% faster than synchronous put.
 * (in testing against a remote queue manager it was actually 30% faster)
 */
func TestAsyncPutPersistentTransactedComparison(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Check the default value for SendCheckCount, which means never check for errors.
	assert.Equal(t, 0, cf.SendCheckCount)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	transactedContext, ctxErr := cf.CreateContextWithSessionMode(jms20subset.JMSContextSESSIONTRANSACTED)
	assert.Nil(t, ctxErr)
	if transactedContext != nil {
		defer transactedContext.Close()
	}

	// --------------------------------------------------------
	// Start by sending a set of messages using the normal synchronous approach, in
	// order that we can get a baseline timing.

	// Set up the producer and consumer with the SYNCHRONOUS (not async yet) queue
	syncQueue := transactedContext.CreateQueue("DEV.QUEUE.1")
	producer := transactedContext.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_PERSISTENT).SetTimeToLive(60000)

	consumer, errCons := transactedContext.CreateConsumer(syncQueue)
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
	// WITHOUT async put (i.e. using normal synchronous put)
	syncStartTime := currentTimeMillis()
	for i := 0; i < numberMessages; i++ {

		// Create a TextMessage and send it.
		msg := transactedContext.CreateTextMessageWithString(syncMsgPrefix + strconv.Itoa(i))

		errSend := producer.Send(syncQueue, msg)
		assert.Nil(t, errSend)
	}
	transactedContext.Commit()
	syncEndTime := currentTimeMillis()

	syncSendTime := syncEndTime - syncStartTime
	//fmt.Println("Took " + strconv.FormatInt(syncSendTime, 10) + "ms to send " + strconv.Itoa(numberMessages) + " transacted synchronous messages.")

	// Receive the messages back again to tidy the queue back to a clean state
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
	transactedContext.Commit()

	// --------------------------------------------------------
	// Now repeat the experiment but with ASYNC message put
	asyncQueue := transactedContext.CreateQueue("DEV.QUEUE.1").SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED)

	asyncStartTime := currentTimeMillis()
	for i := 0; i < numberMessages; i++ {

		// Create a TextMessage and send it.
		msg := transactedContext.CreateTextMessageWithString(asyncMsgPrefix + strconv.Itoa(i))

		errSend := producer.Send(asyncQueue, msg)
		assert.Nil(t, errSend)
	}
	transactedContext.Commit()
	asyncEndTime := currentTimeMillis()

	asyncSendTime := asyncEndTime - asyncStartTime
	//fmt.Println("Took " + strconv.FormatInt(asyncSendTime, 10) + "ms to send " + strconv.Itoa(numberMessages) + " transacted ASYNC messages.")

	// Receive the messages back again to tidy the queue back to a clean state
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
	transactedContext.Commit()

	assert.Equal(t, numberMessages, rcvCount)

	// Expect that async put is at least 10% faster than sync put for non-persistent messages
	// (in testing against a remote queue manager it was actually 30% faster)
	assert.True(t, 100*asyncSendTime < 90*syncSendTime)

}

/*
 * Validate that errors are reported at the correct interval when a problem occurs during a
 * transactional put of persistent messages (i.e. when the commit occurs)
 *
 * This test case forces a failure to occur by sending 50 messages to a queue that has had its
 * maximum depth set to 25. In the transacted async put case the Send method always completes
 * successfully, but the error will be returned on Commit.
 */
func TestAsyncPutTransactedCheckCountWithFailure(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// The SendCheckCount is not used for async put under a transacted session.
	assert.Equal(t, 0, cf.SendCheckCount)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	transactedContext, ctxErr := cf.CreateContextWithSessionMode(jms20subset.JMSContextSESSIONTRANSACTED)
	assert.Nil(t, ctxErr)
	if transactedContext != nil {
		defer transactedContext.Close()
	}

	// Set up the producer and consumer with the async queue.
	QUEUE_25_NAME := "DEV.MAXDEPTH25"
	asyncQueue := transactedContext.CreateQueue(QUEUE_25_NAME).SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED)
	producer := transactedContext.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_PERSISTENT).SetTimeToLive(60000)

	// Create a unique message prefix representing this execution of the test case.
	testcasePrefix := strconv.FormatInt(currentTimeMillis(), 10)
	msgPrefix := "checkCount_" + testcasePrefix + "_"
	numberMessages := 50

	// Variable to track whether the queue exists or not.
	queueExists := true

	// --------------------------------------------------------
	// Send ASYNC message put
	for i := 0; i < numberMessages; i++ {

		// Create a TextMessage and send it.
		msg := transactedContext.CreateTextMessageWithString(msgPrefix + strconv.Itoa(i))

		errSend2 := producer.Send(asyncQueue, msg)

		// In the Transacted case the response from Send is always Nil, because any errors
		// will be reflected on the Commit call.
		assert.Nil(t, errSend2)

		// Skip the test if the destination does not exist on this queue manager.
		if i == 0 {
			if err := transactedContext.Commit(); isUnknownObjectName(err) {
				fmt.Println("Skipping TestAsyncPutTransactedCheckCountWithFailure as queue " + QUEUE_25_NAME + " is not defined.")
				queueExists = false
				break // Stop the loop at this point as we know it won't change.
			}
		}

		if i%10 == 0 {
			commitErr := transactedContext.Commit()

			if i == 30 || i == 40 {

				assert.NotNil(t, commitErr)
				assert.Equal(t, "2003", commitErr.GetErrorCode())
				assert.Equal(t, "MQRC_BACKED_OUT", commitErr.GetReason())

				// Linked error is out normal async put error with message about "N failures"
				linkedErr1 := commitErr.GetLinkedError()
				assert.Equal(t, "AsyncPutFailure", linkedErr1.(jms20subset.JMSExceptionImpl).GetErrorCode())
				assert.True(t, strings.Contains(linkedErr1.(jms20subset.JMSExceptionImpl).GetReason(), "6 failures"))
				assert.True(t, strings.Contains(linkedErr1.(jms20subset.JMSExceptionImpl).GetReason(), "0 warnings"))

				// Second level linked message should have reason of MQRC_Q_FULL
				linkedErr2 := linkedErr1.(jms20subset.JMSExceptionImpl).GetLinkedError()
				assert.NotNil(t, linkedErr2)
				linkedReason := linkedErr2.(jms20subset.JMSExceptionImpl).GetReason()
				assert.Equal(t, "MQRC_Q_FULL", linkedReason)

			} else {
				// Messages 31, 32, ... 39, 41, 42, ...
				// do not give an error because we don't make an error check.
				assert.Nil(t, commitErr)
			}
		}
	}

	// Clear out the transaction context.
	transactedContext.Commit()

	// If the queue exists then tidy up the messages we sent.
	if queueExists {

		// ----------------------------------
		// Receive the messages back again to tidy the queue back to a clean state
		consumer, errCons := transactedContext.CreateConsumer(asyncQueue)
		assert.Nil(t, errCons)
		if consumer != nil {
			defer consumer.Close()
		}

		// Receive the messages back again
		finishedReceiving := false

		for !finishedReceiving {
			rcvMsg, errRvc := consumer.ReceiveNoWait()
			assert.Nil(t, errRvc)

			if rcvMsg == nil {
				finishedReceiving = true
			}
		}

		transactedContext.Commit()
	}
}

/*
 * Validate that NO errors are reported when a problem occurs during a
 * transactional put of a non-persistent message.
 *
 * (per https://www.ibm.com/docs/en/ibm-mq/9.1?topic=application-putting-messages-asynchronously-in-mq-classes-jms)
 *
 * This test case forces a failure to occur by sending 50 messages to a queue that has had its
 * maximum depth set to 25.
 */
func TestAsyncPutTransactedNonPersistentCheckCountWithFailure(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// The SendCheckCount is not used for async put under a transacted session.
	assert.Equal(t, 0, cf.SendCheckCount)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	transactedContext, ctxErr := cf.CreateContextWithSessionMode(jms20subset.JMSContextSESSIONTRANSACTED)
	assert.Nil(t, ctxErr)
	if transactedContext != nil {
		defer transactedContext.Close()
	}

	// Set up the producer and consumer with the async queue.
	QUEUE_25_NAME := "DEV.MAXDEPTH25"
	asyncQueue := transactedContext.CreateQueue(QUEUE_25_NAME).SetPutAsyncAllowed(jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED)
	producer := transactedContext.CreateProducer().SetDeliveryMode(jms20subset.DeliveryMode_NON_PERSISTENT).SetTimeToLive(60000)

	// Create a unique message prefix representing this execution of the test case.
	testcasePrefix := strconv.FormatInt(currentTimeMillis(), 10)
	msgPrefix := "checkCount_" + testcasePrefix + "_"
	numberMessages := 50

	// Variable to track whether the queue exists or not.
	queueExists := true

	// --------------------------------------------------------
	// Send ASYNC message put
	for i := 0; i < numberMessages; i++ {

		// Create a TextMessage and send it.
		msg := transactedContext.CreateTextMessageWithString(msgPrefix + strconv.Itoa(i))

		errSend2 := producer.Send(asyncQueue, msg)

		// In the Transacted case the response from Send is always Nil, because any errors
		// will be reflected on the Commit call.
		assert.Nil(t, errSend2)

		if i%10 == 0 {
			commitErr := transactedContext.Commit()

			// No errors reported for NonPersistent messages in a transaction.
			assert.Nil(t, commitErr)

		}
	}

	// ----------------------------------
	// Receive the messages back again to tidy the queue back to a clean state
	consumer, errCons := transactedContext.CreateConsumer(asyncQueue)

	// Skip the test if the destination does not exist on this queue manager.
	if isUnknownObjectName(errCons) {
		fmt.Println("Skipping TestAsyncPutTransactedNonPersistentCheckCountWithFailure as queue " + QUEUE_25_NAME + " is not defined.")
		queueExists = false
	}

	// If the queue exists then tidy up the messages we sent.
	if queueExists {

		assert.Nil(t, errCons)
		if consumer != nil {
			defer consumer.Close()
		}

		// Receive the messages back again
		finishedReceiving := false

		for !finishedReceiving {
			rcvMsg, errRvc := consumer.ReceiveNoWait()
			assert.Nil(t, errRvc)

			if rcvMsg == nil {
				finishedReceiving = true
			}
		}

		transactedContext.Commit()
	}
}

func isUnknownObjectName(exception jms20subset.JMSException) bool {
	if exception != nil {
		if exception.GetReason() == "MQRC_UNKNOWN_OBJECT_NAME" {
			return true
		}
		if err, ok := exception.GetLinkedError().(jms20subset.JMSExceptionImpl); ok {
			return err.GetReason() == "MQRC_UNKNOWN_OBJECT_NAME" || isUnknownObjectName(err)
		}
	}
	return false
}
