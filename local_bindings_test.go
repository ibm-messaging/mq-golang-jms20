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
	"fmt"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

/*
 * Test the use of a local bindings connection to the queue manager, as
 * opposed to a remote client connection.
 *
 * Assumes a locally running queue manager called QM1, with a defined local
 * queue called LOCAL.QUEUE
 */
func TestLocalBindingsConnect(t *testing.T) {

	// Checking the presence of the mqs.ini file is a reasonably reliable way of
	// identifying if there is a local queue manager installation, so that we
	// can skip this test if necessary.
	if _, err := os.Stat("/var/mqm/mqs.ini"); os.IsNotExist(err) {
		fmt.Println("Skipping local bindings test as there is no local queue manager installation.")
		return
	}

	// Create a connection factory that indicates we should use local bindings.
	// Note that it includes none of the normal parameters like hostname, port etc.
	cf := mqjms.ConnectionFactoryImpl{
		QMName:        "QM1",
		TransportType: mqjms.TransportType_BINDINGS,
	}

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}
	assert.NotNil(t, context)

	// Equivalent to a JNDI lookup or other declarative definition
	queue := context.CreateQueue("LOCAL.QUEUE")

	// Set up the consumer ready to receive messages.
	consumer, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	if consumer != nil {
		defer consumer.Close()
	}

	// Check no message on the queue to start with
	testMsg, err1 := consumer.ReceiveNoWait()
	assert.Nil(t, err1)
	assert.Nil(t, testMsg)

	// Check the getter/setter behaviour on a producer.
	tmpProducer := context.CreateProducer()
	tmpProducer.SetTimeToLive(2000)

	// Send a message that will expire after 1 second, then immediately receive it (will not have expired)
	msgBody := "Local message."
	context.CreateProducer().SendString(queue, msgBody)
	rcvBody, rcvErr := consumer.ReceiveStringBodyNoWait()
	assert.Nil(t, rcvErr)
	assert.NotNil(t, rcvBody)
	assert.Equal(t, msgBody, *rcvBody)

}
