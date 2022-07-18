package trellis

// based on github.com/adafruit/Adafruit_Trellis_Library and
// github.com/adafruit/Adafruit_CircuitPython_Trellis

import (
	"time"

	"tinygo.org/x/drivers"
)

// Default address = 0x70

const (
	HT16K33_OSCILATOR_ON    = 0x21
	HT16K33_BLINK_CMD       = 0x80
	HT16K33_BLINK_DISPLAYON = 0x01
	HT16K33_CMD_BRIGHTNESS  = 0xE0
	HT16K33_KEY_READ_CMD    = 0x40
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
	bus        drivers.I2C
	address    uint16
	brightness uint8
}

func New(bus drivers.I2C, address uint16, brightness uint8) Device {
	return Device{bus, address, brightness}
}

func (d Device) Configure() {
	//println("Begin Configure")
	err := d.bus.Tx(d.address, []byte{HT16K33_OSCILATOR_ON}, nil)
	if err != nil {
		println("Could not turn oscillator on:", err)
		return
	}
	time.Sleep(1 * time.Second)
	d.blinkRate(HT16K33_BLINK_OFF)
	d.setBrightness(d.brightness) // max brightness = 15
	d.bus.Tx(d.address, []byte{0xA1}, nil)
}

func IsKeyPressed(k uint8) bool {
	if k > 15 {
		return false
	}
	key := buttonLUT[k]
	if (keys[key>>4] & (1 << (key & 0x0F))) != 0 {
		return true
	}
	return false
}

func WasKeyPressed(k uint8) bool {
	if k > 15 {
		return false
	}
	key := buttonLUT[k]
	//fmt.Printf("%d: %d: %d - %x & %x\r\n", k, key, key>>4, lastkeys[key>>4], (1 << (key & 0x0F)))
	if (lastkeys[key>>4] & (1 << (key & 0x0F))) != 0 {
		return true
	}
	return false
}

func JustPressed(k uint8) bool {
	return (IsKeyPressed(k) && !WasKeyPressed(k))
}

func justReleased(k uint8) bool {
	return (!IsKeyPressed(k) && WasKeyPressed(k))
}

func isLED(x uint8) bool {
	if x > 15 {
		return false
	}
	led := ledLUT[x]
	return (displaybuffer[led>>4]&(1<<(led&0x0F)) > 0)
}

func SetLED(x uint8) {
	if x > 15 {
		return
	}
	led := ledLUT[x]
	displaybuffer[led>>4] |= (1 << (led & 0x0F))
}

func clrLED(x uint8) {
	if x > 15 {
		return
	}
	led := ledLUT[x]
	displaybuffer[led>>4] &= (1 << (led & 0x0F))
}

func (d Device) ReadSwitches() bool {
	lastkeys := keys
	var buf [6]byte
	d.bus.Tx(d.address, []byte{0x40}, buf[:])
	copy(keys[:], buf[:])

	for i := 0; i < 6; i++ {
		if lastkeys[i] != keys[i] {
			//fmt.Printf("lastkeys = %v\r\n", lastkeys)
			//fmt.Printf("keys = %v\r\n", keys)
			return true
		}
	}
	return false
}

func (d Device) ReadSingleSwitch() (bool, int) {
	var buf [6]byte
	d.bus.Tx(d.address, []byte{uint8(0x40)}, buf[:])

	for i := 0; i < 6; i++ {
		if buf[i] != 0 {
			for k := 0; k < 16; k++ {
				key := buttonLUT[k]
				if (buf[key>>4] & (1 << (key & 0x0F))) != 0 {
					return true, k
				}
			}
		}
	}
	return false, 0
}

func (d Device) setBrightness(b uint8) {
	if b > 15 {
		b = 15
	}
	d.bus.Tx(d.address, []byte{HT16K33_CMD_BRIGHTNESS | b}, nil)
}

func (d Device) blinkRate(b uint8) {
	if b > 3 {
		b = 0
	}
	x := HT16K33_BLINK_CMD | HT16K33_BLINK_DISPLAYON | (b << 1)
	err := d.bus.Tx(d.address, []byte{x}, nil)
	if err != nil {
		println("Could not set blink rate")
		return
	}
}

func (d Device) WriteDisplay() {
	var buf [17]uint8
	buf[0] = 0x00
	for i := 0; i < 8; i++ {
		buf[2*i+1] = uint8(displaybuffer[i] & 0xFF)
		buf[2*i+2] = uint8(displaybuffer[i] >> 8)
	}
	err := d.bus.Tx(d.address, buf[:], nil)
	if err != nil {
		println("Could not write display")
		return
	}
}

func Clear() {
	for i := range displaybuffer {
		displaybuffer[i] = 0
	}
}
