package bmff

import (
	"encoding/binary"
	"klog"
)

// *********************************************************

type MoovBox struct {
	*box
	MovieHeader *MvhdBox
	TrackBoxes  []*TrakBox
	Iods        *IodsBox
	// Meta *MetaBox

}

func (b *MoovBox) parse() error {
	var err error
	for subBox := range readBoxes(b.raw, b.Tag) {
		if subBox == nil {
			return nil
		}
		switch subBox.boxtype {
		case "mvhd":
			b.MovieHeader = &MvhdBox{box: subBox}
			if err1 := b.MovieHeader.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
		case "iods":
			b.Iods = &IodsBox{box: subBox}
			if err1 := b.Iods.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
		case "trak":
			trak := &TrakBox{box: subBox}
			if err1 := trak.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.TrackBoxes = append(b.TrackBoxes, trak)
		default:
			err = kl.KWarn(klog.KlrNotHandled, "%s: Unknown Moov(%s) SubType: %s\n", subBox.Tag.String(), b.Tag.String(), subBox.Type())
			subBox.typeNotDecoded = true
		}
		b.AddSubBox(subBox)
	}

	return err
}

// *********  Meta Data container ************************************************
type MdatBox struct {
	*box
}

func (b *MdatBox) parse() error {
	return nil
}

// *********************************************************

type TrakBox struct {
	*box
	Tkhd *TkhdBox
	Tref *TrefBox
	Mdia *MdiaBox
}

func (b *TrakBox) parse() error {
	var err error
	for subBox := range readBoxes(b.raw, b.Tag) {
		if subBox == nil {
			break
		}

		switch subBox.boxtype {
		case "tkhd":
			header := &TkhdBox{box: subBox}
			if err1 := header.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Tkhd = header
		case "mdia":
			mdia := &MdiaBox{box: subBox}
			if err1 := mdia.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Mdia = mdia
		case "tref":
			tref := &TrefBox{box: subBox}
			if err1 := tref.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Tref = tref
		default:
			err = kl.KWarn(klog.KlrWrapper, "%s: Unknown Trak(%s) SubType: %s\n", subBox.Tag.String(), b.Tag.String(), subBox.Type())
			subBox.typeNotDecoded = true

		}
		b.AddSubBox(subBox)
	}
	return err
}

// *********************************************************

type MdiaBox struct {
	*box
	Mdhd *MdhdBox
	Hdlr *HdlrBox
	Minf *MinfBox
}

