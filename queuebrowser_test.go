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
 * Test the basic Queue Browser behaviour.
 */
func TestQueueBrowser(t *testing.T) {

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

	// Create some TextMessages
	msg1 := context.CreateTextMessageWithString("browser msg 1")
	msg2 := context.CreateTextMessageWithString("browser msg 2")
	msg3 := context.CreateTextMessageWithString("browser msg 3")

	// Now send the messages
	queue := context.CreateQueue("DEV.QUEUE.1")
	producer := context.CreateProducer().SetTimeToLive(20000)
	errSend := producer.Send(queue, msg1)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg2)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg3)
	assert.Nil(t, errSend)

	// Browse for the messages that we put onto the queue.
	browser, errCons := context.CreateBrowser(queue)
	if browser != nil {
		defer browser.Close()
	}
	assert.Nil(t, errCons)

	msgIterator, err := browser.GetEnumeration()
	assert.Nil(t, err)
	assert.NotNil(t, msgIterator)

	// Asking for a second Enumeration from the same Browser returns the
	// same object (only one Enumeration per Browser)
	msgIteratorSame, err := browser.GetEnumeration()
	assert.Equal(t, &msgIterator, &msgIteratorSame)

	gotMsg1, gotErr1 := msgIterator.GetNext()
	assert.Nil(t, gotErr1)
	assert.NotNil(t, gotMsg1)
	assert.Equal(t, msg1.GetJMSMessageID(), gotMsg1.GetJMSMessageID())

	gotMsg2, gotErr2 := msgIterator.GetNext()
	assert.Nil(t, gotErr2)
	assert.NotNil(t, gotMsg2)
	assert.Equal(t, msg2.GetJMSMessageID(), gotMsg2.GetJMSMessageID())

	gotMsg3, gotErr3 := msgIterator.GetNext()
	assert.Nil(t, gotErr3)
	assert.NotNil(t, gotMsg3)
	assert.Equal(t, msg3.GetJMSMessageID(), gotMsg3.GetJMSMessageID())

	// No more messages left.
	gotMsg4, gotErr4 := msgIterator.GetNext()
	assert.Nil(t, gotErr4)
	assert.Nil(t, gotMsg4)

	// Now use a second browser to confirm that the messages were not
	// destructively consumed.
	browser2, errCons := context.CreateBrowser(queue)
	if browser2 != nil {
		defer browser2.Close()
	}
	assert.Nil(t, errCons)

	msgIterator2, err := browser2.GetEnumeration()
	assert.Nil(t, err)
	assert.NotNil(t, msgIterator)

	gotMsg1, gotErr1 = msgIterator2.GetNext()
	assert.Nil(t, gotErr1)
	assert.NotNil(t, gotMsg1) // Failure here shows messages were destructively consumed
	assert.Equal(t, msg1.GetJMSMessageID(), gotMsg1.GetJMSMessageID())

	gotMsg2, gotErr2 = msgIterator2.GetNext()
	assert.Nil(t, gotErr2)
	assert.NotNil(t, gotMsg2)
	assert.Equal(t, msg2.GetJMSMessageID(), gotMsg2.GetJMSMessageID())

	gotMsg3, gotErr3 = msgIterator2.GetNext()
	assert.Nil(t, gotErr3)
	assert.NotNil(t, gotMsg3)
	assert.Equal(t, msg3.GetJMSMessageID(), gotMsg3.GetJMSMessageID())

	// No more messages left.
	gotMsg4, gotErr4 = msgIterator2.GetNext()
	assert.Nil(t, gotErr4)
	assert.Nil(t, gotMsg4)

	// Tidy up the messages by destructively consuming them with a
	// real Consumer.
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	gotMsg1, gotErr1 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr1)
	assert.NotNil(t, gotMsg1)
	assert.Equal(t, msg1.GetJMSMessageID(), gotMsg1.GetJMSMessageID())

	gotMsg2, gotErr2 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr2)
	assert.NotNil(t, gotMsg2)
	assert.Equal(t, msg2.GetJMSMessageID(), gotMsg2.GetJMSMessageID())

	gotMsg3, gotErr3 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr3)
	assert.NotNil(t, gotMsg3)
	assert.Equal(t, msg3.GetJMSMessageID(), gotMsg3.GetJMSMessageID())

	// No more messages left.
	gotMsg4, gotErr4 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr4)
	assert.Nil(t, gotMsg4)

}

/*
 * Test that QueueBrowser provides a live view of the messages on a queue,
 * including if more are added after the browsing begins.
 */
