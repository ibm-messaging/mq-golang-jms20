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

// ConnectionFactoryImpl defines a struct that contains attributes for
// each of the key properties required to establish a connection to an IBM MQ
// queue manager.
//
// The fields are defined as Public so that the struct can be initialised
// programmatically using whatever approach the application prefers.
type ConnectionFactoryImpl struct {
	QMName      string
	Hostname    string
	PortNumber  int
	ChannelName string
	UserName    string
	Password    string

	TransportType int // Default to TransportType_CLIENT (0)

	// Equivalent to SSLCipherSpec and SSLClientAuth in the MQI client, however
	// the names have been updated here to reflect that SSL protocols have all
	// been discredited.
	TLSCipherSpec string
	TLSClientAuth string // Default to TLSClientAuth_NONE

	KeyRepository    string
	CertificateLabel string

	// Allthough only available per MQ 9.1.2 it looks like a good idea to have this present in MQ-JMS
	ApplName string

	// Controls the size of the buffer used when receiving a message (default is 32kb if not set)
	ReceiveBufferSize int
}

// CreateContext implements the JMS method to create a connection to an IBM MQ
// queue manager.
func (cf ConnectionFactoryImpl) CreateContext() (jms20subset.JMSContext, jms20subset.JMSException) {
	return cf.CreateContextWithSessionMode(jms20subset.JMSContextAUTOACKNOWLEDGE)
}

// CreateContextWithSessionMode implements the JMS method to create a connection to an IBM MQ
// queue manager using the specified session mode.
func (cf ConnectionFactoryImpl) CreateContextWithSessionMode(sessionMode int) (jms20subset.JMSContext, jms20subset.JMSException) {

	// Allocate the internal structures required to create an connection to IBM MQ.
	cno := ibmmq.NewMQCNO()

	if cf.TransportType == TransportType_CLIENT {

		// Indicate that we want to use a client (TCP) connection.
		cno.Options = ibmmq.MQCNO_CLIENT_BINDING

		// Fill in the required fields in the channel definition structure
		cd := ibmmq.NewMQCD()
		cd.ChannelName = cf.ChannelName
		cd.ConnectionName = cf.Hostname + "(" + strconv.Itoa(cf.PortNumber) + ")"
		cno.ClientConn = cd

		// Fill in the fields relating to TLS channel connections
		if cf.TLSCipherSpec != "" {
			cd.SSLCipherSpec = cf.TLSCipherSpec
		}

		switch cf.TLSClientAuth {
		case TLSClientAuth_REQUIRED:
			cd.SSLClientAuth = ibmmq.MQSCA_REQUIRED
		case TLSClientAuth_NONE:
		case "":
			cd.SSLClientAuth = ibmmq.MQSCA_OPTIONAL
		default:
			cd.SSLClientAuth = -1 // Trigger an error message
		}

		// Set up the reference to the key repository file, if it has been specified.
		if cf.KeyRepository != "" {
			sco := ibmmq.NewMQSCO()
			sco.KeyRepository = cf.KeyRepository

			if cf.CertificateLabel != "" {
				sco.CertificateLabel = cf.CertificateLabel
			}

			cno.SSLConfig = sco

		}

		// Fill in the optional (possible since MQ 9.1.2) application name
		cno.ApplName = cf.ApplName

	} else if cf.TransportType == TransportType_BINDINGS {

		// Indicate to use Bindings connections.
		cno.Options = ibmmq.MQCNO_LOCAL_BINDING

	}

	if cf.UserName != "" {

		// Store the user credentials in an MQCSP, which ensures that long passwords
		// can be used.
		csp := ibmmq.NewMQCSP()
		csp.AuthenticationType = ibmmq.MQCSP_AUTH_USER_ID_AND_PWD
		csp.UserId = cf.UserName
		csp.Password = cf.Password
		cno.SecurityParms = csp

	}

	var ctx jms20subset.JMSContext
	var retErr jms20subset.JMSException

	// Use the objects that we have configured to create a connection to the
	// queue manager.
	qMgr, err := ibmmq.Connx(cf.QMName, cno)

	if err == nil {

		// Connection was created successfully, so we wrap the MQI object into
		// a new ContextImpl and return it to the caller.
		ctx = ContextImpl{
			qMgr:              qMgr,
			sessionMode:       sessionMode,
			receiveBufferSize: cf.ReceiveBufferSize,
		}

	} else {

		// The underlying MQI call returned an error, so extract the relevant
		// details and pass it back to the caller as a JMSException
		rcInt := int(err.(*ibmmq.MQReturn).MQRC)
		errCode := strconv.Itoa(rcInt)
		reason := ibmmq.MQItoString("RC", rcInt)
		retErr = jms20subset.CreateJMSException(reason, errCode, err)

	}

	return ctx, retErr

}
