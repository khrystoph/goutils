package logUtils

/*
* This utility is for setting up default logging/loggers for API logs through cloudwatch logs, but also
* can log to a file or to standard out/err depending on need.
 */

import (
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	Trace         *log.Logger
	Info          *log.Logger
	Warning       *log.Logger
	Error         *log.Logger
	traceHandle   io.Writer
	infoHandle    io.Writer = os.Stdout
	warningHandle io.Writer = os.Stderr
	errorHandle   io.Writer = os.Stderr
	sess          *session.Session
)

func init() {
	//Set up local logging to stdout/stderr
	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
