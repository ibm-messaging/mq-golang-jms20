// Copyright (c) IBM Corporation 2019.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Implementation of the JMS style Golang interfaces to communicate with IBM MQ.
package mqjms

// Used to configure the TransportType property of the ConnectionFactory,
// to use a TCP client connection to the queue manager. This is the default.
const TransportType_CLIENT int = 0

// Used to configure the TransportType property of the ConnectionFactory,
// to use a local bindings connection to the queue manager
const TransportType_BINDINGS int = 1

// Used to configure the TLSClientAuth property to indicate that a client
// certificate should not be sent.
const TLSClientAuth_NONE string = "NONE"

// Used to configure the TLSClientAuth property to indicate that a client
// certificate must be sent to the queue manager, as part of mutual TLS.
const TLSClientAuth_REQUIRED string = "REQUIRED"
