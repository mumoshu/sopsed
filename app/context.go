package main

import (
	"log"
	"os"
)

// Context contains all the execution context of this app including loggers
type Context struct {
	info *log.Logger
	err  *log.Logger
}

// NewContext returns a new context with the default loggers
func NewContext() *Context {
	return &Context{
		info: log.New(os.Stderr, "", 0),
		err:  log.New(os.Stderr, "error: ", 0),
	}
}
