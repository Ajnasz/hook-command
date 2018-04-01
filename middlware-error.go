package main

import (
	log "github.com/Sirupsen/logrus"
)

// MiddlewareError middleware error struct
type MiddlewareError struct {
	Code      int
	Text      string
	LogFields log.Fields
	Msg       string
}
