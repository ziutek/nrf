package nrf

// Driver contains methods that are need to communicate with nRF24L01(+)
// transceiver (eg. perform SPI conversation and enable its RF part).
type Driver interface {
	// WriteRead perform SPI conversation (see bitbang/spi package).
	WriteRead(oi ...[]byte) (n int, err error)
	// Set CE line.
	SetCE(bool) error
}

// Device wraps driver to provide interface to nRF24L01(+) transceiver.
type Device struct {
	Driver
}
