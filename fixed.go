package bmff

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Uint16_16 uint32
type Int16_16 int32

func (x Uint16_16) String() string {
	const shift, mask = 16, 1<<16 - 1
	intPart := x >> shift
	fractPart := float32(x&mask) / float32(mask+1)
	fractStr := fmt.Sprintf("%3.3f", fractPart)
	fractStr = strings.TrimPrefix(fractStr, "0.")
	return fmt.Sprintf("%d.%s", intPart, fractStr)
}

func (x Int16_16) String() string {
	negStr := ""
	if x < 0 {
		negStr = "-"
		x = -x
	}
	xx := Uint16_16(x)
	uStr := xx.String()
	return fmt.Sprintf("%s%s", negStr, uStr)
}

func (x *Uint16_16) UnmarshalBinary(b []byte) error {
	*x = Uint16_16(binary.BigEndian.Uint32(b))
	return nil
}

func (x *Int16_16) UnmarshalBinary(b []byte) error {
	*x = Int16_16(binary.BigEndian.Uint32(b))
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
	intPart := x >> shift
	fractPart := float32(x&mask) / float32(mask+1)
	fractStr := fmt.Sprintf("%2.2f", fractPart)
	fractStr = strings.TrimPrefix(fractStr, "0.")
	return fmt.Sprintf("%d.%s", intPart, fractStr)
}

func (x Int8_8) String() string {
	negStr := ""
	if x < 0 {
		negStr = "-"
		x = -x
	}
	xx := Uint8_8(x)
	uStr := xx.String()
	return fmt.Sprintf("%s%s", negStr, uStr)
}
