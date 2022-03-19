/*
 * Copyright (c) IBM Corporation 2022
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0, which is available at
 * http://www.eclipse.org/legal/epl-2.0.
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package main

import (
	"math"
	"testing"
	"time"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the retrieval of special header properties
 */
func TestPropertySpecialStringGet(t *testing.T) {

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
	msgBody := "SpecialPropertiesMsg"
	txtMsg := context.CreateTextMessage()
	txtMsg.SetText(msgBody)

	// Set up objects for send/receive
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Now send the message and get it back again, to check that it roundtripped.
	ttlMillis := 20000
	errSend := context.CreateProducer().SetTimeToLive(ttlMillis).Send(queue, txtMsg)
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

	// Check the PutDate
	gotPropValue, propErr := rcvMsg.GetStringProperty("JMS_IBM_PutDate")
	assert.Nil(t, propErr)
	assert.NotNil(t, gotPropValue)
	assert.Equal(t, 8, len(*gotPropValue)) // YYYYMMDD

	putTimestamp := rcvMsg.GetJMSTimestamp()
	unixTimestamp := time.Unix(0, putTimestamp*int64(time.Millisecond))
	expectedDate := unixTimestamp.Format("20060102")
	assert.Equal(t, expectedDate, *gotPropValue)

	// Check the PutTime
	gotPropValue, propErr = rcvMsg.GetStringProperty("JMS_IBM_PutTime")
	assert.Nil(t, propErr)
	assert.NotNil(t, gotPropValue)
	assert.Equal(t, 8, len(*gotPropValue)) // HHMMSSTH

	expectedTime := unixTimestamp.Format("150405")
	assert.Equal(t, expectedTime, (*gotPropValue)[0:6]) // skip the tenths for the check

	// Check the Format
	gotPropValue, propErr = rcvMsg.GetStringProperty("JMS_IBM_Format")
	assert.Nil(t, propErr)
	assert.NotNil(t, gotPropValue)
	assert.Equal(t, "MQSTR", *gotPropValue)

	// Check the expiration is close enough to put timestamp + time to live
	expiration := rcvMsg.GetJMSExpiration()
	expectedExpiration := putTimestamp + int64(ttlMillis)
	expirationDiff := expectedExpiration - expiration
	absDiff := math.Abs(float64(expirationDiff))
	assert.True(t, absDiff < 250) // within 250 ms

}

/*
 * Test the retrieval of special header properties
 */
