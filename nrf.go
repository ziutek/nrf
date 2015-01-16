package nrf

// Driver contains methods that are need to communicate with nRF24L01(+)
// transceiver (eg. perform SPI conversation and enable its RF part).
type Driver interface {
	// WriteRead perform SPI conversation (see bitbang/spi package).
	WriteRead(oi ...[]byte) (n int, err error)
	
	// Set CE line. v==0 sets CE low, v==1 sets CE high, v==2 pulses
	// CE high for 10 µs and leaves it low.
	SetCE(v int) error
}

// Device wraps driver to provide interface to nRF24L01(+) transceiver.
type Device struct {
	Driver
	
	// Err is error value of last executed command. You can freely call many
	// command methods before check error. If one command return an error
	// subsequent commands are not executed.
	Err error
	
	// Status is value of status register read by last executed command.
	Status
}