// Copyright (c) IBM Corporation 2021.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Package main provides the entry point for a executable application.
package main

import (
	"fmt"
	"log"

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
)

func main() {
	fmt.Println("Beginning world!!!")

	cf := mqjms.ConnectionFactoryImpl{
		QMName:      "QM1",
		Hostname:    "myhostname.com",
		PortNumber:  1414,
		ChannelName: "SYSTEM.DEF.SVRCONN",
		UserName:    "username",
		Password:    "password",
	}

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, errCtx := cf.CreateContext()
	if context != nil {
		defer context.Close()
	}

	if errCtx != nil {
		log.Fatal(errCtx)
	}

	fmt.Println("Ending world!!!")
}
