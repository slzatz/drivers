package trellis

import (
	"fmt"
	"time"

	"tinygo.org/x/drivers"
)

// HT16K33 Command Contstants
const (
	HT16K33_OSCILATOR_ON    = 0x21
	HT16K33_BLINK_CMD       = 0x80
	HT16K33_BLINK_DISPLAYON = 0x01
	HT16K33_CMD_BRIGHTNESS  = 0xE0
	HT16K33_KEY_READ_CMD    = 0x40

	Address = 0x70
)
const (
	HT16K33_BLINK_OFF    = 0
	HT16K33_BLINK_2HZ    = 1
	HT16K33_BLINK_1HZ    = 2
	HT16K33_BLINK_HALFHZ = 3
)

var (
	displaybuffer [8]uint16
	keys          [8]uint8
	lastkeys      [8]uint8

	ledLUT = [16]uint8{
		0x3A, 0x37, 0x35, 0x34,
		0x28, 0x29, 0x23, 0x24,
		0x16, 0x1B, 0x11, 0x10,
		0x0E, 0x0D, 0x0C, 0x02}

	buttonLUT = [16]int{
		0x07, 0x04, 0x02, 0x22,
		0x05, 0x06, 0x00, 0x01,
		0x03, 0x10, 0x30, 0x21,
		0x13, 0x12, 0x11, 0x31}
)

type Device struct {
	bus     drivers.I2C
	Address uint16
}

func New(bus drivers.I2C) Device {
	return Device{bus, Address}
}

func (d Device) Configure() {
	//d.bus.WriteRegister(uint8(d.Address), HT16K33_OSCILATOR_ON, []uint8{0})
	println("Begin Configure")
	// need this or no lights
	err := d.bus.Tx(d.Address, []byte{0x21}, nil)
	if err != nil {
		println("Could not turn oscillator on:", err)
		return
	}
	time.Sleep(1 * time.Second)
	//d.blinkRate(HT16K33_BLINK_2HZ)
	d.blinkRate(0)
	/*
		d.setBrightness(15) // max brightness
		d.bus.Tx(d.Address, []byte{0xA1}, nil)
	*/
}

//bool Adafruit_Trellis::isKeyPressed(uint8_t k) {
func isKeyPressed(k uint8) bool {
	if k > 15 {
		return false
	}
	key := buttonLUT[k]
	//#define _BV(bit) (1 << (bit))
	if (keys[key>>4] & (1 << (key & 0x0F))) != 0 {
		return true
	} else {
		return false
	}
}

//bool Adafruit_Trellis::wasKeyPressed(uint8_t k) {
func wasKeyPressed(k uint8) bool {
	if k > 15 {
		return false
	}
	key := buttonLUT[k]
	if (lastkeys[key>>4] & (1 << (key & 0x0F))) != 0 {
		return true
	} else {
		return false
	}
}

//boolean Adafruit_Trellis::justPressed(uint8_t k) {
func justPressed(k uint8) bool {
	return (isKeyPressed(k) && !wasKeyPressed(k))
}

//boolean Adafruit_Trellis::justReleased(uint8_t k) {
func justReleased(k uint8) bool {
	return (!isKeyPressed(k) && wasKeyPressed(k))
}

//boolean Adafruit_Trellis::isLED(uint8_t x) {
func isLED(x uint8) bool {
	if x > 15 {
		return false
	}
	led := ledLUT[x]
	//#define _BV(bit) (1 << (bit))
	return (displaybuffer[led>>4]&(1<<(led&0x0F)) > 0)
}

//void Adafruit_Trellis::setLED(x uint8) {
func SetLED(x uint8) {
	if x > 15 {
		return
	}
	//led := ledLUT[x]
	println("x =", x)
	x = ledLUT[x]
	//println("led =", led)
	println("x =", x)
	//displaybuffer[led>>4] |= (1 << (led & 0x0F))
	//#define _BV(bit) (1 << (bit))
	displaybuffer[x>>4] |= (1 << (x & 0x0F))
	fmt.Printf("db = %v\r\n", displaybuffer)
}

//void Adafruit_Trellis::clrLED(x uint8) {
func clrLED(x uint8) {
	if x > 15 {
		return
	}
	led := ledLUT[x]
	displaybuffer[led>>4] &= (1 << (led & 0x0F))
}

//boolean Adafruit_Trellis::readSwitches(void) {
func (d Device) readSwitches() bool {
	//memcpy(lastkeys, keys, sizeof(keys));
	lastkeys := keys

	//Wire.write(0x40);
	//Wire.requestFrom((byte)i2c_addr, (byte)6);
	var buf [6]byte
	d.bus.Tx(d.Address, []byte{uint8(0x40)}, buf[:])

	/*
		  Wire.requestFrom((byte)i2c_addr, (byte)6);
			for i:=0; i<6; i++
		    keys[i] = Wire.read();
	*/

	copy(keys[:], buf[:])

	for i := 0; i < 6; i++ {
		if lastkeys[i] != keys[i] {
			return true
		}
	}
	return false
}

func (d Device) setBrightness(b uint8) {
	if b > 15 {
		b = 15
	}
	d.bus.Tx(d.Address, []byte{HT16K33_CMD_BRIGHTNESS | b}, nil)
}

func (d Device) blinkRate(b uint8) {
	if b > 3 {
		b = 0
	}
	x := HT16K33_BLINK_CMD | HT16K33_BLINK_DISPLAYON | (b << 1)
	err := d.bus.Tx(d.Address, []byte{x}, nil)
	if err != nil {
		println("Could not set blink rate")
		return
	}
}

//void Adafruit_Trellis::writeDisplay(void) {
func (d Device) WriteDisplay() {
	//Wire.write((uint8_t)0x00); // start at address $00
	err := d.bus.Tx(d.Address, []byte{0x00}, nil)
	if err != nil {
		println("Could not write 00")
		return
	}

	for i := 0; i < 8; i++ {
		//  Wire.write(displaybuffer[i] & 0xFF);
		// this is the first byte of the uint16
		fmt.Printf("i = %d, low = %v\r\n", i, []byte{uint8(displaybuffer[i] & 0xFF)})
		//err = d.bus.Tx(d.Address, []byte{uint8(displaybuffer[i] & 0xFF)}, nil)
		err := d.bus.Tx(d.Address, []byte{uint8(displaybuffer[i] & 0xFF)}, nil)
		if err != nil {
			println("Could not write first byte")
			return
		}
		//  Wire.write(displaybuffer[i] >> 8);
		// this is the second byte of the uint16
		fmt.Printf("i = %d, high = %v\r\n", i, []byte{uint8(displaybuffer[i] >> 8)})
		//err = d.bus.Tx(d.Address, []byte{uint8(displaybuffer[i] >> 8)}, nil)
		err = d.bus.Tx(d.Address, []byte{uint8(displaybuffer[i] >> 8)}, nil)
		if err != nil {
			println("Could not write second byte")
			return
		}
	}
}

//void Adafruit_Trellis::clear(void) {
func Clear() {
	//memset(displaybuffer, 0, sizeof(displaybuffer));
	for i := range displaybuffer {
		displaybuffer[i] = 0
	}
}
