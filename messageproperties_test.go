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
	"testing"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * mq-golang: SetMP, DltMP, InqMP
 * https://github.com/ibm-messaging/mq-golang/blob/95e9b8b09a1fc167747de7d066c49adb86e14dda/ibmmq/mqi.go#L1080
 *
 * mq-golang sample application to set properties
 * https://github.com/ibm-messaging/mq-golang/blob/master/samples/amqsprop.go#L49
 *
 * JMS: SetStringProperty, GetStringProperty,
 * https://github.com/eclipse-ee4j/messaging/blob/master/api/src/main/java/jakarta/jms/Message.java#L1119
 *
 * JMS: PropertyExists, ClearProperties, GetPropertyNames
 */

/*
 * Test the creation of a text message with a string property.
 */
func TestStringPropertyTextMsg(t *testing.T) {

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

	// Create a TextMessage and check that we can populate it
	msgBody := "RequestMsg"
	txtMsg := context.CreateTextMessage()
	txtMsg.SetText(msgBody)

	propName := "myProperty"
	propValue := "myValue"

	// Test the empty value before the property is set.
	assert.Nil(t, txtMsg.GetStringProperty(propName))

	// Test the ability to set properties before the message is sent.
	retErr := txtMsg.SetStringProperty(propName, &propValue)
	assert.Nil(t, retErr)
	assert.Equal(t, propValue, *txtMsg.GetStringProperty(propName))
	assert.Equal(t, msgBody, *txtMsg.GetText())

	// Send an empty string property as well
	emptyPropName := "myEmptyString"
	emptyPropValue := ""
	retErr = txtMsg.SetStringProperty(emptyPropName, &emptyPropValue)
	assert.Nil(t, retErr)
	assert.Equal(t, emptyPropValue, *txtMsg.GetStringProperty(emptyPropName))

	// Set a property then try to unset it by setting to nil
	unsetPropName := "mySendThenRemovedString"
	unsetPropValue := "someValueThatWillBeOverwritten"
	retErr = txtMsg.SetStringProperty(unsetPropName, &unsetPropValue)
	assert.Nil(t, retErr)
	assert.Equal(t, unsetPropValue, *txtMsg.GetStringProperty(unsetPropName))
	retErr = txtMsg.SetStringProperty(unsetPropName, nil)
	assert.Nil(t, retErr)
	assert.Nil(t, txtMsg.GetStringProperty(unsetPropName))

	// Set up objects for send/receive
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Now send the message and get it back again, to check that it roundtripped.
	errSend := context.CreateProducer().SetTimeToLive(10000).Send(queue, txtMsg)
	assert.Nil(t, errSend)

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, msgBody, *msg.GetText())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

	// Check property is available on received message.
	assert.Equal(t, propValue, *rcvMsg.GetStringProperty(propName))

	// Check the empty string property.
	assert.Equal(t, emptyPropValue, *rcvMsg.GetStringProperty(emptyPropName))

	// Properties that are not set should return nil
	assert.Nil(t, rcvMsg.GetStringProperty("nonExistentProperty"))
	assert.Nil(t, rcvMsg.GetStringProperty(unsetPropName))

}

/*
 * Test send and receive of a text message with a string property and no content.
 */
func TestStringPropertyTextMessageNilBody(t *testing.T) {

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

	// Create a TextMessage, and check it has nil content.
	msg := context.CreateTextMessage()
	assert.Nil(t, msg.GetText())

	propName := "myProperty2"
	propValue := "myValue2"
	retErr := msg.SetStringProperty(propName, &propValue)
	assert.Nil(t, retErr)

	// Now send the message and get it back again, to check that it roundtripped.
	queue := context.CreateQueue("DEV.QUEUE.1")
	errSend := context.CreateProducer().SetTimeToLive(10000).Send(queue, msg)
	assert.Nil(t, errSend)

	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.TextMessage:
		assert.Nil(t, msg.GetText())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

	// Check property is available on received message.
	assert.Equal(t, propValue, *rcvMsg.GetStringProperty(propName))

}

/*
 * Test the behaviour for send/receive of a text message with an empty string
 * body. It's difficult to distinguish nil and empty string so we are expecting
 * that the received message will contain a nil body.
 */
func TestStringPropertyTextMessageEmptyBody(t *testing.T) {

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

	// Create a TextMessage
	msg := context.CreateTextMessageWithString("")
	assert.Equal(t, "", *msg.GetText())

	propAName := "myPropertyA"
	propAValue := "myValueA"
	retErr := msg.SetStringProperty(propAName, &propAValue)
	assert.Nil(t, retErr)

	propBName := "myPropertyB"
	propBValue := "myValueB"
	retErr = msg.SetStringProperty(propBName, &propBValue)
	assert.Nil(t, retErr)

	// Now send the message and get it back again.
	queue := context.CreateQueue("DEV.QUEUE.1")
	errSend := context.CreateProducer().Send(queue, msg)
	assert.Nil(t, errSend)

	consumer, errCons := context.CreateConsumer(queue)
	assert.Nil(t, errCons)
	if consumer != nil {
		defer consumer.Close()
	}

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.TextMessage:

		// It's difficult to distinguish between empty string and no string (nil)
		// so we settle for giving back a nil, so that messages containing empty
		// string are equivalent to messages containing no string at all.
		assert.Nil(t, msg.GetText())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

	// Check property is available on received message.
	assert.Equal(t, propAValue, *rcvMsg.GetStringProperty(propAName))
	assert.Equal(t, propBValue, *rcvMsg.GetStringProperty(propBName))

}
