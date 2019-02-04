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
	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

/*
 * Test that the timestamp allocated to the message represents the time at which
 * test message is accepted by the queue manager (i.e during the Put)
 */
func TestJMSTimestamp(t *testing.T) {

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

	// To cope with the fact that the system clock on the queue manager instance
	// probably isn't perfectly in sync with the clock on the machine running the
	// test we execute the testcase twice slightly apart, and check that the delta
	// of the timing window is consistent in each case.

	// First test
	startDeltaOne, endDeltaOne := sendForTimestamp(t, context)

	time.Sleep(250 * time.Millisecond)

	// Second test
	startDeltaTwo, endDeltaTwo := sendForTimestamp(t, context)

	// The start deltas should be basically identical.
	if startDeltaTwo-startDeltaOne > 50 {
		assert.Fail(t, "Start deltas differ by more than 50ms")
	}

	// Likewise the end deltas
	if endDeltaTwo-endDeltaOne > 50 {
		assert.Fail(t, "End deltas differ by more than 50ms")
	}

}

/*
 * Test scenario that sends a message, gets it back again and checks the Timestamp
 * of the received message is during the Send call.
 *
 * Returns the difference between sendStartTime, messageTimestamp and endStartTime
 * in milliseconds
 */
func sendForTimestamp(t *testing.T, context jms20subset.JMSContext) (startDelta, endDelta int64) {

	// Equivalent to a JNDI lookup or other declarative definition
	queue := context.CreateQueue("DEV.QUEUE.1")

	// Create a message
	msgBody := "My message for timestamp"
	txtMsg := context.CreateTextMessageWithString(msgBody)
	time.Sleep(200 * time.Millisecond)

	// Send it, and take timestamps as close as possible before and after the
	// method call, so that we can check later whether the allocated timestamp
	// is in the correct interval during the Send call.
	beforeSendTime := time.Now()
	context.CreateProducer().Send(queue, txtMsg)
	afterSendTime := time.Now()

	time.Sleep(200 * time.Millisecond)

	// Receive the message
	consumer, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	if consumer != nil {
		defer consumer.Close()
	}

	rcvMsg, err := consumer.ReceiveNoWait()
	assert.NotNil(t, rcvMsg)
	assert.Nil(t, err)

	msgTimestamp := rcvMsg.GetJMSTimestamp()
	assert.NotEqual(t, int64(0), msgTimestamp)
	msgTimestampNanos := msgTimestamp * 1000000

	beforeNanos := beforeSendTime.UnixNano()
	afterNanos := afterSendTime.UnixNano()

	// Test the timestamp is within our expected bounds
	if beforeNanos < msgTimestampNanos && msgTimestampNanos < afterNanos {
		// This is what we expect
		startDelta = 0
		endDelta = 0
	} else {
		// Try to mitigate the clocks being out of sync
		startDelta = (msgTimestampNanos - beforeSendTime.UnixNano()) / 1000000
		endDelta = (afterSendTime.UnixNano() - msgTimestampNanos) / 1000000
	}

	return startDelta, endDelta

}
