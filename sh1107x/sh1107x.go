// this driver is pretty specific to the dimensions of the Adafruit feather OLED 128 x 64
// based on the sh1107 driver
// Note that the dimensions have to be 64 width x 128 height to use page adressing mode
// if you want text in 128 x 64 it is easier to rotate the text then use vertical addressing mode
package sh1107x

import (
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
)

/*
Note that the ssd1306 driver does not use the 'builtin' i2c driver Tx command
Note my Trellis driver does not use write register
*/

// Device wraps I2C or SPI connection.
type Device struct {
	bus         drivers.I2C
	buffer      []byte
	width       int16
	height      int16
	bufferSize  int16
	pages       int16
	lineBytes   int16
	size        int16
	externalVCC bool
	address     uint16
	pageMode    bool

	testBuffer  []byte
	clearBuffer []byte
}

// default address of the sh1107 is 0x3C
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
	d.buffer = make([]byte, d.bufferSize)
	d.testBuffer = make([]byte, 16)
	d.testBuffer[0] = 0xff
	d.testBuffer[1] = 0x0
	d.testBuffer[2] = 0x0
	d.testBuffer[3] = 0x0
	d.testBuffer[4] = 0x0
	d.testBuffer[5] = 0x0
	d.testBuffer[6] = 0x0
	d.testBuffer[7] = 0x0
	d.testBuffer[8] = 0x0
	d.testBuffer[9] = 0x0
	d.testBuffer[10] = 0xff
	d.testBuffer[11] = 0x0
	d.testBuffer[12] = 0x0
	d.testBuffer[13] = 0x0
	d.testBuffer[14] = 0x0
	d.testBuffer[15] = 0xff

	d.clearBuffer = make([]byte, 16)
	d.clearBuffer[0] = 0x0
	d.clearBuffer[1] = 0x0
	d.clearBuffer[2] = 0x0
	d.clearBuffer[3] = 0x0
	d.clearBuffer[4] = 0x0
	d.clearBuffer[5] = 0x0
	d.clearBuffer[6] = 0x0
	d.clearBuffer[7] = 0x0
	d.clearBuffer[8] = 0x0
	d.clearBuffer[9] = 0x0
	d.clearBuffer[10] = 0x0
	d.clearBuffer[11] = 0x0
	d.clearBuffer[12] = 0x0
	d.clearBuffer[13] = 0x0
	d.clearBuffer[14] = 0x0
	d.clearBuffer[15] = 0x0
	if d.width == int16(128) && d.height == int16(64) {
		d.pageMode = false
	} else if (d.width == 64 && d.height == 128) || (d.width == 128 && d.height == 128) {
		d.pageMode = true
	} else {
		println("Dimensions don't work")
	}

	time.Sleep(100 * time.Nanosecond)

	d.Command(SET_DISP)
	if d.pageMode {
		d.Command(SET_MEM_MODE) // page mode; | 0x01 for vert mode
	} else {
		d.Command(SET_MEM_MODE | 0x01)
	}
	d.Command(SET_DISP_START_LINE)
	d.Command(0x00)
	d.Command(SET_SEG_REMAP | 0x00) // 0 is normal and 1 is reverse
	if d.pageMode {
		d.Command(SET_COM_OUT_DIR) // for page mode | 0x08 for vert mode
	} else {
		d.Command(SET_COM_OUT_DIR | 0x00) // 0x08 for vert mode if not reverses text
	}
	d.Command(SET_MUX_RATIO)
	d.Command(0x7F) //0x7f
	d.Command(SET_DISP_OFFSET)
	if d.width != d.height {
		//d.Command(0x60) //width != height else when == its 0x00
		d.Command(0x60) //width != height else when == its 0x00
	} else {
		d.Command(0x00) //width != height else when == its 0x00
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

	/*
		if d.pageMode {
			d.SetPixel = SetPixelVLSB
		} else {
			d.SetPixel = SetPixelHMSB
		}
	*/
}

// Command sends a one byte command to the display
func (d *Device) Command(command uint8) {
	d.bus.Tx(d.address, []byte{0x80, command}, nil)
}

func (d *Device) ClearBuffer() {
	for i := int16(0); i < d.bufferSize; i++ {
		d.buffer[i] = 0
	}
}

func (d *Device) ClearDisplay() {
	d.ClearBuffer()
	println("ClearBuffer")
	d.Display()
}

