// Copyright (c) IBM Corporation 2019.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

//
package mqjms

import (
	"encoding/hex"
	"fmt"
	"github.com/ibm-messaging/mq-golang/ibmmq"
	"github.com/matscus/mq-golang-jms20/jms20subset"
	"log"
	"strconv"
	"strings"
	"time"
)

// TextMessageImpl contains the IBM MQ specific attributes necessary to
// present a message that carries a string.
type TextMessageImpl struct {
	bodyStr *string
	mqmd    *ibmmq.MQMD
}

// GetText returns the string that is contained in this TextMessage.
func (msg *TextMessageImpl) GetText() *string {

	return msg.bodyStr

}

// SetText stores the supplied string so that it can be transmitted as part
// of this TextMessage.
func (msg *TextMessageImpl) SetText(newBody string) {

	msg.bodyStr = &newBody

}

// GetJMSDeliveryMode extracts the persistence setting from this message
// and returns it in the JMS delivery mode format.
func (msg *TextMessageImpl) GetJMSDeliveryMode() int {

	// Retrieve the MQ persistence value from the MQ message descriptor.
	mqMsgPersistence := msg.mqmd.Persistence
	var jmsPersistence int

	// Convert the MQ persistence value to the JMS delivery mode value.
	if mqMsgPersistence == ibmmq.MQPER_NOT_PERSISTENT {
		jmsPersistence = jms20subset.DeliveryMode_NON_PERSISTENT
	} else if mqMsgPersistence == ibmmq.MQPER_PERSISTENT {
		jmsPersistence = jms20subset.DeliveryMode_PERSISTENT
	} else {
		// Give some indication if we received something we didn't expect.
		fmt.Println("Unexpected persistence value: " + strconv.Itoa(int(mqMsgPersistence)))
	}

	return jmsPersistence
}

// GetJMSMessageID extracts the message ID from the native MQ message descriptor.
func (msg *TextMessageImpl) GetJMSMessageID() string {
	msgIdStr := ""

	// Extract the MsgId field from the MQ message descriptor if one exists.
	// Note that if there is no MQMD then there is no messageID to return.
	if msg.mqmd != nil && msg.mqmd.MsgId != nil {
		msgIdBytes := msg.mqmd.MsgId
		msgIdStr = hex.EncodeToString(msgIdBytes)
	}

	return msgIdStr
}

// SetJMSReplyTo uses the specified Destination object to configure the reply
// attributes of the native MQ message fields.
func (msg *TextMessageImpl) SetJMSReplyTo(dest jms20subset.Destination) jms20subset.JMSException {

	switch typedDest := dest.(type) {
	case QueueImpl:

		// Reply information is stored in the MQ message descriptor, so we need to
		// add one to this message if it doesn't already exist.
		if msg.mqmd == nil {
			msg.mqmd = ibmmq.NewMQMD()
		}

		// Save the queue information into the MQMD so that it can be transmitted.
		msg.mqmd.ReplyToQ = typedDest.queueName

	default:
		// This "should never happen"(!) apart from in situations where we are
		// part way through adding support for a new destination type to this library.
		log.Fatal(jms20subset.CreateJMSException("UnexpectedDestinationType", "UnexpectedDestinationType", nil))
	}

  // The option to return an error is not currently used.
	return nil
}

// GetJMSReplyTo extracts the native reply information from the MQ message
// and populates it into a Destination object.
func (msg *TextMessageImpl) GetJMSReplyTo() jms20subset.Destination {
	var replyDest jms20subset.Destination
	replyDest = nil

	// Extract the reply information from the native MQ message descriptor.
	// Note that if this message doesn't have an MQMD then there is no reply
	// destination.
	if msg.mqmd != nil && msg.mqmd.ReplyToQ != "" {
		replyQ := strings.TrimSpace(msg.mqmd.ReplyToQ)

		// Create the Destination object and populate it to be returned.
		replyDest = QueueImpl{
			queueName: replyQ,
		}
	}

	return replyDest
}

// SetJMSCorrelationID applies the specified correlation ID string to the native
// MQ message field used for correlation purposes.
func (msg *TextMessageImpl) SetJMSCorrelationID(correlID string) jms20subset.JMSException {

	var retErr jms20subset.JMSException

	// correlID could either be plain text "myCorrel" or hex encoded bytes "01020304..."
	correlHexBytes := convertStringToMQBytes(correlID)

	// The CorrelID is carried in the MQ message descriptor, so if there isn't
	// one already associated with this message then we need to create one.
	if msg.mqmd == nil {
		msg.mqmd = ibmmq.NewMQMD()
	}

	// Store the bytes form of the correlID
	msg.mqmd.CorrelId = correlHexBytes

	return retErr
}

