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
 * Property conversion between types
 *
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
	gotPropValue, propErr := txtMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)

	// Test the ability to set properties before the message is sent.
	retErr := txtMsg.SetStringProperty(propName, &propValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, *gotPropValue)
	assert.Equal(t, msgBody, *txtMsg.GetText())

	// Send an empty string property as well
	emptyPropName := "myEmptyString"
	emptyPropValue := ""
	retErr = txtMsg.SetStringProperty(emptyPropName, &emptyPropValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(emptyPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, emptyPropValue, *gotPropValue)

	// Set a property then try to unset it by setting to nil
	unsetPropName := "mySendThenRemovedString"
	unsetPropValue := "someValueThatWillBeOverwritten"
	retErr = txtMsg.SetStringProperty(unsetPropName, &unsetPropValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, unsetPropValue, *gotPropValue)
	retErr = txtMsg.SetStringProperty(unsetPropName, nil)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)

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
	gotPropValue, propErr = rcvMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, *gotPropValue)

	// Check the empty string property.
	gotPropValue, propErr = rcvMsg.GetStringProperty(emptyPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, emptyPropValue, *gotPropValue)

	// Properties that are not set should return nil
	gotPropValue, propErr = rcvMsg.GetStringProperty("nonExistentProperty")
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	gotPropValue, propErr = rcvMsg.GetStringProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)

}

/*
 * Test the Exists and GetNames functions for message properties
 */
func TestPropertyExistsGetNames(t *testing.T) {

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
	msgBody := "ExistsGetNames-test"
	txtMsg := context.CreateTextMessage()
	txtMsg.SetText(msgBody)

	propName := "myProperty"
	propValue := "myValue"

	// Test the empty value before the property is set.
	gotPropValue, propErr := txtMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	propExists, propErr := txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)
	allPropNames, getNamesErr := txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 0, len(allPropNames))

	// Test the ability to set properties before the message is sent.
	retErr := txtMsg.SetStringProperty(propName, &propValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, *gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists
	allPropNames, getNamesErr = txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 1, len(allPropNames))
	assert.Equal(t, propName, allPropNames[0])

	propName2 := "myPropertyTwo"
	propValue2 := "myValueTwo"
	retErr = txtMsg.SetStringProperty(propName2, &propValue2)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(propName2)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue2, *gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(propName2)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists
	// Check the first property again to be sure
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists
	allPropNames, getNamesErr = txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 2, len(allPropNames))
	assert.Equal(t, propName, allPropNames[0])
	assert.Equal(t, propName2, allPropNames[1])

	// Set a property then try to unset it by setting to nil
	unsetPropName := "mySendThenRemovedString"
	unsetPropValue := "someValueThatWillBeOverwritten"
	retErr = txtMsg.SetStringProperty(unsetPropName, &unsetPropValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, unsetPropValue, *gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(unsetPropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists)
	allPropNames, getNamesErr = txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 3, len(allPropNames))
	retErr = txtMsg.SetStringProperty(unsetPropName, nil)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(unsetPropName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)
	allPropNames, getNamesErr = txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 2, len(allPropNames))

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
	propExists, propErr = rcvMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists

	propExists, propErr = rcvMsg.PropertyExists(propName2)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists

	// Check GetPropertyNames
	allPropNames, getNamesErr = rcvMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 2, len(allPropNames))
	assert.Equal(t, propName, allPropNames[0])
	assert.Equal(t, propName2, allPropNames[1])

	// Properties that are not set should return nil
	nonExistentPropName := "nonExistentProperty"
	gotPropValue, propErr = rcvMsg.GetStringProperty(nonExistentPropName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	propExists, propErr = rcvMsg.PropertyExists(nonExistentPropName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

	// Check for the unset property
	propExists, propErr = rcvMsg.PropertyExists(unsetPropName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

}

/*
 * Test the ClearProperties function for message properties
 */
func TestPropertyClearProperties(t *testing.T) {

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
	msgBody := "ExistsClearProperties-test"
	txtMsg := context.CreateTextMessage()
	txtMsg.SetText(msgBody)

	propName := "myProperty"
	propValue := "myValue"

	// Test the ability to set properties before the message is sent.
	retErr := txtMsg.SetStringProperty(propName, &propValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr := txtMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, *gotPropValue)
	propExists, propErr := txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists
	allPropNames, getNamesErr := txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 1, len(allPropNames))
	assert.Equal(t, propName, allPropNames[0])

	clearErr := txtMsg.ClearProperties()
	assert.Nil(t, clearErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

	allPropNames, getNamesErr = txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 0, len(allPropNames))

	propName2 := "myPropertyTwo"
	propValue2 := 246811

	// Set multiple properties
	retErr = txtMsg.SetStringProperty(propName, &propValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, *gotPropValue)
	retErr = txtMsg.SetIntProperty(propName2, propValue2)
	assert.Nil(t, retErr)
	gotPropValue2, propErr := txtMsg.GetIntProperty(propName2)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue2, gotPropValue2)
	propExists, propErr = txtMsg.PropertyExists(propName2)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists
	// Check the first property again to be sure
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists

	allPropNames, getNamesErr = txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 2, len(allPropNames))

	// Set a property then try to unset it by setting to nil
	unsetPropName := "mySendThenRemovedString"
	unsetPropValue := "someValueThatWillBeOverwritten"
	retErr = txtMsg.SetStringProperty(unsetPropName, &unsetPropValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, unsetPropValue, *gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(unsetPropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists)
	allPropNames, getNamesErr = txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 3, len(allPropNames))
	assert.Equal(t, propName, allPropNames[0])
	assert.Equal(t, propName2, allPropNames[1])
	assert.Equal(t, unsetPropName, allPropNames[2])
	retErr = txtMsg.SetStringProperty(unsetPropName, nil)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(unsetPropName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)
	allPropNames, getNamesErr = txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 2, len(allPropNames))
	assert.Equal(t, propName, allPropNames[0])
	assert.Equal(t, propName2, allPropNames[1])

	clearErr = txtMsg.ClearProperties()
	assert.Nil(t, clearErr)
	gotPropValue, propErr = txtMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)
	allPropNames, getNamesErr = txtMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 0, len(allPropNames))

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
	propExists, propErr = rcvMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

	propExists, propErr = rcvMsg.PropertyExists(propName2)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

	allPropNames, getNamesErr = rcvMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 0, len(allPropNames))

	// Properties that are not set should return nil
	nonExistentPropName := "nonExistentProperty"
	gotPropValue, propErr = rcvMsg.GetStringProperty(nonExistentPropName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	propExists, propErr = rcvMsg.PropertyExists(nonExistentPropName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

	// Check for the unset property
	propExists, propErr = rcvMsg.PropertyExists(unsetPropName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

	// Finally try clearing everything on the received message
	clearErr = rcvMsg.ClearProperties()
	assert.Nil(t, clearErr)
	gotPropValue, propErr = rcvMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	propExists, propErr = rcvMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)
	allPropNames, getNamesErr = rcvMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 0, len(allPropNames))

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
	gotPropValue, propErr := rcvMsg.GetStringProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, *gotPropValue)

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
	gotPropValue, propErr := rcvMsg.GetStringProperty(propAName)
	assert.Nil(t, propErr)
	assert.Equal(t, propAValue, *gotPropValue)
	gotPropValue, propErr = rcvMsg.GetStringProperty(propBName)
	assert.Nil(t, propErr)
	assert.Equal(t, propBValue, *gotPropValue)

}