// Display sends the whole buffer to the screen
func (d *Device) Display() error {
	println("Entering Display()")

	if d.pageMode {
		fmt.Printf("PageMode = %t\r\n", d.pageMode)
		for page := int16(0); page < d.pages; page++ {
			buffer_i := page * d.width
			d.Command(uint8(SET_PAGE_ADDR | page))
			d.Command(uint8(SET_COL_LO_ADDR))
			d.Command(uint8(SET_COL_HI_ADDR))
			d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.width])
		}
	} else {
		fmt.Printf("PageMode = %t\r\n", d.pageMode)
		var col int16
		for col = 0; col < 64; col++ {
			//for col = 0; col < d.width; col++ {
			//buffer_i := col * d.pages
			//buffer_i := col * d.lineBytes
			//buffer_i := col * d.height
			d.Command(uint8(SET_COL_LO_ADDR | (col & 0x0f)))
			//d.Command(uint8(SET_COL_HI_ADDR | (col&0x70)>>4))
			d.Command(uint8(SET_COL_HI_ADDR | (0xF & (col >> 4))))
			d.Command(uint8(SET_PAGE_ADDR | 0))
			//d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.pages])
			d.bus.WriteRegister(uint8(d.address), 0x40, d.testBuffer[:])
			//d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.lineBytes])
			//d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.height])
		}
		fmt.Printf("col = %d\r\n", col)

	}
	println("Leaving Display()")

	return nil
}

func (d *Device) ClearDisplay2() error {
	println("Entering ClearDisplay2()")

	if d.pageMode {
		fmt.Printf("PageMode = %t\r\n", d.pageMode)
		for page := int16(0); page < d.pages; page++ {
			buffer_i := page * d.width
			d.Command(uint8(SET_PAGE_ADDR | page))
			d.Command(uint8(SET_COL_LO_ADDR))
			d.Command(uint8(SET_COL_HI_ADDR))
			d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.width])
		}
	} else {
		fmt.Printf("PageMode = %t\r\n", d.pageMode)
		var col int16
		for col = 0; col < 128; col++ {
			//for col = 0; col < d.width; col++ {
			//buffer_i := col * d.pages
			//buffer_i := col * d.lineBytes
			//buffer_i := col * d.height
			d.Command(uint8(SET_COL_LO_ADDR | (col & 0x0f)))
			//d.Command(uint8(SET_COL_HI_ADDR | (col&0x70)>>4))
			d.Command(uint8(SET_COL_HI_ADDR | (0xF & (col >> 4))))
			d.Command(uint8(SET_PAGE_ADDR | 0))
			//d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.pages])
			d.bus.WriteRegister(uint8(d.address), 0x40, d.clearBuffer[:])
			//d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.lineBytes])
			//d.bus.WriteRegister(uint8(d.address), 0x40, d.buffer[buffer_i:buffer_i+d.height])
		}
		fmt.Printf("col = %d\r\n", col)

	}
	println("Leaving Display()")

	return nil
}

//func SetPixelVLSB(d *Device, x int16, y int16, c color.RGBA) {
//func SetPixelHMSB(x int16, y int16, c color.RGBA) {
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	if d.pageMode {
		byteIndex := x + (y/8)*d.width
		if c.R != 0 || c.G != 0 || c.B != 0 {
			d.buffer[byteIndex] |= 1 << uint8(y%8)
		} else {
			d.buffer[byteIndex] &^= 1 << uint8(y%8)
		}
	} else {
		//if x < 0 || x >= d.width || y < 0 || y >= d.height {
		//		return
		//}
		//byteIndex := (x + y*d.width) >> 3 //####### with no roti: half screen char too tall
		byteIndex := (x*d.height + y) >> 3 //$$$$$$$ with no rot: horizontal chars reversed but correct size half screen
		//byteIndex := y + (x/8)*d.height
		/*
			byteIndex := (x + y*d.height) >> 3
			if x > 128 {
				byteIndex := (x - 127 + y*d.height) >> 3
				byteIndex += 520
			}
		*/

		//yy := bits.Reverse8(uint8(y))
		if c.R != 0 || c.G != 0 || c.B != 0 {
			//d.buffer[byteIndex] |= 1 << uint8(x%8) //#######
			d.buffer[byteIndex] |= 1 << uint8(y%8) //$$$$$$
			//d.buffer[byteIndex] |= 1 << bits.Reverse8(uint8(y%8))
		} else {
			//d.buffer[byteIndex] &^= 1 << uint8(x%8) //########
			d.buffer[byteIndex] |= 1 << uint8(y%8) //$$$$$$
			//d.buffer[byteIndex] &^= 1 << yy % 8 //x
		}
	}
}

// GetPixel returns if the specified pixel is on (true) or off (false)
/*
func (d *Device) GetPixel(x int16, y int16) bool {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return false
	}
	byteIndex := x + (y/8)*d.width
	return (d.buffer[byteIndex] >> uint8(y%8) & 0x1) == 1
}
*/

// Size returns the current size of the display.
// Needed by tinyfont
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}