func TestQueueBrowserWhilePutting(t *testing.T) {

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

	// Browse for the messages that we put onto the queue.
	queue := context.CreateQueue("DEV.QUEUE.1")
	browser, errCons := context.CreateBrowser(queue)
	if browser != nil {
		defer browser.Close()
	}
	assert.Nil(t, errCons)

	msgIterator, err := browser.GetEnumeration()
	assert.Nil(t, err)
	assert.NotNil(t, msgIterator)

	// Not currently any messages on the queue.
	gotPreMsg, gotPreErr := msgIterator.GetNext()
	assert.Nil(t, gotPreErr)
	assert.Nil(t, gotPreMsg) // no message

	// Create and send some TextMessages
	producer := context.CreateProducer().SetTimeToLive(20000)
	msg1 := context.CreateTextMessageWithString("browser msg 1")
	msg2 := context.CreateTextMessageWithString("browser msg 2")
	msg3 := context.CreateTextMessageWithString("browser msg 3")
	errSend := producer.Send(queue, msg1)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg2)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg3)
	assert.Nil(t, errSend)

	gotMsg1, gotErr1 := msgIterator.GetNext()
	assert.Nil(t, gotErr1)
	assert.NotNil(t, gotMsg1)
	assert.Equal(t, msg1.GetJMSMessageID(), gotMsg1.GetJMSMessageID())

	gotMsg2, gotErr2 := msgIterator.GetNext()
	assert.Nil(t, gotErr2)
	assert.NotNil(t, gotMsg2)
	assert.Equal(t, msg2.GetJMSMessageID(), gotMsg2.GetJMSMessageID())

	gotMsg3, gotErr3 := msgIterator.GetNext()
	assert.Nil(t, gotErr3)
	assert.NotNil(t, gotMsg3)
	assert.Equal(t, msg3.GetJMSMessageID(), gotMsg3.GetJMSMessageID())

	// No more messages left.
	gotMsg4, gotErr4 := msgIterator.GetNext()
	assert.Nil(t, gotErr4)
	assert.Nil(t, gotMsg4)

	// Send some more messages
	msg4 := context.CreateTextMessageWithString("browser msg 4")
	msg5 := context.CreateTextMessageWithString("browser msg 5")
	errSend = producer.Send(queue, msg4)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg5)
	assert.Nil(t, errSend)

	gotMsg4, gotErr4 = msgIterator.GetNext()
	assert.Nil(t, gotErr4)
	assert.NotNil(t, gotMsg4)
	assert.Equal(t, msg4.GetJMSMessageID(), gotMsg4.GetJMSMessageID())

	gotMsg5, gotErr5 := msgIterator.GetNext()
	assert.Nil(t, gotErr5)
	assert.NotNil(t, gotMsg5)
	assert.Equal(t, msg5.GetJMSMessageID(), gotMsg5.GetJMSMessageID())

	// No more messages left.
	gotMsg6, gotErr6 := msgIterator.GetNext()
	assert.Nil(t, gotErr6)
	assert.Nil(t, gotMsg6)

	// Tidy up the messages by destructively consuming them with a
	// real Consumer.
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	gotMsg1, gotErr1 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr1)
	assert.NotNil(t, gotMsg1)
	assert.Equal(t, msg1.GetJMSMessageID(), gotMsg1.GetJMSMessageID())

	gotMsg2, gotErr2 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr2)
	assert.NotNil(t, gotMsg2)
	assert.Equal(t, msg2.GetJMSMessageID(), gotMsg2.GetJMSMessageID())

	gotMsg3, gotErr3 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr3)
	assert.NotNil(t, gotMsg3)
	assert.Equal(t, msg3.GetJMSMessageID(), gotMsg3.GetJMSMessageID())

	gotMsg4, gotErr4 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr4)
	assert.NotNil(t, gotMsg4)
	assert.Equal(t, msg4.GetJMSMessageID(), gotMsg4.GetJMSMessageID())

	gotMsg5, gotErr5 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr5)
	assert.NotNil(t, gotMsg5)
	assert.Equal(t, msg5.GetJMSMessageID(), gotMsg5.GetJMSMessageID())

	// No more messages left.
	gotMsg6, gotErr6 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr6)
	assert.Nil(t, gotMsg6)

}

/*
 * Test that QueueBrowser provides a live view of the messages on a queue,
 * including if messages are removed after the browsing begins.
 */
