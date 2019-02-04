// Derived from the Eclipse Project for JMS, available at;
//     https://github.com/eclipse-ee4j/jms-api
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

//
package jms20subset

// Go doesn't allow constants in structs so the naming of this file is only for
// logical grouping purposes. The constants are package scoped, but we use a
// prefix to the naming in order to maintain similarity with Java JMS.

// Used to configure messages to be sent Persistently.
const DeliveryMode_PERSISTENT int = 2

// Used to configure messages to be sent Non-Persistently.
const DeliveryMode_NON_PERSISTENT int = 1