/*
 * Test the creation of a text message with an int property.
 */
func TestIntProperty(t *testing.T) {

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
	msgBody := "IntPropertyRequestMsg"
	txtMsg := context.CreateTextMessage()
	txtMsg.SetText(msgBody)

	propName := "myProperty"
	propValue := 6

	// Test the empty value before the property is set.
	gotPropValue, propErr := txtMsg.GetIntProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, 0, gotPropValue)
	propExists, propErr := txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

	// Test the ability to set properties before the message is sent.
	retErr := txtMsg.SetIntProperty(propName, propValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetIntProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, gotPropValue)
	assert.Equal(t, msgBody, *txtMsg.GetText())
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists

	propName2 := "myProperty2"
	propValue2 := 246810
	retErr = txtMsg.SetIntProperty(propName2, propValue2)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetIntProperty(propName2)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue2, gotPropValue)

	// Set a property then try to "unset" it by setting to 0
	unsetPropName := "mySendThenRemovedString"
	unsetPropValue := 12345
	retErr = txtMsg.SetIntProperty(unsetPropName, unsetPropValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetIntProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, unsetPropValue, gotPropValue)
	retErr = txtMsg.SetIntProperty(unsetPropName, 0)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetIntProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, 0, gotPropValue)

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
	gotPropValue, propErr = rcvMsg.GetIntProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists

	gotPropValue, propErr = rcvMsg.GetIntProperty(propName2)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue2, gotPropValue)

	// Properties that are not set should return nil
	gotPropValue, propErr = rcvMsg.GetIntProperty("nonExistentProperty")
	assert.Nil(t, propErr)
	assert.Equal(t, 0, gotPropValue)
	gotPropValue, propErr = rcvMsg.GetIntProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, 0, gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(unsetPropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // exists, even though it is set to zero

}

