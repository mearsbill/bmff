package bmff

import (
	"encoding/binary"
	"fmt"
)

type Uint16_16 uint32

func (x Uint16_16) String() string {
	const shift, mask = 16, 1<<16 - 1
	return fmt.Sprintf("%d.%d", uint32(x)>>shift, uint32(x)&mask)
}

func (x *Uint16_16) UnmarshalBinary(b []byte) error {
	*x = Uint16_16(binary.BigEndian.Uint32(b))
	return nil
}

type Uint8_8 uint16
type Int8_8 int16

func (x *Uint8_8) UnmarshalBinary(b []byte) error {
	*x = Uint8_8(binary.BigEndian.Uint16(b))
	return nil
}
func (x *Int8_8) UnmarshalBinary(b []byte) error {
	*x = Int8_8(binary.BigEndian.Uint16(b))
	return nil
}
func (x Uint8_8) String() string {
	const shift, mask = 8, 1<<8 - 1
    intPart := x>>shift
    fractPart := float32(x & mask) / float32(mask + 1)
	return fmt.Sprintf("%d.%3.3f", intPart, fractPart )
}
func (x Int8_8) String() string {
	const shift, mask = 8, 1<<8 - 1
	return fmt.Sprintf("%d.%d", x>>shift, x&mask)
}
