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
}
