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

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test for memory leak when there is no message to be received.
 *
 * This test is not included in the normal bucket as it sends an enormous number of
 * messages, and requires human observation of the total process size to establish whether
 * it passes or not, so can only be run under human supervision
 */
func DONT_RUNTestLeakOnEmptyGet(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	//cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	//assert.Nil(t, cfErr)

	// Initialise the attributes of the CF in whatever way you like
	cf := mqjms.ConnectionFactoryImpl{
		QMName:      "QM1",
		Hostname:    "localhost",
		PortNumber:  1414,
		ChannelName: "DEV.APP.SVRCONN",
		UserName:    "app",
		Password:    "passw0rd",
	}

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

/*
 * Test for memory leak when sending and receiving messages
 *
 * This test is not included in the normal bucket as it sends an enormous number of
 * messages, and requires human observation of the total process size to establish whether
 * it passes or not, so can only be run under human supervision
 */
func DONTRUN_TestLeakOnPutGet(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	//cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	//assert.Nil(t, cfErr)

	// Initialise the attributes of the CF in whatever way you like
	cf := mqjms.ConnectionFactoryImpl{
		QMName:      "QM1",
		Hostname:    "localhost",
		PortNumber:  1414,
		ChannelName: "DEV.APP.SVRCONN",
		UserName:    "app",
		Password:    "passw0rd",
	}

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

	ttlMillis := 20000
	producer := context.CreateProducer().SetTimeToLive(ttlMillis)

	for i := 1; i < 25000; i++ {

		// Create a TextMessage and check that we can populate it
		msgBody := "Message " + fmt.Sprint(i)
		txtMsg := context.CreateTextMessage()
		txtMsg.SetText(msgBody)
		txtMsg.SetIntProperty("MessageNumber", i)

		errSend := producer.Send(queue, txtMsg)
		assert.Nil(t, errSend)

		rcvMsg, errRvc := consumer.ReceiveNoWait()
		assert.Nil(t, errRvc)
		assert.NotNil(t, rcvMsg)

		// Check message body.
		switch msg := rcvMsg.(type) {
		case jms20subset.TextMessage:
			assert.Equal(t, msgBody, *msg.GetText())
		default:
			assert.Fail(t, "Got something other than a text message")
		}

		// Check messageID
		assert.Equal(t, txtMsg.GetJMSMessageID(), rcvMsg.GetJMSMessageID())

		// Check int property
		rcvMsgNum, propErr := rcvMsg.GetIntProperty("MessageNumber")
		assert.Nil(t, propErr)
		assert.Equal(t, i, rcvMsgNum)

		if i%1000 == 0 {
			fmt.Println("Messages:", i)
		}

	}

	fmt.Println("Finished receive calls - waiting for cooldown.")
	runtime.GC()

	time.Sleep(30 * time.Second)

}
