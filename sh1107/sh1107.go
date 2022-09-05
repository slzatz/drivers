package sh1107

import (
	"errors"
	"fmt"
	"image/color"
	"time"

	"tinygo.org/x/drivers"
)

const (
	SET_CONTRAST        = 0x81
	SET_ENTIRE_ON       = 0xA4
	SET_NORM_INV        = 0xA6
	SET_DISP            = 0xAE
	SET_DCDC_MODE       = 0xAD
	SET_MEM_MODE        = 0x20
	SET_PAGE_ADDR       = 0xB0
	SET_COL_LO_ADDR     = 0x00
	SET_COL_HI_ADDR     = 0x10
	SET_DISP_START_LINE = 0xDC
	SET_SEG_REMAP       = 0xA0
	SET_MUX_RATIO       = 0xA8
	SET_COM_OUT_DIR     = 0xC0
	SET_DISP_OFFSET     = 0xD3
	SET_DISP_CLK_DIV    = 0xD5
	SET_PRECHARGE       = 0xD9
	SET_VCOM_DESEL      = 0xDB

	TEST_CHUNK = 8

	ADDRESS_128_64 = 0x3C
)

/*
Note that the ssd1306 driver does not use the 'builtin' i2c driver Tx command
Note my Trellis driver does not use write register
*/

// Device wraps I2C or SPI connection.
type Device struct {
	//bus        Buser
	bus        drivers.I2C
	buffer     []byte
	width      int16
	height     int16
	bufferSize int16
	//vccState   VccMode
	//canReset   bool
	pageMode    bool
	pages       int16
	lineBytes   int16
	size        int16
	externalVCC bool
	address     uint16
	//currBuffer []byte
	//prevBuffer []byte
}

/*
// NewI2C creates a new sh1107. The I2C wire must already be configured.
func NewI2C(bus drivers.I2C) Device {
	return Device{
		bus: &I2CBus{
			wire:    bus,
			Address: Address,
		},
	}
}
*/

// default address is 0x3C
func New(bus drivers.I2C, address uint16, width int16, height int16, extVCC bool) Device {
	return Device{bus: bus,
		address:     address,
		width:       width,
		height:      height,
		externalVCC: extVCC,
	}
}

func (d *Device) Configure() {
	d.pages = d.height / 8
	d.lineBytes = d.width / 8
	d.bufferSize = d.width * d.height / 8
	/*
		d.currBuffer = make([]byte, size)
		d.prevBuffer = make([]byte, size)
		for i := range prevBuffer {
			prevBuffer[i] = 0xFF
		}
	*/
	d.buffer = make([]byte, d.bufferSize)

	if d.width == 128 && d.height == 64 {
		d.pageMode = false
	} else if (d.width == 64 && d.height == 128) || (d.width == 128 && d.height == 128) {
		d.pageMode = true
	}

	time.Sleep(100 * time.Nanosecond)

	d.Command(SET_DISP)
	if d.pageMode {
		d.Command(SET_MEM_MODE)
	} else {
		d.Command(SET_MEM_MODE | 0x01)
		//d.Command(SET_MEM_MODE)
	}
	d.Command(SET_DISP_START_LINE)
	d.Command(0x00)
	d.Command(SET_SEG_REMAP)
	if d.pageMode {
		d.Command(SET_COM_OUT_DIR)
	} else {
		d.Command(SET_COM_OUT_DIR | 0x08)
	}

	d.Command(SET_MUX_RATIO)
	d.Command(0x7F)

	d.Command(SET_DISP_OFFSET)
	if d.width != d.height {
		d.Command(0x60)
	} else {
		d.Command(0x00)
	}

	d.Command(SET_DISP_CLK_DIV)
	d.Command(0x50)
	d.Command(SET_PRECHARGE)
	if d.externalVCC {
		d.Command(0x22)
	} else {
		d.Command(0xF1)
	}
	d.Command(SET_VCOM_DESEL)
	d.Command(0x35)
	d.Command(SET_DCDC_MODE)
	d.Command(0x81)
	d.Command(SET_CONTRAST)
	d.Command(0x10)
	d.Command(SET_ENTIRE_ON)
	d.Command(SET_NORM_INV)
	d.Command(SET_DISP | 0x01)
}

// Command sends a one byte command to the display
func (d *Device) Command(command uint8) {
	//d.bus.tx([]byte{command}, true)
	d.bus.Tx(d.address, []byte{0x80, command}, nil)
}

