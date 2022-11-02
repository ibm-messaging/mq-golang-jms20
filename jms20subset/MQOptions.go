package jms20subset

import "github.com/ibm-messaging/mq-golang/v5/ibmmq"

type MQOptions func(cno *ibmmq.MQCNO)

func WithMaxMsgLength(maxMsgLength int32) MQOptions {
	return func(cno *ibmmq.MQCNO) {
		cno.ClientConn.MaxMsgLength = maxMsgLength
	}
}