func TestPropertyFormatSet(t *testing.T) {

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

	// Set up objects for send/receive
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Now send the message and get it back again, to check that it roundtripped.
	ttlMillis := 20000
	producer := context.CreateProducer().SetTimeToLive(ttlMillis)

	// Create a TextMessage and check that we can populate it
	msgBody := []byte{'f', 'm', 't', 'c', 'h', 'k'}
	msg1 := context.CreateBytesMessage()
	msg1.WriteBytes(msgBody)

	// Set the format field.
	myFormat := "MYFMT" // max 8 characters for MQ field
	msg1.SetStringProperty("JMS_IBM_Format", &myFormat)

	errSend := producer.Send(queue, msg1)
	assert.Nil(t, errSend)

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, 6, msg.GetBodyLength())
		assert.Equal(t, msgBody, *msg.ReadBytes())
	default:
		assert.Fail(t, "Got something other than a bytes message")
	}

	// Check the Format round-tripped
	gotPropValue, propErr := rcvMsg.GetStringProperty("JMS_IBM_Format")
	assert.Nil(t, propErr)
	assert.Equal(t, myFormat, *gotPropValue)
	gotPropValue, propErr = rcvMsg.GetStringProperty("JMS_IBM_MQMD_Format")
	assert.Nil(t, propErr)
	assert.Equal(t, myFormat, *gotPropValue)

	// Check we can unset the property
	msg2 := context.CreateBytesMessage()

	// Before it is set.
	gotPropValue, propErr = msg2.GetStringProperty("JMS_IBM_Format")
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	gotPropValue, propErr = msg2.GetStringProperty("JMS_IBM_MQMD_Format")
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)

	// Now start setting
	msg2.SetStringProperty("JMS_IBM_Format", &myFormat)
	gotPropValue, propErr = msg2.GetStringProperty("JMS_IBM_Format")
	assert.Nil(t, propErr)
	assert.NotNil(t, gotPropValue)
	assert.Equal(t, myFormat, *gotPropValue)
	gotPropValue, propErr = msg2.GetStringProperty("JMS_IBM_MQMD_Format")
	assert.Nil(t, propErr)
	assert.NotNil(t, gotPropValue)
	assert.Equal(t, myFormat, *gotPropValue)

	msg2.SetStringProperty("JMS_IBM_Format", nil)
	gotPropValue, propErr = msg2.GetStringProperty("JMS_IBM_Format")
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)
	gotPropValue, propErr = msg2.GetStringProperty("JMS_IBM_MQMD_Format")
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)

	// Send a text message with a custom format, and it will come back
	// as bytes
	msg3 := context.CreateTextMessageWithString("Hello string")
	myStrFmt := "MYTXT"
	msg3.SetStringProperty("JMS_IBM_Format", &myStrFmt)

	errSend = producer.Send(queue, msg3)
	assert.Nil(t, errSend)

	rcvMsg, errRvc = consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	gotPropValue, propErr = rcvMsg.GetStringProperty("JMS_IBM_Format")
	assert.Nil(t, propErr)
	assert.NotNil(t, gotPropValue)
	assert.Equal(t, myStrFmt, *gotPropValue)
	gotPropValue, propErr = rcvMsg.GetStringProperty("JMS_IBM_MQMD_Format")
	assert.Nil(t, propErr)
	assert.NotNil(t, gotPropValue)
	assert.Equal(t, myStrFmt, *gotPropValue)

	switch msg := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, 12, msg.GetBodyLength()) // length of "Hello string"
	default:
		assert.Fail(t, "Got something other than a bytes message")
	}

	// Send a text message with a custom format using the MQMD parameter,
	msg4 := context.CreateTextMessageWithString("Hello string")
	myStrFmt2 := "MYTXTMD"
	msg4.SetStringProperty("JMS_IBM_MQMD_Format", &myStrFmt2)

	errSend = producer.Send(queue, msg4)
	assert.Nil(t, errSend)

	rcvMsg, errRvc = consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	gotPropValue, propErr = rcvMsg.GetStringProperty("JMS_IBM_Format")
	assert.Nil(t, propErr)
	assert.NotNil(t, gotPropValue)
	assert.Equal(t, myStrFmt2, *gotPropValue)
	gotPropValue, propErr = rcvMsg.GetStringProperty("JMS_IBM_MQMD_Format")
	assert.Nil(t, propErr)
	assert.NotNil(t, gotPropValue)
	assert.Equal(t, myStrFmt2, *gotPropValue)

}

/*
 * Test the retrieval of special header properties
 */
func TestPropertyApplData(t *testing.T) {

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

	// Set up objects for send/receive
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Now send the message and get it back again, to check that it roundtripped.
	ttlMillis := 20000
	producer := context.CreateProducer().SetTimeToLive(ttlMillis)

	// Create a TextMessage and check that we can populate it
	msg := context.CreateBytesMessage()
	//txtMsg.SetText(msgBody)

	gotPropValue, propErr := msg.GetStringProperty("JMS_IBM_MQMD_ApplOriginData")
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)

	// Special security privileges are required in order to set ApplOriginData

	errSend := producer.Send(queue, msg)
	assert.Nil(t, errSend)

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, 0, msg.GetBodyLength())
	default:
		assert.Fail(t, "Got something other than a text message")
	}

	// Get the value back again
	gotPropValue, propErr = rcvMsg.GetStringProperty("JMS_IBM_MQMD_ApplOriginData")
	assert.Nil(t, propErr)
	assert.Nil(t, gotPropValue)

}

/*
 * Test the retrieval of special header properties that are Integers
 */