/*
 * Test the creation of a text message with a double property.
 */
func TestDoubleProperty(t *testing.T) {

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
	msgBody := "DoublePropertyRequestMsg"
	txtMsg := context.CreateTextMessage()
	txtMsg.SetText(msgBody)

	propName := "myProperty"
	propValue := float64(15867494.43857438)

	// Test the empty value before the property is set.
	gotPropValue, propErr := txtMsg.GetDoubleProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, float64(0), gotPropValue)
	propExists, propErr := txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

	// Test the ability to set properties before the message is sent.
	retErr := txtMsg.SetDoubleProperty(propName, propValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetDoubleProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, gotPropValue)
	assert.Equal(t, msgBody, *txtMsg.GetText())
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists

	propName2 := "myProperty2"
	propValue2 := float64(-246810.2255343676)
	retErr = txtMsg.SetDoubleProperty(propName2, propValue2)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetDoubleProperty(propName2)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue2, gotPropValue)

	// Set a property then try to "unset" it by setting to 0
	unsetPropName := "mySendThenRemovedString"
	unsetPropValue := float64(12345.123456)
	retErr = txtMsg.SetDoubleProperty(unsetPropName, unsetPropValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetDoubleProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, unsetPropValue, gotPropValue)
	retErr = txtMsg.SetDoubleProperty(unsetPropName, 0)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetDoubleProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, float64(0), gotPropValue)

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
	gotPropValue, propErr = rcvMsg.GetDoubleProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists

	gotPropValue, propErr = rcvMsg.GetDoubleProperty(propName2)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue2, gotPropValue)

	// Properties that are not set should return nil
	gotPropValue, propErr = rcvMsg.GetDoubleProperty("nonExistentProperty")
	assert.Nil(t, propErr)
	assert.Equal(t, float64(0), gotPropValue)
	gotPropValue, propErr = rcvMsg.GetDoubleProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, float64(0), gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(unsetPropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // exists, even though it is set to zero

}

/*
 * Test the creation of a text message with a boolean property.
 */
func TestBooleanProperty(t *testing.T) {

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
	msgBody := "BooleanPropertyRequestMsg"
	txtMsg := context.CreateTextMessage()
	txtMsg.SetText(msgBody)

	propName := "myProperty"
	propValue := true

	// Test the empty value before the property is set.
	gotPropValue, propErr := txtMsg.GetBooleanProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, false, gotPropValue)
	propExists, propErr := txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)

	// Test the ability to set properties before the message is sent.
	retErr := txtMsg.SetBooleanProperty(propName, propValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetBooleanProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, gotPropValue)
	assert.Equal(t, msgBody, *txtMsg.GetText())
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists

	propName2 := "myProperty2"
	propValue2 := false
	retErr = txtMsg.SetBooleanProperty(propName2, propValue2)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetBooleanProperty(propName2)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue2, gotPropValue)

	// Set a property then try to "unset" it by setting to 0
	unsetPropName := "mySendThenRemovedString"
	unsetPropValue := true
	retErr = txtMsg.SetBooleanProperty(unsetPropName, unsetPropValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetBooleanProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, unsetPropValue, gotPropValue)
	retErr = txtMsg.SetBooleanProperty(unsetPropName, false)
	assert.Nil(t, retErr)
	gotPropValue, propErr = txtMsg.GetBooleanProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, false, gotPropValue)

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
	gotPropValue, propErr = rcvMsg.GetBooleanProperty(propName)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue, gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(propName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // now exists

	gotPropValue, propErr = rcvMsg.GetBooleanProperty(propName2)
	assert.Nil(t, propErr)
	assert.Equal(t, propValue2, gotPropValue)

	// Properties that are not set should return nil
	gotPropValue, propErr = rcvMsg.GetBooleanProperty("nonExistentProperty")
	assert.Nil(t, propErr)
	assert.Equal(t, false, gotPropValue)
	gotPropValue, propErr = rcvMsg.GetBooleanProperty(unsetPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, false, gotPropValue)
	propExists, propErr = txtMsg.PropertyExists(unsetPropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists) // exists, even though it is set to zero

}

/*
 * Test the creation of a bytes message with message properties.
 */
