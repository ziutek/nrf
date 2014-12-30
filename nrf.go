package nrf

import (
	"fmt"
)

type Driver interface {
	// Perform SPI conversation.
	WriteRead(out, in []byte) (int, error)
	// Set CE line.
	Enable(bool) error
}

type Device struct {
	Driver
}

type Status byte

const (
	TxFull Status = 1 << iota // Tx FIFO full flag.
	_
	_
	_
	MaxRT // Maximum number of Tx retransmits interrupt.
	TxDS  // Data Sent Tx FIFO interrupt.
	RxDR  // Data Ready Rx FIFO interrupt.
)

// RxPipe returns data pipe number for the payload available for reading from
// RxFifo or -1 if RxFifo is empty
func (s Status) RxPipe() int {
	n := int(s) & 0x0e
	if n == 0x0e {
		return -1
	}
	return n >> 1
}

func flag(s Status) byte {
	if s == 0 {
		return '-'
	}
	return '+'
}

func (s Status) String() string {
	return fmt.Sprintf(
		"TxFull%c RxPipe:%d MaxRT%c TxDS%c RxDR%c",
		flag(s&TxFull), s.RxPipe(), flag(s&MaxRT), flag(s&TxDS), flag(s&RxDR),
	)
}

func (d Device) NOP() (Status, error) {
	buf := []byte{0xff}
	_, err := d.WriteRead(buf, buf)
	return Status(buf[0]), err
}