func TestPropertySpecialIntGet(t *testing.T) {

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

	// Create a BytesMessage and check that we can populate it
	sendMsg := context.CreateBytesMessage()

	// Set the special properties.
	putApplType := 6 // MQAT_DEFAULT.  (seems to get written by queue manager, not application)
	sendMsg.SetIntProperty("JMS_IBM_PutApplType", putApplType)
	encoding := 273
	sendMsg.SetIntProperty("JMS_IBM_Encoding", encoding)
	ccsid := 1208
	sendMsg.SetIntProperty("JMS_IBM_Character_Set", ccsid)
	msgType := 8 // MQMT_DATAGRAM
	sendMsg.SetIntProperty("JMS_IBM_MsgType", msgType)

	// Set up objects for send/receive
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Now send the message and get it back again, to check that it roundtripped.
	ttlMillis := 20000
	errSend := context.CreateProducer().SetTimeToLive(ttlMillis).Send(queue, sendMsg)
	assert.Nil(t, errSend)

	rcvMsg, errRvc := consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, 0, msg.GetBodyLength())
	default:
		assert.Fail(t, "Got something other than a bytes message")
	}

	// Check the properties came back as expected.
	gotPropValue, propErr := rcvMsg.GetIntProperty("JMS_IBM_PutApplType")
	assert.Nil(t, propErr)
	assert.Equal(t, putApplType, gotPropValue)

	gotPropValue, propErr = rcvMsg.GetIntProperty("JMS_IBM_Encoding")
	assert.Nil(t, propErr)
	assert.Equal(t, encoding, gotPropValue)

	gotPropValue, propErr = rcvMsg.GetIntProperty("JMS_IBM_Character_Set")
	assert.Nil(t, propErr)
	assert.Equal(t, ccsid, gotPropValue)

	gotPropValue, propErr = rcvMsg.GetIntProperty("JMS_IBM_MQMD_CodedCharSetId")
	assert.Nil(t, propErr)
	assert.Equal(t, ccsid, gotPropValue)

	gotPropValue, propErr = rcvMsg.GetIntProperty("JMS_IBM_MsgType")
	assert.Nil(t, propErr)
	assert.Equal(t, msgType, gotPropValue)

	gotPropValue, propErr = rcvMsg.GetIntProperty("JMS_IBM_MQMD_MsgType")
	assert.Nil(t, propErr)
	assert.Equal(t, msgType, gotPropValue)

	// Create a BytesMessage and check that we can populate it
	sendMsg2 := context.CreateBytesMessage()

	// Set the special properties, using the MQMD property name variants
	ccsid2 := 850
	sendMsg2.SetIntProperty("JMS_IBM_MQMD_CodedCharSetId", ccsid2)
	msgType2 := 2 // MQMT_REPLY
	sendMsg2.SetIntProperty("JMS_IBM_MQMD_MsgType", msgType2)

	// Now send the message and get it back again, to check that it roundtripped.
	errSend = context.CreateProducer().SetTimeToLive(ttlMillis).Send(queue, sendMsg2)
	assert.Nil(t, errSend)

	rcvMsg, errRvc = consumer.ReceiveNoWait()
	assert.Nil(t, errRvc)
	assert.NotNil(t, rcvMsg)

	switch msg := rcvMsg.(type) {
	case jms20subset.BytesMessage:
		assert.Equal(t, 0, msg.GetBodyLength())
	default:
		assert.Fail(t, "Got something other than a bytes message")
	}

	// Check the properties came back as expected.
	gotPropValue, propErr = rcvMsg.GetIntProperty("JMS_IBM_Character_Set")
	assert.Nil(t, propErr)
	assert.Equal(t, ccsid2, gotPropValue)

	gotPropValue, propErr = rcvMsg.GetIntProperty("JMS_IBM_MQMD_CodedCharSetId")
	assert.Nil(t, propErr)
	assert.Equal(t, ccsid2, gotPropValue)

	gotPropValue, propErr = rcvMsg.GetIntProperty("JMS_IBM_MsgType")
	assert.Nil(t, propErr)
	assert.Equal(t, msgType2, gotPropValue)

	gotPropValue, propErr = rcvMsg.GetIntProperty("JMS_IBM_MQMD_MsgType")
	assert.Nil(t, propErr)
	assert.Equal(t, msgType2, gotPropValue)

}