// Convert a string which is either plain text or an hex encoded strings of bytes
// into an array of bytes that can be used in MQ message descriptors.
func convertStringToMQBytes(strText string) []byte {

	// First try to decode the hex string
	correlHexBytes, err := hex.DecodeString(strText)

	if err != nil {
		// Failed to decode hex string, so assume it is plain text and hex encode it
		// into bytes.
		correlBytes := []byte(strText)
		encodedLen := hex.EncodedLen(len(correlBytes))
		if encodedLen < 24 {
			encodedLen = 24
		}
		correlHexBytes = make([]byte, encodedLen)
		hex.Encode(correlHexBytes, correlBytes)
	}

	// Make sure we don't try to store more bytes than MQ is expecting.
	if len(correlHexBytes) > 48 {
		correlHexBytes = correlHexBytes[0:48]
	}

	return correlHexBytes

}

// GetJMSCorrelationID retrieves the correl ID from the native MQ message
// descriptor field.
func (msg *TextMessageImpl) GetJMSCorrelationID() string {
	correlID := ""

	// Note that if there is no MQMD then there is no correlID stored.
	if msg.mqmd != nil && msg.mqmd.CorrelId != nil {

		// Get hold of the bytes representation of the correlation ID.
		correlIdBytes := msg.mqmd.CorrelId

		// We want to be able to give back the same content the application
		// originally gave us, which could either be an encoded set of bytes, or
		// alternative a plain text string.
		// Here we identify any padding zero bytes to trim off so that we can try
		// to turn it back into a string.
		realLength := len(correlIdBytes)
		if realLength > 0 {
			for correlIdBytes[realLength-1] == 0 {
				realLength--
			}
		}

		// Attempt to decode the content back into a string.
		dst := make([]byte, hex.DecodedLen(realLength))
		n, err := hex.Decode(dst, correlIdBytes[0:realLength])

		if err == nil {
			// The decode back to a string was successful so pass back that plain
			// text string to the caller.
			correlID = string(dst[:n])

		} else {

			// An error occurred while decoding to a plain text string, so encode
			// the bytes that we have into a raw string representation themselves.
			correlID = hex.EncodeToString(correlIdBytes)
		}

	}

	return correlID
}

// GetJMSTimestamp retrieves the timestamp at which the message was sent from
// the native MQ message descriptor fields.
func (msg *TextMessageImpl) GetJMSTimestamp() int64 {

	// Details on the format for the MQMD PutDate and PutTime are defined here;
	// https://www.ibm.com/support/knowledgecenter/en/SSFKSJ_9.0.0/com.ibm.mq.ref.dev.doc/q097650_.html
	// PutDate    YYYYMMDD
	// PutTime    HHMMSSTH (GMT)

	timestamp := int64(0)

	// Note that if there is no MQMD then there is no stored timestamp.
	if msg.mqmd != nil && msg.mqmd.PutDate != "" {

		// Extract the year, month and day segments from the PutDate
		dateStr := msg.mqmd.PutDate
		yearStr := dateStr[0:4]
		monthStr := dateStr[4:6]
		dayStr := dateStr[6:8]

		hourStr := "0"
		minStr := "0"
		secStr := "0"
		millisStr := "0"

		// If a PutTime is specified then extract the pieces of that time as well.
		if msg.mqmd.PutTime != "" {
			timeStr := msg.mqmd.PutTime
			hourStr = timeStr[0:2]
			minStr = timeStr[2:4]
			secStr = timeStr[4:6]

			// The MQMD time format only gives hundredths of second, so add an extra
			// digit to make millis.
			// On average picking "5" will be more accurate than "0" as it is in the
			// middle of the possible range of real values.
			millisStr = timeStr[6:8] + "5"
		}

		// Turn the string representations into numeric variables.
		yearNum, _ := strconv.Atoi(yearStr)
		monthNum, _ := strconv.Atoi(monthStr)
		dayNum, _ := strconv.Atoi(dayStr)
		hourNum, _ := strconv.Atoi(hourStr)
		minNum, _ := strconv.Atoi(minStr)
		secNum, _ := strconv.Atoi(secStr)
		nanosNum, _ := strconv.Atoi(millisStr + "000000")

		// Populate a Date object based on the individual parts, and turn it into a
		// milliseconds-since-Epoch format, which is what is returned by this method.
		timestampObj := time.Date(yearNum, time.Month(monthNum), dayNum, hourNum, minNum, secNum, nanosNum, time.UTC)
		timestamp = timestampObj.UnixNano() / 1000000
	}

	return timestamp
}
