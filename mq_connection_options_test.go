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
	"github.com/ibm-messaging/mq-golang/v5/ibmmq"

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

func TestMQConnectionOptions(t *testing.T) {

	t.Run("MaxMsgLength set on CreateContext", func(t *testing.T) {
		// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
		cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
		assert.Nil(t, cfErr)

		// Ensure that the options were applied when setting connection options on Context creation
		msg := "options were not applied"
		context, ctxErr := cf.CreateContext(
			jms20subset.WithMaxMsgLength(2000),
			func(cno *ibmmq.MQCNO) {
				assert.Equal(t, int32(2000), cno.ClientConn.MaxMsgLength)
				msg = "options applied"
			},
		)
		assert.Nil(t, ctxErr)

		if context != nil {
			defer context.Close()
		}

		assert.Equal(t, "options applied", msg)
	})

	t.Run("MaxMsgLength is respected when receive message on CreateContextWithSessionMode", func(t *testing.T) {
		// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
		cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
		assert.Nil(t, cfErr)

		sContext, ctxErr := cf.CreateContext()
		assert.Nil(t, ctxErr)
		if sContext != nil {
			defer sContext.Close()
		}

		// Create a Queue object that points at an IBM MQ queue
		sQueue := sContext.CreateQueue("DEV.QUEUE.1")
		// Send a message to the queue that contains a large string
		errSend := sContext.CreateProducer().SendString(sQueue, "This has more than data than the max message length")
		assert.NoError(t, errSend)

		// create consumer with low msg length
		rContext, ctxErr := cf.CreateContextWithSessionMode(
			jms20subset.JMSContextAUTOACKNOWLEDGE,
			jms20subset.WithMaxMsgLength(20),
		)
		assert.Nil(t, ctxErr)
		if rContext != nil {
			defer rContext.Close()
		}

		// Create a Queue object that points at an IBM MQ queue
		rQueue := rContext.CreateQueue("DEV.QUEUE.1")
		// Send a message to the queue that contains a large string
		consumer, errCons := rContext.CreateConsumer(rQueue)
		assert.NoError(t, errCons)

		// expect that receiving the message will cause an JMS Data Length error
		_, err := consumer.ReceiveStringBodyNoWait()
		assert.Error(t, err)
		jmsErr, ok := err.(jms20subset.JMSExceptionImpl)
		assert.True(t, ok)
		assert.Equal(t, "MQRC_DATA_LENGTH_ERROR", jmsErr.GetReason())

		// Now consume the message sucessfully to tidy up before the next test.
		tidyConsumer, errCons := sContext.CreateConsumer(sQueue)
		assert.NoError(t, errCons)
		gotMsg, err := tidyConsumer.ReceiveNoWait()
		assert.NoError(t, err)
		assert.NotNil(t, gotMsg)
	})
}
