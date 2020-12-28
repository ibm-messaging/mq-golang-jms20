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

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Demonstrate the ability to interrogate error codes when failing to connect
 * to a queue manager.
 */
func TestFailToConnect(t *testing.T) {

	// Create a ConnectionFactory using some property files
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Set a value we know will cause a failure
	cf.UserName = "wrong_user"

	// Creates a connection to the queue manager
	context, err := cf.CreateContext()
	assert.NotNil(t, err)
	if context != nil {
		defer context.Close()
	}

	// Check the error code that comes back in the Error (JMS uses a String for the code)
	assert.Equal(t, "2035", err.GetErrorCode())
	assert.Equal(t, "MQRC_NOT_AUTHORIZED", err.GetReason())

}

/*
 * Demonstrate the ability to interrogate error codes when failing to open a
 * queue on a successfully connected queue manager.
 */
func TestFailToConnectToQueue(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function
	context, err := cf.CreateContext()
	assert.Nil(t, err)
	if context != nil {
		defer context.Close()
	}

	// Deliberately target a queue that does not exist.
	queue := context.CreateQueue("DOES.NOT.EXIST.QUEUE")

	// Send a message!
	err2 := context.CreateProducer().SendString(queue, "My first message")

	assert.NotNil(t, err2)

	// Check the error code that comes back in the Error (JMS uses a String for code)
	assert.Equal(t, "2085", err2.GetErrorCode())
	assert.Equal(t, "MQRC_UNKNOWN_OBJECT_NAME", err2.GetReason())

	_, err3 := context.CreateConsumer(queue)

	assert.NotNil(t, err3)
	assert.Equal(t, "2085", err3.GetErrorCode())
	assert.Equal(t, "MQRC_UNKNOWN_OBJECT_NAME", err3.GetReason())

}
