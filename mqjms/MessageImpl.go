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
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	ibmmq "github.com/ibm-messaging/mq-golang/v5/ibmmq"
)

const MessageImpl_PROPERTY_CONVERT_FAILED_REASON string = "MQJMS_E_BAD_TYPE"
const MessageImpl_PROPERTY_CONVERT_FAILED_CODE string = "1055"
const MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_REASON string = "MQJMS_E_UNSUPPORTED_TYPE"
const MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_CODE string = "1056	"

// MessageImpl contains the IBM MQ specific attributes that are
// common to all types of message.
type MessageImpl struct {
	mqmd      *ibmmq.MQMD
	msgHandle *ibmmq.MQMessageHandle
}

// GetJMSDeliveryMode extracts the persistence setting from this message
// and returns it in the JMS delivery mode format.
func (msg *MessageImpl) GetJMSDeliveryMode() int {

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

// GetJMSPriority extracts the message priority from the native MQ message descriptor.
func (msg *MessageImpl) GetJMSPriority() int {

	pri := 4

	// Extract the MsgId field from the MQ message descriptor if one exists.
	// Note that if there is no MQMD then there is no messageID to return.
	if msg.mqmd != nil {
		pri = int(msg.mqmd.Priority)
	}

	return pri
}

// GetJMSMessageID extracts the message ID from the native MQ message descriptor.
func (msg *MessageImpl) GetJMSMessageID() string {
	msgIDStr := ""

	// Extract the MsgId field from the MQ message descriptor if one exists.
	// Note that if there is no MQMD then there is no messageID to return.
	if msg.mqmd != nil && msg.mqmd.MsgId != nil {
		msgIDBytes := msg.mqmd.MsgId
		msgIDStr = hex.EncodeToString(msgIDBytes)
	}

	return msgIDStr
}

// SetJMSReplyTo uses the specified Destination object to configure the reply
// attributes of the native MQ message fields.
func (msg *MessageImpl) SetJMSReplyTo(dest jms20subset.Destination) jms20subset.JMSException {

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
func (msg *MessageImpl) GetJMSReplyTo() jms20subset.Destination {
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
func (msg *MessageImpl) SetJMSCorrelationID(correlID string) jms20subset.JMSException {

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
func (msg *MessageImpl) GetJMSCorrelationID() string {
	correlID := ""

	// Note that if there is no MQMD then there is no correlID stored.
	if msg.mqmd != nil && msg.mqmd.CorrelId != nil {

		// Get hold of the bytes representation of the correlation ID.
		correlIDBytes := msg.mqmd.CorrelId

		// We want to be able to give back the same content the application
		// originally gave us, which could either be an encoded set of bytes, or
		// alternative a plain text string.
		// Here we identify any padding zero bytes to trim off so that we can try
		// to turn it back into a string.
		realLength := len(correlIDBytes)
		for realLength > 0 && correlIDBytes[realLength-1] == 0 {
			realLength--
		}

		// Attempt to decode the content back into a string.
		dst := make([]byte, hex.DecodedLen(realLength))
		n, err := hex.Decode(dst, correlIDBytes[0:realLength])

		if err == nil {
			// The decode back to a string was successful so pass back that plain
			// text string to the caller.
			correlID = string(dst[:n])

		} else {

			// An error occurred while decoding to a plain text string, so encode
			// the bytes that we have into a raw string representation themselves.
			correlID = hex.EncodeToString(correlIDBytes)
		}

	}

	return correlID
}

// GetJMSTimestamp retrieves the timestamp at which the message was sent from
// the native MQ message descriptor fields.
func (msg *MessageImpl) GetJMSTimestamp() int64 {

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

// GetJMSExpiration returns the timestamp at which the message is due to
// expire.
func (msg *MessageImpl) GetJMSExpiration() int64 {

	// mqmd.Expiry gives tenths of a second after which message should expire
	timestamp := msg.GetJMSTimestamp()

	if timestamp != 0 && msg.mqmd.Expiry != 0 {
		timestamp += (int64(msg.mqmd.Expiry) * 100)
	}

	return timestamp
}

// SetStringProperty enables an application to set a string-type message property.
//
// value is *string which allows a nil value to be specified, to unset an individual
// property.
func (msg *MessageImpl) SetStringProperty(name string, value *string) jms20subset.JMSException {
	var retErr jms20subset.JMSException

	var linkedErr error

	// Different code path and shortcut for special header properties
	isSpecial, specialErr := msg.setSpecialStringPropertyValue(name, value)
	if isSpecial {

		if specialErr != nil {
			retErr = jms20subset.CreateJMSException("4125", "MQJMS4125", specialErr)
		}
		return retErr
	}

	if value != nil {
		// Looking to set a value
		var valueStr string
		valueStr = *value

		smpo := ibmmq.NewMQSMPO()
		pd := ibmmq.NewMQPD()

		linkedErr = msg.msgHandle.SetMP(smpo, name, pd, valueStr)
	} else {
		// Looking to unset a value
		dmpo := ibmmq.NewMQDMPO()

		linkedErr = msg.msgHandle.DltMP(dmpo, name)
	}

	if linkedErr != nil {
		rcInt := int(linkedErr.(*ibmmq.MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := ibmmq.MQItoString("RC", rcInt)
		retErr = jms20subset.CreateJMSException(reason, errCode, linkedErr)
	}

	return retErr
}

// setSpecialStringPropertyValue sets the special header properties that are of type String
func (msg *MessageImpl) setSpecialStringPropertyValue(name string, value *string) (bool, error) {

	// Special properties always start with a known prefix.
	if !strings.HasPrefix(name, "JMS") {
		return false, nil
	}

	// Check first that there is an MQMD to write to
	if msg.mqmd == nil {
		msg.mqmd = ibmmq.NewMQMD()
	}

	// Assume for now that this property is special as it has passed the basic
	// checks, and this value will be set back to false if it doesn't match any
	// of the specific fields.
	isSpecial := true

	var err error

	switch name {
	case "JMS_IBM_Format":
		if value != nil {
			msg.mqmd.Format = *value
		} else {
			msg.mqmd.Format = ibmmq.MQFMT_NONE // unset
		}

	case "JMS_IBM_MQMD_Format":
		if value != nil {
			msg.mqmd.Format = *value
		} else {
			msg.mqmd.Format = ibmmq.MQFMT_NONE // unset
		}

	case "JMSXGroupID":
		err = errors.New("Not yet implemented")
		/* Implementation not yet complete
		if value != nil {
			groupBytes := convertStringToMQBytes(*value)
			msg.mqmd.GroupId = groupBytes
			msg.mqmd.MsgFlags |= ibmmq.MQMF_MSG_IN_GROUP
		} */

	default:
		isSpecial = false
	}

	return isSpecial, err
}

// setSpecialIntPropertyValue sets the special header properties of type int
func (msg *MessageImpl) setSpecialIntPropertyValue(name string, value int) (bool, error) {

	// Special properties always start with a known prefix.
	if !strings.HasPrefix(name, "JMS") {
		return false, nil
	}

	// Check first that there is an MQMD to write to
	if msg.mqmd == nil {
		msg.mqmd = ibmmq.NewMQMD()
	}

	// Assume for now that this property is special as it has passed the basic
	// checks, and this value will be set back to false if it doesn't match any
	// of the specific fields.
	isSpecial := true

	var err error

	switch name {
	case "JMS_IBM_PutApplType":
		msg.mqmd.PutApplType = int32(value)

	case "JMS_IBM_Encoding":
		msg.mqmd.Encoding = int32(value)

	case "JMS_IBM_Character_Set":
		msg.mqmd.CodedCharSetId = int32(value)

	case "JMS_IBM_MQMD_CodedCharSetId":
		msg.mqmd.CodedCharSetId = int32(value)

	case "JMS_IBM_MsgType":
		msg.mqmd.MsgType = int32(value)

	case "JMS_IBM_MQMD_MsgType":
		msg.mqmd.MsgType = int32(value)

	case "JMSXGroupSeq":
		err = errors.New("Not yet implemented")
		//msg.mqmd.MsgSeqNumber = int32(value)

	default:
		isSpecial = false
	}

	return isSpecial, err
}

// getSpecialPropertyValue returns the value of special header properties such as
// values from the MQMD that are mapped to JMS properties.
func (msg *MessageImpl) getSpecialPropertyValue(name string) (bool, interface{}, error) {

	// Special properties always start with a known prefix.
	if !strings.HasPrefix(name, "JMS") {
		return false, nil, nil
	}

	// Assume for now that this property is special as it has passed the basic
	// checks, and this value will be set back to false if it doesn't match any
	// of the specific fields.
	isSpecial := true

	var value interface{}
	var err error

	switch name {
	case "JMS_IBM_PutDate":
		if msg.mqmd != nil {
			value = msg.mqmd.PutDate
		}

	case "JMS_IBM_PutTime":
		if msg.mqmd != nil {
			value = msg.mqmd.PutTime
		}

	case "JMS_IBM_Format":
		if msg.mqmd != nil && msg.mqmd.Format != ibmmq.MQFMT_NONE {
			value = msg.mqmd.Format
		}

	case "JMS_IBM_MQMD_Format": // same as JMS_IBM_Format
		if msg.mqmd != nil && msg.mqmd.Format != ibmmq.MQFMT_NONE {
			value = msg.mqmd.Format
		}

	case "JMSXAppID":
		if msg.mqmd != nil {
			value = msg.mqmd.PutApplName
		}

	case "JMS_IBM_MQMD_ApplOriginData":
		if msg.mqmd != nil && msg.mqmd.ApplOriginData != ibmmq.MQFMT_NONE {
			value = msg.mqmd.ApplOriginData
		}

	case "JMS_IBM_PutApplType":
		if msg.mqmd != nil {
			value = msg.mqmd.PutApplType
		}

	case "JMS_IBM_Encoding":
		if msg.mqmd != nil {
			value = msg.mqmd.Encoding
		}

	case "JMS_IBM_Character_Set":
		if msg.mqmd != nil {
			value = msg.mqmd.CodedCharSetId
		}

	case "JMS_IBM_MQMD_CodedCharSetId":
		if msg.mqmd != nil {
			value = msg.mqmd.CodedCharSetId
		}

	case "JMS_IBM_MsgType":
		if msg.mqmd != nil {
			value = msg.mqmd.MsgType
		}

	case "JMS_IBM_MQMD_MsgType":
		if msg.mqmd != nil {
			value = msg.mqmd.MsgType
		}

	case "JMSXGroupID":
		if msg.mqmd != nil {
			valueBytes := msg.mqmd.GroupId

			// See whether this is a non-zero response.
			nonZeros := false
			for _, thisByte := range valueBytes {
				if thisByte != 0 {
					nonZeros = true
					break
				}
			}

			if nonZeros {
				value = hex.EncodeToString(valueBytes)
			}
		}

	case "JMSXGroupSeq":
		if msg.mqmd != nil {
			value = msg.mqmd.MsgSeqNumber
		} else {
			value = int32(1)
		}

	case "JMS_IBM_Last_Msg_In_Group":
		if msg.mqmd != nil {
			value = ((msg.mqmd.MsgFlags & ibmmq.MQMF_LAST_MSG_IN_GROUP) != 0)
		} else {
			value = false
		}

	default:
		isSpecial = false
	}

	return isSpecial, value, err
}

// GetStringProperty returns the string value of a named message property.
// Returns nil if the named property is not set.
func (msg *MessageImpl) GetStringProperty(name string) (*string, jms20subset.JMSException) {

	var valueStrPtr *string
	var retErr jms20subset.JMSException

	impo := ibmmq.NewMQIMPO()
	pd := ibmmq.NewMQPD()

	// Check first if this is a special property
	isSpecialProp, value, err := msg.getSpecialPropertyValue(name)

	if !isSpecialProp {
		// If not then look for a user property
		_, value, err = msg.msgHandle.InqMP(impo, pd, name)
	}

	if err == nil {

		var parseErr error

		if value != nil {

			switch valueTyped := value.(type) {
			case string:
				valueStrPtr = &valueTyped
			case int64:
				valueStr := strconv.FormatInt(valueTyped, 10)
				valueStrPtr = &valueStr
				if parseErr != nil {
					retErr = jms20subset.CreateJMSException(MessageImpl_PROPERTY_CONVERT_FAILED_REASON,
						MessageImpl_PROPERTY_CONVERT_FAILED_CODE, parseErr)
				}
			case bool:
				valueStr := strconv.FormatBool(valueTyped)
				valueStrPtr = &valueStr
			case float64:
				valueStr := fmt.Sprintf("%g", valueTyped)
				valueStrPtr = &valueStr
			default:
				retErr = jms20subset.CreateJMSException(MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_REASON,
					MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_CODE, parseErr)
			}

		}

	} else {

		mqret := err.(*ibmmq.MQReturn)
		if mqret.MQRC == ibmmq.MQRC_PROPERTY_NOT_AVAILABLE {
			// This indicates that the requested property does not exist.
			// valueStr will remain with its default value
			return nil, nil
		} else {
			// Err was not nil
			rcInt := int(mqret.MQRC)
			errCode := strconv.Itoa(rcInt)
			reason := ibmmq.MQItoString("RC", rcInt)
			retErr = jms20subset.CreateJMSException(reason, errCode, mqret)

			valueStrPtr = nil
		}
	}
	return valueStrPtr, retErr
}

// SetIntProperty enables an application to set a int-type message property.
func (msg *MessageImpl) SetIntProperty(name string, value int) jms20subset.JMSException {
	var retErr jms20subset.JMSException

	var linkedErr error

	// Different code path and shortcut for special header properties
	isSpecial, specialErr := msg.setSpecialIntPropertyValue(name, value)
	if isSpecial {

		if specialErr != nil {
			retErr = jms20subset.CreateJMSException("4125", "MQJMS4125", specialErr)
		}
		return retErr
	}

	smpo := ibmmq.NewMQSMPO()
	pd := ibmmq.NewMQPD()

	linkedErr = msg.msgHandle.SetMP(smpo, name, pd, value)

	if linkedErr != nil {
		rcInt := int(linkedErr.(*ibmmq.MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := ibmmq.MQItoString("RC", rcInt)
		retErr = jms20subset.CreateJMSException(reason, errCode, linkedErr)
	}

	return retErr
}

// GetIntProperty returns the int value of a named message property.
// Returns 0 if the named property is not set.
func (msg *MessageImpl) GetIntProperty(name string) (int, jms20subset.JMSException) {

	var valueRet int
	var retErr jms20subset.JMSException

	impo := ibmmq.NewMQIMPO()
	pd := ibmmq.NewMQPD()

	// Check first if this is a special property
	isSpecialProp, value, err := msg.getSpecialPropertyValue(name)

	if !isSpecialProp {
		// If not then look for a user property
		_, value, err = msg.msgHandle.InqMP(impo, pd, name)
	}

	if err == nil {

		var parseErr error

		switch valueTyped := value.(type) {
		case int:
			valueRet = valueTyped
		case int32:
			valueRet = int(valueTyped)
		case int64:
			valueRet = int(valueTyped)
		case string:
			valueRet, parseErr = strconv.Atoi(valueTyped)
		case bool:
			if valueTyped {
				valueRet = 1
			}
		case float64:
			s := fmt.Sprintf("%.0f", valueTyped)
			valueRet, parseErr = strconv.Atoi(s)
		default:
			retErr = jms20subset.CreateJMSException(MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_REASON,
				MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_CODE, parseErr)
		}

		if parseErr != nil {
			retErr = jms20subset.CreateJMSException(MessageImpl_PROPERTY_CONVERT_FAILED_REASON,
				MessageImpl_PROPERTY_CONVERT_FAILED_CODE, parseErr)
		}

	} else {

		mqret := err.(*ibmmq.MQReturn)
		if mqret.MQRC == ibmmq.MQRC_PROPERTY_NOT_AVAILABLE {
			// This indicates that the requested property does not exist.
			// valueRet will remain with its default value
			return 0, nil
		} else {
			// Err was not nil
			rcInt := int(mqret.MQRC)
			errCode := strconv.Itoa(rcInt)
			reason := ibmmq.MQItoString("RC", rcInt)
			retErr = jms20subset.CreateJMSException(reason, errCode, mqret)
		}
	}
	return valueRet, retErr
}

// SetDoubleProperty enables an application to set a double-type (float64) message property.
func (msg *MessageImpl) SetDoubleProperty(name string, value float64) jms20subset.JMSException {
	var retErr jms20subset.JMSException

	var linkedErr error

	smpo := ibmmq.NewMQSMPO()
	pd := ibmmq.NewMQPD()

	linkedErr = msg.msgHandle.SetMP(smpo, name, pd, value)

	if linkedErr != nil {
		rcInt := int(linkedErr.(*ibmmq.MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := ibmmq.MQItoString("RC", rcInt)
		retErr = jms20subset.CreateJMSException(reason, errCode, linkedErr)
	}

	return retErr
}

// GetDoubleProperty returns the double (float64) value of a named message property.
// Returns 0 if the named property is not set.
func (msg *MessageImpl) GetDoubleProperty(name string) (float64, jms20subset.JMSException) {

	var valueRet float64
	var retErr jms20subset.JMSException

	impo := ibmmq.NewMQIMPO()
	pd := ibmmq.NewMQPD()

	// Check first if this is a special property
	isSpecialProp, value, err := msg.getSpecialPropertyValue(name)

	if !isSpecialProp {
		// If not then look for a user property
		_, value, err = msg.msgHandle.InqMP(impo, pd, name)
	}

	if err == nil {

		var parseErr error

		switch valueTyped := value.(type) {
		case float64:
			valueRet = valueTyped
		case string:
			valueRet, parseErr = strconv.ParseFloat(valueTyped, 64)
			if parseErr != nil {
				retErr = jms20subset.CreateJMSException(MessageImpl_PROPERTY_CONVERT_FAILED_REASON,
					MessageImpl_PROPERTY_CONVERT_FAILED_CODE, parseErr)
			}
		case int64:
			valueRet = float64(valueTyped)
		case bool:
			if valueTyped {
				valueRet = 1
			}
		default:
			retErr = jms20subset.CreateJMSException(MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_REASON,
				MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_CODE, parseErr)
		}
	} else {

		mqret := err.(*ibmmq.MQReturn)
		if mqret.MQRC == ibmmq.MQRC_PROPERTY_NOT_AVAILABLE {
			// This indicates that the requested property does not exist.
			// valueRet will remain with its default value
			return 0, nil
		} else {
			// Err was not nil
			rcInt := int(mqret.MQRC)
			errCode := strconv.Itoa(rcInt)
			reason := ibmmq.MQItoString("RC", rcInt)
			retErr = jms20subset.CreateJMSException(reason, errCode, mqret)
		}
	}
	return valueRet, retErr
}

// SetBooleanProperty enables an application to set a bool-type message property.
func (msg *MessageImpl) SetBooleanProperty(name string, value bool) jms20subset.JMSException {
	var retErr jms20subset.JMSException

	var linkedErr error

	smpo := ibmmq.NewMQSMPO()
	pd := ibmmq.NewMQPD()

	// Different code path and shortcut for special header properties
	isSpecial, specialErr := msg.setSpecialBooleanPropertyValue(name, value)
	if isSpecial {

		if specialErr != nil {
			retErr = jms20subset.CreateJMSException("4125", "MQJMS4125", specialErr)
		}
		return retErr
	}

	linkedErr = msg.msgHandle.SetMP(smpo, name, pd, value)

	if linkedErr != nil {
		rcInt := int(linkedErr.(*ibmmq.MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := ibmmq.MQItoString("RC", rcInt)
		retErr = jms20subset.CreateJMSException(reason, errCode, linkedErr)
	}

	return retErr
}

// setSpecialBooleanPropertyValue sets the special header properties of type bool
func (msg *MessageImpl) setSpecialBooleanPropertyValue(name string, value bool) (bool, error) {

	// Special properties always start with a known prefix.
	if !strings.HasPrefix(name, "JMS") {
		return false, nil
	}

	// Check first that there is an MQMD to write to
	if msg.mqmd == nil {
		msg.mqmd = ibmmq.NewMQMD()
	}

	// Assume for now that this property is special as it has passed the basic
	// checks, and this value will be set back to false if it doesn't match any
	// of the specific fields.
	isSpecial := true

	var err error

	switch name {
	case "JMS_IBM_Last_Msg_In_Group":
		err = errors.New("Not yet implemented")

	default:
		isSpecial = false
	}

	return isSpecial, err
}

// GetBooleanProperty returns the bool value of a named message property.
// Returns false if the named property is not set.
func (msg *MessageImpl) GetBooleanProperty(name string) (bool, jms20subset.JMSException) {

	var valueRet bool
	var retErr jms20subset.JMSException

	impo := ibmmq.NewMQIMPO()
	pd := ibmmq.NewMQPD()

	// Check first if this is a special property
	isSpecialProp, value, err := msg.getSpecialPropertyValue(name)

	if !isSpecialProp {
		// If not then look for a user property
		_, value, err = msg.msgHandle.InqMP(impo, pd, name)
	}

	if err == nil {

		var parseErr error

		switch valueTyped := value.(type) {
		case bool:
			valueRet = valueTyped
		case string:
			valueRet, parseErr = strconv.ParseBool(valueTyped)
			if parseErr != nil {
				retErr = jms20subset.CreateJMSException(MessageImpl_PROPERTY_CONVERT_FAILED_REASON,
					MessageImpl_PROPERTY_CONVERT_FAILED_CODE, parseErr)
			}
		case int64:
			// Conversion from int to bool is true iff n=1
			if valueTyped == 1 {
				valueRet = true
			}
		case float64:
			// Conversion from float64 to bool is true iff n=1
			if valueTyped == 1 {
				valueRet = true
			}
		default:
			retErr = jms20subset.CreateJMSException(MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_REASON,
				MessageImpl_PROPERTY_CONVERT_NOTSUPPORTED_CODE, parseErr)
		}
	} else {

		mqret := err.(*ibmmq.MQReturn)
		if mqret.MQRC == ibmmq.MQRC_PROPERTY_NOT_AVAILABLE {
			// This indicates that the requested property does not exist.
			// valueRet will remain with its default value
			return false, nil
		} else {
			// Err was not nil
			rcInt := int(mqret.MQRC)
			errCode := strconv.Itoa(rcInt)
			reason := ibmmq.MQItoString("RC", rcInt)
			retErr = jms20subset.CreateJMSException(reason, errCode, mqret)
		}
	}
	return valueRet, retErr
}

// PropertyExists returns true if the named message property exists on this message.
func (msg *MessageImpl) PropertyExists(name string) (bool, jms20subset.JMSException) {

	found, _, retErr := msg.getPropertiesInternal(name)
	return found, retErr

}

// GetPropertyNames returns a slice of strings containing the name of every message
// property on this message.
// Returns a zero length slice if no message properties are set.
func (msg *MessageImpl) GetPropertyNames() ([]string, jms20subset.JMSException) {

	_, propNames, retErr := msg.getPropertiesInternal("")
	return propNames, retErr
}

// getPropertiesInternal is an internal helper function that provides a largely
// identical implication for two application-facing functions;
// - PropertyExists supplies a non-empty name parameter to check whether that property exists
// - GetPropertyNames supplies an empty name parameter to get a []string of all property names
func (msg *MessageImpl) getPropertiesInternal(name string) (bool, []string, jms20subset.JMSException) {

	impo := ibmmq.NewMQIMPO()
	pd := ibmmq.NewMQPD()
	propNames := []string{}

	impo.Options = ibmmq.MQIMPO_CONVERT_VALUE | ibmmq.MQIMPO_INQ_FIRST
	for propsToRead := true; propsToRead; {

		gotName, _, err := msg.msgHandle.InqMP(impo, pd, "%")
		impo.Options = ibmmq.MQIMPO_CONVERT_VALUE | ibmmq.MQIMPO_INQ_NEXT

		if err != nil {
			mqret := err.(*ibmmq.MQReturn)
			if mqret.MQRC != ibmmq.MQRC_PROPERTY_NOT_AVAILABLE {

				rcInt := int(mqret.MQRC)
				errCode := strconv.Itoa(rcInt)
				reason := ibmmq.MQItoString("RC", rcInt)
				retErr := jms20subset.CreateJMSException(reason, errCode, mqret)
				return false, nil, retErr

			} else {
				// Read all properties (property not available)
				return false, propNames, nil
			}

		} else if "" == name {
			// We are looking to get back a list of all properties
			propNames = append(propNames, gotName)

		} else if gotName == name {
			// We are just checking for the existence of this one property (shortcut)
			return true, nil, nil
		}

	}

	// Went through all properties and didn't find a match
	return false, propNames, nil
}

// ClearProperties removes all message properties from this message.
func (msg *MessageImpl) ClearProperties() jms20subset.JMSException {

	// Get the list of all property names, as we have to delete
	// them individually
	allPropNames, jmsErr := msg.GetPropertyNames()

	if jmsErr == nil {

		dmpo := ibmmq.NewMQDMPO()

		for _, propName := range allPropNames {

			// Delete this property
			err := msg.msgHandle.DltMP(dmpo, propName)

			if err != nil {
				rcInt := int(err.(*ibmmq.MQReturn).MQRC)
				errCode := strconv.Itoa(rcInt)
				reason := ibmmq.MQItoString("RC", rcInt)
				jmsErr = jms20subset.CreateJMSException(reason, errCode, err)
				break
			}
		}

	}

	return jmsErr

}
