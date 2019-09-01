// Copyright (c) IBM Corporation 2019.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Implementation of the JMS style Golang interfaces to communicate with IBM MQ.
package mqjms

import (
	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang/ibmmq"
	"strconv"
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
}

// CreateContext implements the JMS method to create a connection to an IBM MQ
// queue manager.
func (cf ConnectionFactoryImpl) CreateContext() (jms20subset.JMSContext, jms20subset.JMSException) {

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
			qMgr: qMgr,
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