func (b *MdiaBox) parse() error {
	var err error
	for subBox := range readBoxes(b.raw, b.Tag) {
		if subBox == nil {
			break
		}

		switch subBox.boxtype {
		case "mdhd":
			mdhd := MdhdBox{box: subBox}
			if err1 := mdhd.parse(); err != nil {
				return kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Mdhd = &mdhd
		case "hdlr":
			hdlr := HdlrBox{box: subBox}
			if err1 := hdlr.parse(); err != nil {
				return kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Hdlr = &hdlr
		case "minf":
			minf := MinfBox{box: subBox}
			if err1 := minf.parse(); err != nil {
				return kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Minf = &minf
		default:
			subBox.typeNotDecoded = true
			err = kl.KWarn(klog.KlrNotHandled, "Unknown MdiaBox SubType: %s\n", subBox.Type())

		}
		b.AddSubBox(subBox)
	}
	return err
}

// *********************************************************

func langString(langCode uint16) string {
	b0 := (langCode >> 10) + 0x60
	b1 := ((langCode >> 5) & 0x3f) + 0x60
	b2 := ((langCode) & 0x3f) + 0x60
	return string([]byte{byte(b0), byte(b1), byte(b2)})
	//fmt.Printf("langCode = 0x%x langStr = %s\n",b.langCode, b.langStr)
}

type MdhdBox struct {
	*box
	CreationTime     uint64
	ModificationTime uint64
	TimeScale        uint32
	Duration         uint64
	langCode         uint16
	langStr          string
}

// <Media Header Box
func (b *MdhdBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags

	// Fullbox payload begins at offset 4

	var offset int
	if b.version == 0 {
		b.CreationTime = uint64(binary.BigEndian.Uint32(b.raw[4:8]))
		b.ModificationTime = uint64(binary.BigEndian.Uint32(b.raw[8:12]))
		b.TimeScale = binary.BigEndian.Uint32(b.raw[12:16])
		b.Duration = uint64(binary.BigEndian.Uint32(b.raw[16:20]))
		offset = 20
	} else if b.version == 1 {
		b.CreationTime = binary.BigEndian.Uint64(b.raw[4:12])
		b.ModificationTime = binary.BigEndian.Uint64(b.raw[12:20])
		b.TimeScale = binary.BigEndian.Uint32(b.raw[20:24])
		b.Duration = binary.BigEndian.Uint64(b.raw[24:32])
		offset = 32
	}

	//b.langCode = binary.BigEndian.Uint16(b.raw[offset : offset+2])
	b.langCode = (uint16(b.raw[offset])<<8 | uint16(b.raw[offset+1]))
	b.langStr = langString(b.langCode)

	return nil
}

// *********************************************************

type HdlrBox struct {
	*box
	// pre-defined 0 (32)
	handlerType uint32
	name        string
	// reserved 0 3*(32)
}

func (b *HdlrBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags

	// Fullbox payload begins at offset 4
	// skip over Predefined (4 bytes)
	b.handlerType = binary.BigEndian.Uint32(b.raw[8:12])
	return nil
}

// *********************************************************

// exactly 1 minf required in mdia
type MinfBox struct {
	*box
	// no decodes... just sub boxes
	Vmhd *VmhdBox
	Smhd *SmhdBox
	Hmhd *HmhdBox
	Nmhd *NmhdBox
	Dinf *DinfBox
	Stbl *StblBox
}

func (b *MinfBox) parse() error {
	var err error
	for subBox := range readBoxes(b.raw, b.Tag) {
		if subBox == nil {
			break
		}

		switch subBox.boxtype {
		case "vmhd":
			vmhd := VmhdBox{box: subBox}
			if err1 := vmhd.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Vmhd = &vmhd
		case "smhd":
			smhd := SmhdBox{box: subBox}
			if err1 := smhd.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Smhd = &smhd
		case "hmhd":
			hmhd := HmhdBox{box: subBox}
			if err1 := hmhd.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Hmhd = &hmhd
		case "nmhd":
			nmhd := NmhdBox{box: subBox}
			if err1 := nmhd.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Nmhd = &nmhd
		case "dinf":
			dinf := DinfBox{box: subBox}
			if err1 := dinf.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Dinf = &dinf
		case "stbl":
			stbl := StblBox{box: subBox}
			if err1 := stbl.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Stbl = &stbl
		default:
			err = kl.KWarn(klog.KlrNotHandled, "Unknown Minf SubType: %s\n", subBox.Type())
			subBox.typeNotDecoded = true

		}
		b.AddSubBox(subBox)
	}
	return err
}

//
// *********************************************************

// Video Media Header
type VmhdBox struct {
	*box
	graphicsmode uint16
	opcolor      [3]uint16
}

func (b *VmhdBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags
	//no decoding required
	return nil
}

// *********************************************************

// Sound Media Header
type SmhdBox struct {
	*box
	balance  Int8_8
	reserved uint16
}

func (b *SmhdBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags

	//no decoding required
	return nil
}

// *********************************************************

// HintMediaHeader
type HmhdBox struct {
	*box
	maxPDUsize uint16
	avgPDUsize uint16
	maxbitrate uint32
	avgbitrate uint32
	reserved   uint32
}

func (b *HmhdBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags
	b.maxPDUsize = binary.BigEndian.Uint16(b.raw[4:6])
	b.avgPDUsize = binary.BigEndian.Uint16(b.raw[6:8])
	b.maxbitrate = binary.BigEndian.Uint32(b.raw[8:12])
	b.avgbitrate = binary.BigEndian.Uint32(b.raw[12:16])
	// reserved not decoded

	return nil
}

// Null Media Header
type NmhdBox struct {
	*box
}

func (b *NmhdBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags

	//no decoding required
	return nil
}

// Data Information box, container
type DinfBox struct {
	*box
	balance  Int8_8
	reserved uint16
}

func (b *DinfBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags

	//no decoding required
	return nil
}

// Data Information box, container
type StblBox struct {
	*box
}

func (b *StblBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags

	//no decoding required
	return nil
}

// *********************************************************

// exactly 0 or 1 udta in moov or trak
// User data box
type UdtaBox struct {
	*box
	Cprt *CprtBox
}

func (b *UdtaBox) parse() error {
	var err error
	for subBox := range readBoxes(b.raw, b.Tag) {
		if subBox == nil {
			break
		}

		switch subBox.boxtype {
		case "cprt":
			cprt := CprtBox{box: subBox}
			if err1 := cprt.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Cprt = &cprt
		default:
			err = kl.KWarn(klog.KlrNotHandled, "Unknown Udta SubType: %s\n", subBox.Type())
			subBox.typeNotDecoded = true

		}
		b.AddSubBox(subBox)
	}
	return err
}

// HintMediaHeader
type CprtBox struct {
	*box
	langCode uint16
	langStr  string
	notice   string
}

func (b *CprtBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags

	offset := 4
	//b.langCode = binary.BigEndian.Uint16(b.raw[offset : offset+2])
	b.langCode = (uint16(b.raw[offset])<<8 | uint16(b.raw[offset+1]))
	b.langStr = langString(b.langCode)
	offset += 2

	// null terminated string in UTF-8 or UTF-16
	// if UTF-16 starts with BYTE_ORDER_MARK (0xfeff)
	b.notice = string(b.raw[offset:])

	return nil
}

// *********************************************************

type MvhdBox struct {
	*box
	CreationTime     uint64
	ModificationTime uint64
	TimeScale        uint32
	Duration         uint64
	NextTrackID      uint32
	Rate             Uint16_16
	Volume           Uint8_8
	Reserved         []byte
	Matrix           [9]int32
	Predefined       []byte
}

func (b *MvhdBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags
	var offset int
	var err error
	if b.version == 0 {
		b.CreationTime = uint64(binary.BigEndian.Uint32(b.raw[4:8]))
		b.ModificationTime = uint64(binary.BigEndian.Uint32(b.raw[8:12]))
		b.TimeScale = binary.BigEndian.Uint32(b.raw[12:16])
		b.Duration = uint64(binary.BigEndian.Uint32(b.raw[16:20]))
		offset = 20
	} else if b.version == 1 {
		b.CreationTime = binary.BigEndian.Uint64(b.raw[4:12])
		b.ModificationTime = binary.BigEndian.Uint64(b.raw[12:20])
		b.TimeScale = binary.BigEndian.Uint32(b.raw[20:24])
		b.Duration = binary.BigEndian.Uint64(b.raw[24:32])
		offset = 32
	}

	if err1 := b.Rate.UnmarshalBinary(b.raw[offset : offset+4]); err1 != nil {
		err = kl.KWarn(klog.KlrFuncFail, "%v: failed to unmarshal rate", err1)
	}

	if err1 := b.Volume.UnmarshalBinary(b.raw[offset+4 : offset+6]); err1 != nil {
		err = kl.KWarn(klog.KlrFuncFail, "%v: failed to unmarshal volume", err1)
	}

	b.Reserved = b.raw[offset+6 : offset+16]
	offset += 16
	for i := 0; i < 9; i++ {
		b.Matrix[i] = int32(binary.BigEndian.Uint32(b.raw[offset+i : offset+i+4]))
	}
	offset += 36

	b.Predefined = b.raw[offset : offset+24]
	b.NextTrackID = binary.BigEndian.Uint32(b.raw[offset+24 : offset+28])
	return err
}

type IodsBox struct {
	*box
}

func (b *IodsBox) parse() error {
	return nil
}

type TkhdBox struct {
	*box
	CreationTime     uint64
	ModificationTime uint64
	TrackID          uint32
	Duration         uint64
	Layer            int16
	AlternateGroup   int16
	Volume           int16
	Matrix           [9]int32
	Width            Uint16_16
	Height           Uint16_16
}

func (b *TkhdBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags
	var offset int
	if b.version == 0 {
		b.CreationTime = uint64(binary.BigEndian.Uint32(b.raw[4:8]))
		b.ModificationTime = uint64(binary.BigEndian.Uint32(b.raw[8:12]))
		b.TrackID = binary.BigEndian.Uint32(b.raw[12:16])
		// 16:20 reserved
		b.Duration = uint64(binary.BigEndian.Uint32(b.raw[20:24]))
		offset = 24
	} else if b.version == 1 {
		b.CreationTime = binary.BigEndian.Uint64(b.raw[4:12])
		b.ModificationTime = binary.BigEndian.Uint64(b.raw[12:20])
		b.TrackID = binary.BigEndian.Uint32(b.raw[20:24])
		// 24:28 reserved
		b.Duration = binary.BigEndian.Uint64(b.raw[28:36])
		offset = 36
	}
	offset += 8 // reserved bytes
	b.Layer = int16(binary.BigEndian.Uint16(b.raw[offset : offset+2]))
	b.AlternateGroup = int16(binary.BigEndian.Uint16(b.raw[offset+2 : offset+4]))
	b.Volume = int16(binary.BigEndian.Uint16(b.raw[offset+4 : offset+6]))
	offset += 8 // previous bytes + 2 reserved

	for i := 0; i < 9; i++ {
		b.Matrix[i] = int32(binary.BigEndian.Uint32(b.raw[offset+i : offset+i+4]))
	}
	offset += 36
	b.Width = Uint16_16(binary.BigEndian.Uint32(b.raw[offset : offset+4]))
	b.Height = Uint16_16(binary.BigEndian.Uint32(b.raw[offset+4 : offset+8]))
	return nil
}

// ******** Track reference containter

type TrefBox struct {
	*box
	TypeBoxes []*TrefTypeBox
}

func (b *TrefBox) parse() error {
	for subBox := range readBoxes(b.raw, b.Tag) {
		if subBox == nil {
			break
		}
		t := TrefTypeBox{box: subBox}
		for i := 0; i < len(t.raw); i += 4 {
			t.TrackIDs = append(t.TrackIDs, binary.BigEndian.Uint32(t.raw[i:i+4]))

		}
		b.TypeBoxes = append(b.TypeBoxes, &t)
		b.AddSubBox(subBox)
	}
	return nil
}

type TrefTypeBox struct {
	*box
	TrackIDs []uint32
}
