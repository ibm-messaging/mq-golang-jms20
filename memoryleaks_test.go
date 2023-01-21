/*
 * Copyright (c) IBM Corporation 2023
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
	"runtime"
	"testing"
	"time"

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test for memory leak when there is no message to be received.
 */
func DONTRUN_TestLeakOnEmptyGet(t *testing.T) {

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

	// Now send the message and get it back again, to check that it roundtripped.
	queue := context.CreateQueue("DEV.QUEUE.1")

	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	for i := 1; i < 35000; i++ {

		//time.Sleep(100 * time.Millisecond)
		rcvMsg, errRvc := consumer.ReceiveNoWait()
		assert.Nil(t, errRvc)
		assert.Nil(t, rcvMsg)

		if i%1000 == 0 {
			fmt.Println("Messages:", i)
		}

	}

	fmt.Println("Finished receive calls - waiting for cooldown.")
	runtime.GC()

	time.Sleep(30 * time.Second)

}
