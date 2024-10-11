# gate

## Usage

```
gate version 0.1.0 usage:
NAME
	gate

DESCRIPTION
	gate uses a named fifo to enable one command to gate execution of another.

	The named fifo path is intialized with either the environment variable
	GATE_FIFO or the command-line flag [-f|--fifo].
	It is removed after both the sender and receiver have closed the fifo.

	The sender closes the fifo after writing all data and signals EOF.

	By default, the receiver closes the fifo after reading all data and
	receiving EOF. The receiver can instead keep the fifo open indefinitely
	until it matches a user-provided pattern ("glob" or "/regexp/").
	Matching the pattern will close the fifo after receiving EOF so that
	the sender can finish writing gracefully.

	Currently, only line-buffered matching is supported.

	[NOTE] Command-line flags take precedence over environment variables.

SYNOPSIS
	gate -              (send) stdin => named fifo
	gate                (recv) named fifo => stdout

	gate < file         (send) file => named fifo
	gate > file         (recv) named fifo => file

	cmd | gate          (send) cmd stdout => named fifo
	gate | cmd          (recv) named fifo => cmd stdin

OPTIONS
	-f|--fifo FILE       		file path to named fifo        
	-t|--timeout DURATION		duration to wait for named fifo
	-m|--match PATTERN   		(recv-only) close after pattern

ENVIRONMENT
	GATE_FIFO            		file path to named fifo        
	GATE_TIMEOUT         		duration to wait for named fifo
	GATE_MATCH           		(recv-only) close after pattern
```
