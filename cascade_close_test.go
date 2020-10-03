/*
 * Copyright (c) IBM Corporation 2020
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

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the behaviour of the cascading close.
 *
 * - Create three consumers
 * - Receive with timeout, get no message
 * - Close second consumer, try to receive again, get error, check 1 + 3 still ok
 * - Ctx.close, try to receive on the other two, get error
 */
func TestCascadeClose(t *testing.T) {

	// Loads CF parameters from connection_info.json and apiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	// We are testing Close behaviour here, but auto-cleanup just in case.
	if context != nil {
		defer context.Close()
	}

	// Equivalent to a JNDI lookup or other declarative definition
	queue := context.CreateQueue("DEV.QUEUE.1")

	// Set up the consumers 1-3 ready to receive messages.
	consumer1, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	// We are testing Close behaviour here, but auto-cleanup just in case.
	if consumer1 != nil {
		defer consumer1.Close()
	}

	consumer2, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	// We are testing Close behaviour here, but auto-cleanup just in case.
	if consumer2 != nil {
		defer consumer2.Close()
	}

	consumer3, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	// We are testing Close behaviour here, but auto-cleanup just in case.
	if consumer3 != nil {
		defer consumer3.Close()
	}

	// Check no message on the queue to start with
	testMsg, err := consumer1.ReceiveNoWait()
	assert.Nil(t, err)
	assert.Nil(t, testMsg)

	testMsg, err = consumer2.ReceiveNoWait()
	assert.Nil(t, err)
	assert.Nil(t, testMsg)

	testMsg, err = consumer3.ReceiveNoWait()
	assert.Nil(t, err)
	assert.Nil(t, testMsg)

	// Close second consumer, try to receive again, get error, check 1 + 3 still ok
	consumer2.Close()

	testMsg, err = consumer1.ReceiveNoWait()
	assert.Nil(t, err)
	assert.Nil(t, testMsg)

	testMsg, err = consumer2.ReceiveNoWait()
	assert.NotNil(t, err)
	assert.Equal(t, "2019", err.GetErrorCode())
	assert.Equal(t, "MQRC_HOBJ_ERROR", err.GetReason())
	assert.Nil(t, testMsg)

	testMsg, err = consumer3.ReceiveNoWait()
	assert.Nil(t, err)
	assert.Nil(t, testMsg)

	// Close a closed thing to check it doesn't complain.
	consumer2.Close()

	// Ctx.close, try to receive on the other two, get error
	context.Close()

	testMsg, err = consumer1.ReceiveNoWait()
	assert.NotNil(t, err)
	assert.Equal(t, "2018", err.GetErrorCode())
	assert.Equal(t, "MQRC_HCONN_ERROR", err.GetReason())
	assert.Nil(t, testMsg)

	testMsg, err = consumer2.ReceiveNoWait()
	assert.NotNil(t, err)
	assert.Equal(t, "2018", err.GetErrorCode())
	assert.Equal(t, "MQRC_HCONN_ERROR", err.GetReason())
	assert.Nil(t, testMsg)

	testMsg, err = consumer3.ReceiveNoWait()
	assert.NotNil(t, err)
	assert.Equal(t, "2018", err.GetErrorCode())
	assert.Equal(t, "MQRC_HCONN_ERROR", err.GetReason())
	assert.Nil(t, testMsg)

}
