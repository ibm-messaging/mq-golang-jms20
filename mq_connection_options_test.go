/*
 * Copyright (c) IBM Corporation 2019
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0, which is available at
 * http://www.eclipse.org/legal/epl-2.0.
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package main

import (
	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
	"testing"

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

func TestMQConnectionOptions(t *testing.T) {
	t.Run("MaxMsgLength", func(t *testing.T) {
		// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
		cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
		assert.Nil(t, cfErr)

		// Ensure that the options were applied when setting connection options on Context creation
		msg := "options were not applied"
		context, ctxErr := cf.CreateContext(
			mqjms.WithMaxMsgLength(2000),
			func(cno *ibmmq.MQCNO) {
				assert.Equal(t, 2000, cno.ClientConn.MaxMsgLength)
				msg = "options applied"
			},
		)
		assert.Nil(t, ctxErr)

		if context != nil {
			defer context.Close()
		}

		assert.Equal(t, "options applied", msg)
	})
}