func TestQueueBrowserWhileGetting(t *testing.T) {

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

	// Create and send some TextMessages
	queue := context.CreateQueue("DEV.QUEUE.1")
	producer := context.CreateProducer().SetTimeToLive(20000)
	msg1 := context.CreateTextMessageWithString("browser msg 1")
	errSend := producer.Send(queue, msg1)
	assert.Nil(t, errSend)

	// Create a real consumer ready to destructively consume messages.
	consumer, errCons := context.CreateConsumer(queue)
	if consumer != nil {
		defer consumer.Close()
	}
	assert.Nil(t, errCons)

	// Browse for the messages that we put onto the queue.
	browser, errCons := context.CreateBrowser(queue)
	if browser != nil {
		defer browser.Close()
	}
	assert.Nil(t, errCons)

	msgIterator, err := browser.GetEnumeration()
	assert.Nil(t, err)
	assert.NotNil(t, msgIterator)

	// Destructively consume the (one) message we had put.
	gotMsg1, gotErr1 := consumer.ReceiveNoWait()
	assert.Nil(t, gotErr1)
	assert.NotNil(t, gotMsg1)
	assert.Equal(t, msg1.GetJMSMessageID(), gotMsg1.GetJMSMessageID())

	// Not currently any messages on the queue.
	gotPreMsg, gotPreErr := msgIterator.GetNext()
	assert.Nil(t, gotPreErr)
	assert.Nil(t, gotPreMsg) // no message

	// Create and send some TextMessages
	msg2 := context.CreateTextMessageWithString("browser msg 2")
	msg3 := context.CreateTextMessageWithString("browser msg 3")
	msg4 := context.CreateTextMessageWithString("browser msg 4")
	msg5 := context.CreateTextMessageWithString("browser msg 5")
	msg6 := context.CreateTextMessageWithString("browser msg 6")
	errSend = producer.Send(queue, msg2)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg3)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg4)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg5)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg6)
	assert.Nil(t, errSend)

	// Browse a few messages
	gotMsg2, gotErr2 := msgIterator.GetNext()
	assert.Nil(t, gotErr2)
	assert.NotNil(t, gotMsg2)
	assert.Equal(t, msg2.GetJMSMessageID(), gotMsg2.GetJMSMessageID())

	gotMsg3, gotErr3 := msgIterator.GetNext()
	assert.Nil(t, gotErr3)
	assert.NotNil(t, gotMsg3)
	assert.Equal(t, msg3.GetJMSMessageID(), gotMsg3.GetJMSMessageID())

	gotMsg4, gotErr4 := msgIterator.GetNext()
	assert.Nil(t, gotErr4)
	assert.NotNil(t, gotMsg4)
	assert.Equal(t, msg4.GetJMSMessageID(), gotMsg4.GetJMSMessageID())

	// Now destructively read a couple of messages "behind" the browser
	gotMsg2, gotErr2 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr2)
	assert.NotNil(t, gotMsg2)
	assert.Equal(t, msg2.GetJMSMessageID(), gotMsg2.GetJMSMessageID())

	gotMsg3, gotErr3 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr3)
	assert.NotNil(t, gotMsg3)
	assert.Equal(t, msg3.GetJMSMessageID(), gotMsg3.GetJMSMessageID())

	// Browse another message
	gotMsg5, gotErr5 := msgIterator.GetNext()
	assert.Nil(t, gotErr5)
	assert.NotNil(t, gotMsg5)
	assert.Equal(t, msg5.GetJMSMessageID(), gotMsg5.GetJMSMessageID())

	// Put a few more messages
	msg7 := context.CreateTextMessageWithString("browser msg 7")
	msg8 := context.CreateTextMessageWithString("browser msg 8")
	msg9 := context.CreateTextMessageWithString("browser msg 9")
	msg10 := context.CreateTextMessageWithString("browser msg 10")
	errSend = producer.Send(queue, msg7)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg8)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg9)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg10)
	assert.Nil(t, errSend)

	// Browse another message
	gotMsg6, gotErr6 := msgIterator.GetNext()
	assert.Nil(t, gotErr6)
	assert.NotNil(t, gotMsg6)
	assert.Equal(t, msg6.GetJMSMessageID(), gotMsg6.GetJMSMessageID())

	// Destructively consume a few more
	gotMsg4, gotErr4 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr4)
	assert.NotNil(t, gotMsg4)
	assert.Equal(t, msg4.GetJMSMessageID(), gotMsg4.GetJMSMessageID())

	gotMsg5, gotErr5 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr5)
	assert.NotNil(t, gotMsg5)
	assert.Equal(t, msg5.GetJMSMessageID(), gotMsg5.GetJMSMessageID())

	// Browse another message
	gotMsg7, gotErr7 := msgIterator.GetNext()
	assert.Nil(t, gotErr7)
	assert.NotNil(t, gotMsg7)
	assert.Equal(t, msg7.GetJMSMessageID(), gotMsg7.GetJMSMessageID())

	// Destructively consume a message from the "middle" of the queue
	// by using its MessageID.
	selectorStr := "JMSMessageID = '" + msg9.GetJMSMessageID() + "'"
	msgIdConsumer, errCons := context.CreateConsumerWithSelector(queue, selectorStr)
	if msgIdConsumer != nil {
		defer msgIdConsumer.Close()
	}
	assert.Nil(t, errCons)
	gotMsg9, gotErr9 := msgIdConsumer.ReceiveNoWait()
	assert.Nil(t, gotErr9)
	assert.NotNil(t, gotMsg9)
	assert.Equal(t, msg9.GetJMSMessageID(), gotMsg9.GetJMSMessageID())

	// Put more messages
	msg11 := context.CreateTextMessageWithString("browser msg 11")
	msg12 := context.CreateTextMessageWithString("browser msg 12")
	msg13 := context.CreateTextMessageWithString("browser msg 13")
	errSend = producer.Send(queue, msg11)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg12)
	assert.Nil(t, errSend)
	errSend = producer.Send(queue, msg13)
	assert.Nil(t, errSend)

	// Browse over the "missing middle" (msg9)
	gotMsg8, gotErr8 := msgIterator.GetNext()
	assert.Nil(t, gotErr8)
	assert.NotNil(t, gotMsg8)
	assert.Equal(t, msg8.GetJMSMessageID(), gotMsg8.GetJMSMessageID())

	// not 9

	gotMsg10, gotErr10 := msgIterator.GetNext()
	assert.Nil(t, gotErr10)
	assert.NotNil(t, gotMsg10)
	assert.Equal(t, msg10.GetJMSMessageID(), gotMsg10.GetJMSMessageID())

	// Destructively consume "past" the browser
	gotMsg6, gotErr6 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr6)
	assert.NotNil(t, gotMsg6)
	assert.Equal(t, msg6.GetJMSMessageID(), gotMsg6.GetJMSMessageID())

	gotMsg7, gotErr7 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr7)
	assert.NotNil(t, gotMsg7)
	assert.Equal(t, msg7.GetJMSMessageID(), gotMsg7.GetJMSMessageID())

	gotMsg8, gotErr8 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr8)
	assert.NotNil(t, gotMsg8)
	assert.Equal(t, msg8.GetJMSMessageID(), gotMsg8.GetJMSMessageID())

	// Not 9

	gotMsg10, gotErr10 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr10)
	assert.NotNil(t, gotMsg10)
	assert.Equal(t, msg10.GetJMSMessageID(), gotMsg10.GetJMSMessageID())

	gotMsg11, gotErr11 := consumer.ReceiveNoWait()
	assert.Nil(t, gotErr11)
	assert.NotNil(t, gotMsg11)
	assert.Equal(t, msg11.GetJMSMessageID(), gotMsg11.GetJMSMessageID())

	// Browse another message
	// 11 got consumed before being browsed, so next is 12.
	gotMsg12, gotErr12 := msgIterator.GetNext()
	assert.Nil(t, gotErr12)
	assert.NotNil(t, gotMsg12)
	assert.Equal(t, msg12.GetJMSMessageID(), gotMsg12.GetJMSMessageID())

	// Destructively consume 12 + 13
	gotMsg12, gotErr12 = consumer.ReceiveNoWait()
	assert.Nil(t, gotErr12)
	assert.NotNil(t, gotMsg12)
	assert.Equal(t, msg12.GetJMSMessageID(), gotMsg12.GetJMSMessageID())

	gotMsg13, gotErr13 := consumer.ReceiveNoWait()
	assert.Nil(t, gotErr13)
	assert.NotNil(t, gotMsg13)
	assert.Equal(t, msg13.GetJMSMessageID(), gotMsg13.GetJMSMessageID())

	// No more messages left.
	gotMsg13, gotErr13 = msgIterator.GetNext()
	assert.Nil(t, gotErr13)
	assert.Nil(t, gotMsg13)

	// No more messages left to descructively consume.
	gotMsg14, gotErr14 := consumer.ReceiveNoWait()
	assert.Nil(t, gotErr14)
	assert.Nil(t, gotMsg14)

}