// ClearBuffer clears the image buffer
func (d *Device) ClearBuffer() {
	for i := int16(0); i < d.bufferSize; i++ {
		d.buffer[i] = 0
	}
}

// ClearDisplay clears the image buffer and clear the display
func (d *Device) ClearDisplay() {
	d.ClearBuffer()
	d.Display()
}

/*
// setAddress sets the address to the I2C bus
func (b *I2CBus) setAddress(address uint16) {
	b.Address = address
}

// configure does nothing, but it's required to avoid reflection
func (b *I2CBus) configure() {}

// Tx sends data to the display
func (d *Device) Tx(data []byte, isCommand bool) {
	d.bus.tx(data, isCommand)
	d.bus.Tx(d.Address, []byte{0x80, SET_NORM_INV | 0x00}, nil)
}

// tx sends data to the display (I2CBus implementation)
func (b *I2CBus) tx(data []byte, isCommand bool) {
	if isCommand {
		b.wire.WriteRegister(uint8(b.Address), 0x00, data)
	} else {
		b.wire.WriteRegister(uint8(b.Address), 0x40, data)
	}
}
*/

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}

// Display sends the whole buffer to the screen
func (d *Device) Display() error {
	fmt.Println("Display.Display()")

	/* didn't work
	d.Command(SET_PAGE_ADDR | 1)
	d.Command(SET_COL_LO_ADDR | (1 & 0x0F)) //0x00
	d.Command(SET_COL_HI_ADDR | ((1 & 0x70) >> 4))
	*/
	if d.pageMode {
		for page := int16(0); page < d.pages; page++ {
			//d.Command(uint8(SET_PAGE_ADDR | (page1 - noffs)))
			buffer_i := page * d.width
			d.Command(uint8(SET_PAGE_ADDR | page))
			d.Command(uint8(SET_COL_LO_ADDR | 2)) //0x00
			d.Command(uint8(SET_COL_HI_ADDR))
			d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.width])
		}
	} else {
		/*
			for page := int16(0); page < d.pages; page++ {
				for row := range [...]int16{page * 8, (page + 1) * 8} {
					d.Command(uint8(SET_COL_LO_ADDR | (row & 0x0F)))
					d.Command(uint8(SET_COL_HI_ADDR | (row >> 4)))
					buffer_i := int16(row) * d.lineBytes
					d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.lineBytes])
				}
			}
		*/
		for col := int16(0); col < d.height; col++ {
			noffs := col * d.lineBytes
			for page := int16(0); page < d.pages; page++ {
				d.Command(uint8(SET_PAGE_ADDR | page))
				d.Command(uint8(SET_COL_LO_ADDR | (col & 0x0f)))
				d.Command(uint8(SET_COL_HI_ADDR | ((col & 0x70) >> 4)))
				d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[noffs:noffs+d.lineBytes])
			}
		}
	}

	/*
			0x00, //SET_COL_LO_ADDR
			SH110X_SETPAGEADDR + p,
		  0x10 + ((page_start + _page_start_offset) >> 4), //SET_COL_HI_ADDR
		  (page_start + _page_start_offset) & 0xF

				d.Command(SET_COL_LO_ADDR)
				d.Command(0)
				d.Command(uint8(d.width - 1))
				d.Command(SET_PAGE_ADDR)
				d.Command(0)
				d.Command(uint8(d.height/8) - 1)
	*/

	//d.Tx(d.buffer, false)
	//d.bus.Tx(d.address, d.buffer, nil)
	//	d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer)

	return nil
}

func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	byteIndex := x + (y/8)*d.width
	if c.R != 0 || c.G != 0 || c.B != 0 {
		d.buffer[byteIndex] |= 1 << uint8(y%8)
	} else {
		d.buffer[byteIndex] &^= 1 << uint8(y%8)
	}
}

// GetPixel returns if the specified pixel is on (true) or off (false)
func (d *Device) GetPixel(x int16, y int16) bool {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return false
	}
	byteIndex := x + (y/8)*d.width
	return (d.buffer[byteIndex] >> uint8(y%8) & 0x1) == 1
}

// SetBuffer changes the whole buffer at once
func (d *Device) SetBuffer(buffer []byte) error {
	if int16(len(buffer)) != d.bufferSize {
		//return ErrBuffer
		return errors.New("wrong size buffer")
	}
	for i := int16(0); i < d.bufferSize; i++ {
		d.buffer[i] = buffer[i]
	}
	return nil
}
