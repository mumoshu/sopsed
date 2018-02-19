package app

import (
	"fmt"
	"log"
	"os"
)

// Context contains all the execution context of this app including loggers
type Context struct {
	// TODO Move to a serious leveled-logger
	info    *log.Logger
	err     *log.Logger
	warn    *log.Logger
	debug   *log.Logger
	Verbose bool
}

// NewContext returns a new context with the default loggers
func NewContext() *Context {
	return &Context{
		info:    log.New(os.Stderr, "", 0),
		err:     log.New(os.Stderr, "error: ", 0),
		warn:    log.New(os.Stderr, "warn: ", 0),
		debug:   log.New(os.Stderr, "debug: ", 0),
		Verbose: os.Getenv("DEBUG") != "",
	}
}

// Debug prints a message when the verbose-logging is enabled
func (c *Context) Debug(msg string) {
	if c.Verbose {
		c.debug.Println(msg)
	}
}

// Info prints an info message
func (c *Context) Info(msg string) {
	if c.Verbose {
		c.info.Println(msg)
	}
}

// Warn prints a warn message
func (c *Context) Warn(msg string) {
	if c.Verbose {
		c.warn.Println(msg)
	}
}

// ExitWithError os.Exit with an error
func (c *Context) ExitWithError(err error) {
	c.err.Println(fmt.Sprintf("%v", err))
	os.Exit(1)
}
