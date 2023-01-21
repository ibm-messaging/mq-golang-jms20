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
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	ibmmq "github.com/ibm-messaging/mq-golang/v5/ibmmq"
)

// ProducerImpl defines a struct that contains the necessary objects for
// sending messages to a queue on an IBM MQ queue manager.
type ProducerImpl struct {
	ctx          ContextImpl
	deliveryMode int
	timeToLive   int
	priority     int
}

// SendString sends a TextMessage with the specified body to the specified Destination
// using any message options that are defined on this JMSProducer.
func (producer ProducerImpl) SendString(dest jms20subset.Destination, bodyStr string) jms20subset.JMSException {

	// This is essentially just a helper method that avoids the application having
	// to create its own TextMessage object.
	msg := producer.ctx.CreateTextMessage()
	msg.SetText(bodyStr)

	return producer.Send(dest, msg)

}

// SendBytes sends a BytesMessage with the specified body to the specified Destination
// using any message options that are defined on this JMSProducer.
func (producer ProducerImpl) SendBytes(dest jms20subset.Destination, body []byte) jms20subset.JMSException {

	// This is essentially just a helper method that avoids the application having
	// to create its own TextMessage object.
	msg := producer.ctx.CreateBytesMessage()
	msg.WriteBytes(body)

	return producer.Send(dest, msg)

}

