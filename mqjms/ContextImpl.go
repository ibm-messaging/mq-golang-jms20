// Copyright (c) IBM Corporation 2019.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Package mqjms provides the implementation of the JMS style Golang interfaces to communicate with IBM MQ.
package mqjms

import (
	"strconv"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	ibmmq "github.com/ibm-messaging/mq-golang/v5/ibmmq"
)

// ContextImpl encapsulates the objects necessary to maintain an active
// connection to an IBM MQ queue manager.
type ContextImpl struct {
	qMgr              ibmmq.MQQueueManager
	sessionMode       int
	receiveBufferSize int
}

// CreateQueue implements the logic necessary to create a provider-specific
// object representing an IBM MQ queue.
func (ctx ContextImpl) CreateQueue(queueName string) jms20subset.Queue {

	// Store the name of the queue
	queue := QueueImpl{
		queueName: queueName,
	}

	return queue
}

// CreateProducer implements the logic necessary to create a JMSProducer object
// that allows messages to be sent to destinations in IBM MQ.
func (ctx ContextImpl) CreateProducer() jms20subset.JMSProducer {

	// Initialise the Producer with the attributes necessary for it to send
	// messages.
	producer := ProducerImpl{
		ctx:          ctx,
		deliveryMode: jms20subset.DeliveryMode_PERSISTENT,
	}

	return &producer
}

// CreateConsumer creates a consumer object that allows an application to
// receive messages from the specified Destination.
func (ctx ContextImpl) CreateConsumer(dest jms20subset.Destination) (jms20subset.JMSConsumer, jms20subset.JMSException) {
	return ctx.CreateConsumerWithSelector(dest, "")
}

// CreateConsumerWithSelector creates a consumer object that allows an application to
// receive messages that match the specified selector from the given Destination.
func (ctx ContextImpl) CreateConsumerWithSelector(dest jms20subset.Destination, selector string) (jms20subset.JMSConsumer, jms20subset.JMSException) {

	// First validate the selector string format (we don't make use of it at
	// runtime until the receive is called)
	if selector != "" {
		getmqmd := ibmmq.NewMQMD()
		gmo := ibmmq.NewMQGMO()

		selectorErr := applySelector(selector, getmqmd, gmo)
		if selectorErr != nil {
			return nil, jms20subset.CreateJMSException("Invalid selector syntax", "MQJMS0004", selectorErr)
		}
	}

	// Set up the necessary objects to open the queue
	mqod := ibmmq.NewMQOD()
	var openOptions int32
	openOptions = ibmmq.MQOO_FAIL_IF_QUIESCING
	openOptions |= ibmmq.MQOO_INPUT_AS_Q_DEF
	mqod.ObjectType = ibmmq.MQOT_Q
	mqod.ObjectName = dest.GetDestinationName()

	var retErr jms20subset.JMSException
	var consumer jms20subset.JMSConsumer

	// Invoke the MQ command to open the queue.
	qObject, err := ctx.qMgr.Open(mqod, openOptions)

	if err == nil {

		// Success - store the necessary objects away for later use to receive
		// messages.
		consumer = ConsumerImpl{
			ctx:      ctx,
			qObject:  qObject,
			selector: selector,
		}

	} else {

		// Error occurred - extract the failure details and return to the caller.
		rcInt := int(err.(*ibmmq.MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := ibmmq.MQItoString("RC", rcInt)
		retErr = jms20subset.CreateJMSException(reason, errCode, err)

	}

	return consumer, retErr
}

// CreateTextMessage is a JMS standard mechanism for creating a TextMessage.
func (ctx ContextImpl) CreateTextMessage() jms20subset.TextMessage {
	return &TextMessageImpl{}
}

// CreateTextMessageWithString is a JMS standard mechanism for creating a TextMessage
// and initialise it with the chosen text string.
func (ctx ContextImpl) CreateTextMessageWithString(txt string) jms20subset.TextMessage {
	msg := TextMessageImpl{}
	msg.SetText(txt)
	return &msg
}

// CreateBytesMessage is a JMS standard mechanism for creating a BytesMessage.
func (ctx ContextImpl) CreateBytesMessage() jms20subset.BytesMessage {
	return &BytesMessageImpl{}
}

// CreateBytesMessageWithBytes is a JMS standard mechanism for creating a BytesMessage.
func (ctx ContextImpl) CreateBytesMessageWithBytes(bytes []byte) jms20subset.BytesMessage {
	msg := BytesMessageImpl{}
	msg.WriteBytes(bytes)
	return &msg
}

// Commit confirms all messages that were sent under this transaction.
func (ctx ContextImpl) Commit() {

	if (ibmmq.MQQueueManager{}) != ctx.qMgr {
		ctx.qMgr.Cmit()
	}

}

// Rollback releases all messages that were sent under this transaction.
func (ctx ContextImpl) Rollback() {

	if (ibmmq.MQQueueManager{}) != ctx.qMgr {
		ctx.qMgr.Back()
	}

}

// Close this connection to the MQ queue manager, and release any resources
// that were allocated to support this connection.
func (ctx ContextImpl) Close() {

	// JMS semantics are to roll back an active transaction on Close.
	ctx.Rollback()

	if (ibmmq.MQQueueManager{}) != ctx.qMgr {
		ctx.qMgr.Disc()
	}

}
