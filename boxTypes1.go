package bmff

import (
	"encoding/binary"
	"fmt"
)

// File is the top level containter for the decode
type File_s struct {
	*box
	Ftyp *FtypBox // file tyoe box
	// pdin  progressive download info
	Moov *MoovBox // container all metadata
	Moof *MoofBox // movie fragment
	// mfra  movie fragment random access
	Mdat *MdatBox
	// free  free space
	// skip  free space
	Meta *MetaBox // metadata
	Styp *StypBox // segment type
	Emsg *EmsgBox // event message box
	Sidx *SidxBox // segment index
	// meco  additional metadata container
	// ssix  subsegement index
	// prft  produceer reference time
	// AllBoxes []Box
}

// PrintAll  Helper function to output a tree of the whole file
func (f *File_s) PrintDetail() {
	fmt.Printf("File Contents in order:\n\n")
}
func (b *File_s) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}
func (f *File_s) PrintAll() {
	f.PrintRecursive()
}

type FtypBox struct {
	*box
	MajorBrand       string
	MinorVersion     int
	CompatibleBrands []string
}

func (b *FtypBox) parse() error {
	b.MajorBrand, b.MinorVersion = string(b.raw[0:4]), int(binary.BigEndian.Uint32(b.raw[4:8]))
	if l := len(b.raw); l > 8 {
		for i := 8; i < l; i += 4 {
			b.CompatibleBrands = append(b.CompatibleBrands, string(b.raw[i:i+4]))
		}
	}
	return nil
}

// specific funciton for this typwe
func (b *FtypBox) PrintDetail() {
	children := "   "
	cCount := b.GetSubBoxCount()
	if cCount > 0 {
		children = fmt.Sprintf("%2d ", cCount)
	}
	fmt.Printf("%-16s %-19s %7d", b.Tag.String(), b.Tag.Indent()+children+b.boxtype+" "+b.typeNotDecoded.String(), b.size)
	fmt.Printf(" MajBrand:%s MinVer:%d  CompBrands(%d): ", b.MajorBrand, b.MinorVersion, len(b.CompatibleBrands))
	for i := 0; i < len(b.CompatibleBrands); i++ {
		fmt.Printf("%4s ", b.CompatibleBrands[i])
	}
	fmt.Printf("\n")

}
func (b *FtypBox) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}

type StypBox struct {
	*box
	MajorBrand       string
	MinorVersion     int
	CompatibleBrands []string
}

func (b *StypBox) parse() error {
	b.MajorBrand, b.MinorVersion = string(b.raw[0:4]), int(binary.BigEndian.Uint32(b.raw[4:8]))
	if l := len(b.raw); l > 8 {
		for i := 8; i < l; i += 4 {
			b.CompatibleBrands = append(b.CompatibleBrands, string(b.raw[i:i+4]))
		}
	}
	return nil
}

// specific funciton for this typwe
func (b *StypBox) PrintDetail() {
	children := "   "
	cCount := b.GetSubBoxCount()
	if cCount > 0 {
		children = fmt.Sprintf("%2d ", cCount)
	}
	fmt.Printf("%-16s %-19s %7d", b.Tag.String(), b.Tag.Indent()+children+b.boxtype+" "+b.typeNotDecoded.String(), b.size)
	fmt.Printf(" MajBrand:%s MinVer:%d  CompBrands(%d): ", b.MajorBrand, b.MinorVersion, len(b.CompatibleBrands))
	for i := 0; i < len(b.CompatibleBrands); i++ {
		fmt.Printf("%4s ", b.CompatibleBrands[i])
	}
	fmt.Printf("\n")

}
func (b *StypBox) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}

// *********************************************************
type SidxRef struct {
	// reference_type      uint8        // 1 bit
	// reference_size      uint32      // 31 bits
	// subsegment_duration uint32
	// starts_with_SAP     uint8       // 1 bit
	// SAP_type            uint8       // 3 bits
	// SAP_delta_time      uint32      // 28 bits
	rawDat []byte // remains encoded until we need it otherwise
}

type SidxBox struct {
	*box
	reference_ID               uint32
	timescale                  uint32
	earliest_presentation_time uint64
	first_offset               uint64
	reserved                   uint16
	reference_count            uint16
	refs                       []*SidxRef
}

func (b *SidxBox) parse() error {

	b.parseFullBoxExt() // consume [0:4] => version and flags
	offset := 4         // from decoding the FullBoxExt
	b.reference_ID = binary.BigEndian.Uint32(b.raw[offset : offset+4])
	offset += 4
	b.timescale = binary.BigEndian.Uint32(b.raw[offset : offset+4])
	offset += 4
	if b.version == 0 {
		b.earliest_presentation_time = uint64(binary.BigEndian.Uint32(b.raw[offset : offset+4]))
		b.first_offset = uint64(binary.BigEndian.Uint32(b.raw[offset+4 : offset+8]))
		offset += 8
	} else if b.version == 1 {
		b.earliest_presentation_time = binary.BigEndian.Uint64(b.raw[offset : offset+8])
		b.first_offset = binary.BigEndian.Uint64(b.raw[offset+8 : offset+16])
		offset += 16
	}
	b.reserved = binary.BigEndian.Uint16(b.raw[offset : offset+2])
	offset += 2
	b.reference_count = binary.BigEndian.Uint16(b.raw[offset : offset+2])
	offset += 2
	//fmt.Printf("SIDX:  offset:%d raw[0:%d] refCnt:%d \n BOX: %+v, lBox:%+v\n", offset, len(b.raw), b.reference_count, b, b.box)
	for i := 0; i < int(b.reference_count); i++ {
		sr := SidxRef{rawDat: b.raw[offset : offset+12]}
		//copy(b.refs[i].rawDat[:], b.raw[offset:offset+12])
		offset += 12
		b.refs = append(b.refs, &sr)
	}
	return nil
}

// specific funciton for this typwe
func (b *SidxBox) PrintDetail() {
	children := "   "
	cCount := b.GetSubBoxCount()
	if cCount > 0 {
		children = fmt.Sprintf("%2d ", cCount)
	}
	fmt.Printf("%-16s %-19s %7d", b.Tag.String(), b.Tag.Indent()+children+b.boxtype+" "+b.typeNotDecoded.String(), b.size)
	fmt.Printf(" Ver:%d Flags:0x%02x%02x%02x", b.version, b.flags[0], b.flags[1], b.flags[2])
	fmt.Printf(" EPT:%d fOffs:%d rsrvd:%d  RefCnt(%d) ", b.earliest_presentation_time, b.first_offset, b.reserved, len(b.refs))
	// for i := 0; i < len(b.CompatibleBrands); i++ {
	// 	fmt.Printf("%4s ", b.CompatibleBrands[i])
	// }
	fmt.Printf("\n")

}
func (b *SidxBox) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}