// Send a message to the specified IBM MQ queue, using the message options
// that are defined on this JMSProducer.
func (producer ProducerImpl) Send(dest jms20subset.Destination, msg jms20subset.Message) jms20subset.JMSException {

	// Lock the context while we are making calls to the queue manager so that it
	// doesn't conflict with the finalizer we use (below) to delete unused MessageHandles.
	producer.ctx.ctxLock.Lock()
	defer producer.ctx.ctxLock.Unlock()

	// Set up the basic objects we need to send the message.
	mqod := ibmmq.NewMQOD()
	putmqmd := ibmmq.NewMQMD()
	pmo := ibmmq.NewMQPMO()

	var retErr jms20subset.JMSException

	// Setup destination
	mqod.ObjectType = ibmmq.MQOT_Q
	mqod.ObjectName = dest.GetDestinationName()

	// Calculate the syncpoint value
	syncpointSetting := ibmmq.MQPMO_NO_SYNCPOINT
	if producer.ctx.sessionMode == jms20subset.JMSContextSESSIONTRANSACTED {
		syncpointSetting = ibmmq.MQPMO_SYNCPOINT
	}

	// Configure the put message options, including asking MQ to allocate a
	// unique message ID
	pmo.Options = syncpointSetting | ibmmq.MQPMO_NEW_MSG_ID

	// Is async put has been requested then apply the appropriate PMO option
	if dest.GetPutAsyncAllowed() == jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED {
		pmo.Options |= ibmmq.MQPMO_ASYNC_RESPONSE
	}

	var buffer []byte

	// We have a "Message" object and can use a switch to safely convert it
	// to the implementation type in order to extract generic MQ message
	switch typedMsg := msg.(type) {
	case *TextMessageImpl:

		// If the message already has an MQMD then use that (for example it might
		// contain ReplyTo information)
		if typedMsg.mqmd != nil {
			putmqmd = typedMsg.mqmd
		}

		// Pass up the handle containing the message properties
		pmo.OriginalMsgHandle = *typedMsg.msgHandle

		// Store the Put MQMD so that we can later retrieve "out" fields like MsgId
		typedMsg.mqmd = putmqmd

		// Set up this MQ message to contain the string from the JMS message.
		trimmedFormat := strings.TrimSpace(putmqmd.Format)
		if trimmedFormat == ibmmq.MQFMT_NONE {
			putmqmd.Format = ibmmq.MQFMT_STRING
		}

		msgStr := typedMsg.GetText()
		if msgStr != nil {
			buffer = []byte(*msgStr)
		}

	case *BytesMessageImpl:

		// If the message already has an MQMD then use that (for example it might
		// contain ReplyTo information)
		if typedMsg.mqmd != nil {
			putmqmd = typedMsg.mqmd
		}

		// Pass up the handle containing the message properties
		pmo.OriginalMsgHandle = *typedMsg.msgHandle

		// Store the Put MQMD so that we can later retrieve "out" fields like MsgId
		typedMsg.mqmd = putmqmd

		// Set up this MQ message to contain the bytes from the JMS message.
		buffer = *typedMsg.ReadBytes()

	default:
		// This "should never happen"(!) apart from in situations where we are
		// part way through adding support for a new message type to this library.
		log.Fatal(jms20subset.CreateJMSException("UnexpectedMessageType", "UnexpectedMessageType-send1", nil))
	}

	// Convert the JMS persistence into the equivalent MQ message descriptor
	// attribute.
	if producer.deliveryMode == jms20subset.DeliveryMode_NON_PERSISTENT {
		putmqmd.Persistence = ibmmq.MQPER_NOT_PERSISTENT
	} else {
		putmqmd.Persistence = ibmmq.MQPER_PERSISTENT
	}

	// If the producer has a TTL specified then apply it to the put MQMD so
	// that MQ will honour it.
	if producer.timeToLive > 0 {
		// Note that JMS timeToLive in milliseconds, whereas MQMD Expiry expects
		// 10ths of a second
		putmqmd.Expiry = (int32(producer.timeToLive) / 100)
	}

	// Convert the JMS priority into the equivalent MQ message descriptor
	// attribute.
	putmqmd.Priority = int32(producer.priority)

	// Invoke the MQ command to put the message using MQPUT1 to avoid MQOPEN and MQCLOSE.
	// Any Err that occurs will be handled below.
	err := producer.ctx.qMgr.Put1(mqod, putmqmd, pmo, buffer)

	// If the user is using non-transactional async-put and requested non-zero send check
	// count then this is the point at which we carry out the check for errors.
	//
	// Note that if there is already an error returned from Put then just pass that back to
	// the user (only go into this if err is nil).
	if dest.GetPutAsyncAllowed() == jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED &&
		syncpointSetting == ibmmq.MQPMO_NO_SYNCPOINT &&
		producer.ctx.sendCheckCount > 0 &&
		err == nil {

		// Decrement the counter to indicate that a message has been put
		*producer.ctx.sendCheckCountInc--

		// If we have reached zero then it is time to do an error check.
		//
		// Note that the counter is initialized to 1 (in ConnectionFactoryImpl.go) when
		// first configured so that we carry out an error check after the first message
		// in order to catch any errors quickly. After that the check takes place at the
		// interval the user requested in ConnectionFactoryImpl.SendCheckCount
		if *producer.ctx.sendCheckCountInc == 0 {

			// Reset the counter back to the check interval so that we wait until
			// the necessary number of messages have been sent before running the
			// next error check.
			*producer.ctx.sendCheckCountInc = producer.ctx.sendCheckCount

			// Invoke the Stat call agains the queue manager to check for errors.
			sts := ibmmq.NewMQSTS()
			statErr := producer.ctx.qMgr.Stat(ibmmq.MQSTAT_TYPE_ASYNC_ERROR, sts)

			if statErr != nil {

				// Problem occurred invoking the Stat call, pass this back to
				// the user.
				err = statErr

			} else {

				// If there are any Warnings or Failures then we have found a problem that
				// needs to be reported to the user.
				if sts.PutWarningCount+sts.PutFailureCount > 0 {

					retErr = populateAsyncPutError(sts)

				}

			}

		}

	}

	// If the user is using transactional async-put of persistent messages then we need to
	// inform the ContextImpl object that an async-put message has been sent, so that it can
	// check for failures when the Commit call is made.
	//
	// No error checks are made for non-persistent async put messages under a transaction,
	// and the application does not receive any feedback whether those messages arrived safely.
	//
	// Note that if there is already an error returned from Put then just pass that back to
	// the user (only go into this if err is nil).
	if dest.GetPutAsyncAllowed() == jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED &&
		syncpointSetting == ibmmq.MQPMO_SYNCPOINT &&
		putmqmd.Persistence == ibmmq.MQPER_PERSISTENT &&
		*producer.ctx.sendCheckCountInc != ContextImpl_TRANSACTED_ASYNCPUT_ACTIVE &&
		err == nil {

		// Set the flag to indicate the a transacted async put has taken place.
		*producer.ctx.sendCheckCountInc = ContextImpl_TRANSACTED_ASYNCPUT_ACTIVE
	}

	// Note that the following block handles errors for both opening the queue
	// and putting the message.
	if err != nil {

		rcInt := int(err.(*ibmmq.MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := ibmmq.MQItoString("RC", rcInt)
		retErr = jms20subset.CreateJMSException(reason, errCode, err)

	}

	return retErr

}

// populateAsyncPutError is a common function used in several places to generate a
// consistent error message in response to failures during asynchronous put operations.
func populateAsyncPutError(sts *ibmmq.MQSTS) jms20subset.JMSException {

	// sts.Reason contains the detail of the first failure
	errCode2 := strconv.Itoa(int(sts.CompCode))
	reason2 := ibmmq.MQItoString("RC", int(sts.Reason))
	linkedErr := jms20subset.CreateJMSException(reason2, errCode2, nil)

	// Create an error that describes what has failed.
	reason := fmt.Sprintf("%d failures and %d warnings for asynchronous message put", sts.PutFailureCount, sts.PutWarningCount)
	errCode := "AsyncPutFailure"
	return jms20subset.CreateJMSException(reason, errCode, linkedErr)

}

// SetDeliveryMode contains the MQ logic necessary to store the specified
// delivery mode parameter inside the Producer object so that it can be
// applied when sending messages using this Producer.
func (producer *ProducerImpl) SetDeliveryMode(mode int) jms20subset.JMSProducer {

	// Check that the specified mode parameter is one of the values that we permit,
	// and if so store that value inside producer.
	if mode == jms20subset.DeliveryMode_PERSISTENT || mode == jms20subset.DeliveryMode_NON_PERSISTENT {
		producer.deliveryMode = mode

	} else {
		// Normally we would throw an error here to indicate that an invalid value
		// was specified, however we have decided that it is more useful to support
		// method chaining, which prevents us from returning an error object.
		// Instead we settle for printing an error message to the console.
		fmt.Println("Invalid DeliveryMode specified: " + strconv.Itoa(mode))
	}

	return producer
}

// GetDeliveryMode returns the current delivery mode that is set on this
// Producer.
func (producer *ProducerImpl) GetDeliveryMode() int {
	return producer.deliveryMode
}

// SetTimeToLive contains the MQ logic necessary to store the specified
// time to live parameter inside the Producer object so that it can be
// applied when sending messages using this Producer.
func (producer *ProducerImpl) SetTimeToLive(timeToLive int) jms20subset.JMSProducer {

	// Only accept a non-negative value for time to live.
	if timeToLive >= 0 {
		producer.timeToLive = timeToLive

	} else {
		// Normally we would throw an error here to indicate that an invalid value
		// was specified, however we have decided that it is more useful to support
		// method chaining, which prevents us from returning an error object.
		// Instead we settle for printing an error message to the console.
		fmt.Println("Invalid TimeToLive specified: " + strconv.FormatInt(int64(timeToLive), 10))
	}

	return producer
}

// GetTimeToLive returns the current time to live that is set on this
// Producer.
func (producer *ProducerImpl) GetTimeToLive() int {
	return producer.timeToLive
}

// SetPriority contains the MQ logic necessary to store the specified
// priority parameter inside the Producer object so that it can be
// applied when sending messages using this Producer.
func (producer *ProducerImpl) SetPriority(priority int) jms20subset.JMSProducer {

	// Only accept a non-negative value for priority.
	if priority >= 0 {
		producer.priority = priority

	} else {
		// Normally we would throw an error here to indicate that an invalid value
		// was specified, however we have decided that it is more useful to support
		// method chaining, which prevents us from returning an error object.
		// Instead we settle for printing an error message to the console.
		fmt.Println("Invalid Priority specified: " + strconv.FormatInt(int64(priority), 10))
	}

	return producer
}

// GetPriority returns the priority for all messages sent by this producer.
func (producer *ProducerImpl) GetPriority() int {
	return producer.priority
}
