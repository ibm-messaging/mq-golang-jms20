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
	"testing"

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Test the behaviour of message groups.
 *
 * JMSXGroupID
 * JMSXGroupSeq
 * JMS_IBM_Last_Msg_In_Group
 *
 * https://www.ibm.com/docs/en/ibm-mq/9.2?topic=ordering-grouping-logical-messages
 */
func TestMessageGroup(t *testing.T) {

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

	// Since we need more work to support the "set" operations (see big comment below)
	// lets just do a short test of the "get" behaviour.

	txtMsg1 := context.CreateTextMessage()

	// Force the population of the MQMD field.
	myFormat := "MYFMT"
	txtMsg1.SetStringProperty("JMS_IBM_Format", &myFormat)

	groupId, err := txtMsg1.GetStringProperty("JMSXGroupID")
	assert.Nil(t, err)
	assert.Nil(t, groupId)

	groupSeq, err := txtMsg1.GetIntProperty("JMSXGroupSeq")
	assert.Nil(t, err)
	assert.Equal(t, 1, groupSeq)

	gotLastMsg, err := txtMsg1.GetBooleanProperty("JMS_IBM_Last_Msg_In_Group")
	assert.Equal(t, false, gotLastMsg)

	myGroup := "hello"
	err = txtMsg1.SetStringProperty("JMSXGroupID", &myGroup)
	assert.NotNil(t, err)
	assert.Equal(t, "Not yet implemented", err.GetLinkedError().Error())

	err = txtMsg1.SetIntProperty("JMSXGroupSeq", 2)
	assert.NotNil(t, err)
	assert.Equal(t, "Not yet implemented", err.GetLinkedError().Error())

	err = txtMsg1.SetBooleanProperty("JMS_IBM_Last_Msg_In_Group", true)
	assert.NotNil(t, err)
	assert.Equal(t, "Not yet implemented", err.GetLinkedError().Error())

	/*
		 * Setting these properties requires an MQMD V2 header and is also
		 * not supported for PUT1 operations so there is some more extensive
		 * implementation work required in order to enable the "set" scenarios
		 * for these Group properties.

		// Create a TextMessage and check that we can populate it
		txtMsg1 := context.CreateTextMessage()
		txtMsg1.SetText(msgBody)
		txtMsg1.SetStringProperty("JMSXGroupID", &groupID)
		txtMsg1.SetIntProperty("JMSXGroupSeq", 1)
		errSend := producer.Send(queue, txtMsg1)
		assert.Nil(t, errSend)

		txtMsg2 := context.CreateTextMessage()
		txtMsg2.SetText(msgBody)
		txtMsg2.SetStringProperty("JMSXGroupID", &groupID)
		txtMsg2.SetIntProperty("JMSXGroupSeq", 2)
		errSend = producer.Send(queue, txtMsg2)
		assert.Nil(t, errSend)

		txtMsg3 := context.CreateTextMessage()
		txtMsg3.SetText(msgBody)
		txtMsg3.SetStringProperty("JMSXGroupID", &groupID)
		txtMsg3.SetIntProperty("JMSXGroupSeq", 3)
		txtMsg3.SetBooleanProperty("JMS_IBM_Last_Msg_In_Group", true)
		errSend = producer.Send(queue, txtMsg3)
		assert.Nil(t, errSend)

		// Check the first message.
		rcvMsg, errRvc := consumer.ReceiveNoWait()
		assert.Nil(t, errRvc)
		assert.NotNil(t, rcvMsg)
		assert.Equal(t, txtMsg1.GetJMSMessageID(), rcvMsg.GetJMSMessageID())
		gotGroupIDValue, gotErr := rcvMsg.GetStringProperty("JMSXGroupID")
		assert.Nil(t, gotErr)
		assert.Equal(t, groupID, *gotGroupIDValue)
		gotSeqValue, gotErr := rcvMsg.GetIntProperty("JMSXGroupSeq")
		assert.Equal(t, 1, gotSeqValue)
		gotLastMsgValue, gotErr := rcvMsg.GetBooleanProperty("JMS_IBM_Last_Msg_In_Group")
		assert.Equal(t, false, gotLastMsgValue)

		// Check the second message.
		rcvMsg, errRvc = consumer.ReceiveNoWait()
		assert.Nil(t, errRvc)
		assert.NotNil(t, rcvMsg)
		assert.Equal(t, txtMsg2.GetJMSMessageID(), rcvMsg.GetJMSMessageID())
		gotGroupIDValue, gotErr = rcvMsg.GetStringProperty("JMSXGroupID")
		assert.Nil(t, gotErr)
		assert.Equal(t, groupID, *gotGroupIDValue)
		gotSeqValue, gotErr = rcvMsg.GetIntProperty("JMSXGroupSeq")
		assert.Equal(t, 2, gotSeqValue)
		gotLastMsgValue, gotErr = rcvMsg.GetBooleanProperty("JMS_IBM_Last_Msg_In_Group")
		assert.Equal(t, false, gotLastMsgValue)

		// Check the third message.
		rcvMsg, errRvc = consumer.ReceiveNoWait()
		assert.Nil(t, errRvc)
		assert.NotNil(t, rcvMsg)
		assert.Equal(t, txtMsg3.GetJMSMessageID(), rcvMsg.GetJMSMessageID())
		gotGroupIDValue, gotErr = rcvMsg.GetStringProperty("JMSXGroupID")
		assert.Nil(t, gotErr)
		assert.Equal(t, groupID, *gotGroupIDValue)
		gotSeqValue, gotErr = rcvMsg.GetIntProperty("JMSXGroupSeq")
		assert.Equal(t, 3, gotSeqValue)
		gotLastMsgValue, gotErr = rcvMsg.GetBooleanProperty("JMS_IBM_Last_Msg_In_Group")
		assert.Equal(t, true, gotLastMsgValue)
	*/

}