func TestPropertyBytesMsg(t *testing.T) {

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

	// Create a BytesMessage
	msgBody := []byte{'b', 'y', 't', 'e', 's', 'p', 'r', 'o', 'p', 'e', 'r', 't', 'i', 'e', 's'}
	bytesMsg := context.CreateBytesMessage()
	bytesMsg.WriteBytes(msgBody)
	assert.Equal(t, 15, bytesMsg.GetBodyLength())
	assert.Equal(t, msgBody, *bytesMsg.ReadBytes())

	stringPropName := "myProperty"
	stringPropValue := "myValue"

	// Test the empty value before the property is set.
	gotPropValue, propErr := bytesMsg.GetStringProperty(stringPropName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)

	// Test the ability to set properties before the message is sent.
	retErr := bytesMsg.SetStringProperty(stringPropName, &stringPropValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = bytesMsg.GetStringProperty(stringPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, stringPropValue, *gotPropValue)

	// Send an empty string property as well
	emptyPropName := "myEmptyString"
	emptyPropValue := ""
	retErr = bytesMsg.SetStringProperty(emptyPropName, &emptyPropValue)
	assert.Nil(t, retErr)
	gotPropValue, propErr = bytesMsg.GetStringProperty(emptyPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, emptyPropValue, *gotPropValue)

	// Now an int property
	intPropName := "myIntProperty"
	intPropValue := 553786
	retErr = bytesMsg.SetIntProperty(intPropName, intPropValue)
	assert.Nil(t, retErr)
	gotIntPropValue, propErr := bytesMsg.GetIntProperty(intPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, intPropValue, gotIntPropValue)

	// Now a double property
	doublePropName := "myDoubleProperty"
	doublePropValue := float64(3.1415926535)
	retErr = bytesMsg.SetDoubleProperty(doublePropName, doublePropValue)
	assert.Nil(t, retErr)
	gotDoublePropValue, propErr := bytesMsg.GetDoubleProperty(doublePropName)
	assert.Nil(t, propErr)
	assert.Equal(t, doublePropValue, gotDoublePropValue)

	// Now a bool property
	boolPropName := "myBoolProperty"
	boolPropValue := true
	retErr = bytesMsg.SetBooleanProperty(boolPropName, boolPropValue)
	assert.Nil(t, retErr)
	gotBoolPropValue, propErr := bytesMsg.GetBooleanProperty(boolPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, boolPropValue, gotBoolPropValue)

	// Set up objects for send/receive
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Now send the message and get it back again, to check that it roundtripped.
	errSend := context.CreateProducer().SetTimeToLive(10000).Send(queue, bytesMsg)
	assert.Nil(t, errSend)

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg2 := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, len(msgBody), msg2.GetBodyLength())
		assert.Equal(t, msgBody, *msg2.ReadBytes())
	default:
		assert.Fail(t, "Got something other than a bytes message")
	}

	// Check property is available on received message.
	propExists, propErr := rcvMsg.PropertyExists(stringPropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists)
	gotPropValue, propErr = rcvMsg.GetStringProperty(stringPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, stringPropValue, *gotPropValue)

	// Check the empty string property.
	propExists, propErr = rcvMsg.PropertyExists(emptyPropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists)
	gotPropValue, propErr = rcvMsg.GetStringProperty(emptyPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, emptyPropValue, *gotPropValue)

	// Properties that are not set should return nil
	nonExistPropName := "nonExistentProperty"
	propExists, propErr = rcvMsg.PropertyExists(nonExistPropName)
	assert.Nil(t, propErr)
	assert.False(t, propExists)
	gotPropValue, propErr = rcvMsg.GetStringProperty(nonExistPropName)
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)

	propExists, propErr = rcvMsg.PropertyExists(intPropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists)
	gotIntPropValue, propErr = rcvMsg.GetIntProperty(intPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, intPropValue, gotIntPropValue)

	propExists, propErr = rcvMsg.PropertyExists(doublePropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists)
	gotDoublePropValue, propErr = rcvMsg.GetDoubleProperty(doublePropName)
	assert.Nil(t, propErr)
	assert.Equal(t, doublePropValue, gotDoublePropValue)

	propExists, propErr = rcvMsg.PropertyExists(boolPropName)
	assert.Nil(t, propErr)
	assert.True(t, propExists)
	gotBoolPropValue, propErr = rcvMsg.GetBooleanProperty(boolPropName)
	assert.Nil(t, propErr)
	assert.Equal(t, boolPropValue, gotBoolPropValue)

	allPropNames, getNamesErr := rcvMsg.GetPropertyNames()
	assert.Nil(t, getNamesErr)
	assert.Equal(t, 5, len(allPropNames))

}
