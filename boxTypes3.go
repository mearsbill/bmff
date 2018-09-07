package bmff

import (
	"encoding/binary"
	"fmt"
	"klog"
)

// *******  Movie Fragment Box **************************************************

type MoofBox struct {
	*box
	Mfhd *MfhdBox
	Meta *MetaBox
	Traf []*TrafBox
}

func (b *MoofBox) parse() error {
	var err error
	for subBox := range readBoxes(b.raw, b.Tag) {
		if subBox == nil {
			return nil
		}
		switch subBox.boxtype {
		case "mfhd":
			b.Mfhd = &MfhdBox{box: subBox}
			if err1 := b.Mfhd.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.AddSubBox(b.Mfhd)
		case "meta":
			b.Meta = &MetaBox{box: subBox}
			if err1 := b.Meta.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.AddSubBox(b.Meta)
		case "traf":
			traf := &TrafBox{box: subBox}
			if err1 := traf.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Traf = append(b.Traf, traf)
			b.AddSubBox(traf)
		default:
			kl.KWarn(klog.KlrNotHandled, "%s: Unknown Moof subtype(%s) SubType: %s\n", subBox.Tag.String(), b.Tag.String(), subBox.Type())
			subBox.typeNotDecoded = true
			b.AddSubBox(subBox)
		}
	}

	return err
}

// specific funciton for this typwe
func (b *MoofBox) PrintDetail() {
	children := "   "
	cCount := b.GetSubBoxCount()
	if cCount > 0 {
		children = fmt.Sprintf("%2d ", cCount)
	}
	fmt.Printf("%-16s %-19s %7d", b.Tag.String(), b.Tag.Indent()+children+b.boxtype+b.typeNotDecoded.String(), b.size)
	fmt.Printf("\n")

}
func (b *MoofBox) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}

// *********************************************************
//  MovieFragmentHeaderBox
type MfhdBox struct {
	*box
	sequence_number uint32
}

func (b *MfhdBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags
	b.sequence_number = binary.BigEndian.Uint32(b.raw[4:8])
	return nil
}

func (b *MfhdBox) PrintDetail() {
	fmt.Printf("%-16s %-19s %7d", b.Tag.String(), b.Tag.Indent()+"   "+b.boxtype+b.typeNotDecoded.String(), b.size)
	fmt.Printf(" SeqNum: %d", b.sequence_number)
	fmt.Printf("\n")

}
func (b *MfhdBox) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	if err != nil {
		return // kl.KError(KlrWrapper, "%v", err)
	}
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
		if err != nil {
			return // kl.KError(KlrWrapper, "%v", err)
		}
	}
}

// ****** Meta Data***************************************************
// can appear in moov trak meco moof trak
//

type MetaBox struct {
	*box
	// not fleshed out
}

func (b *MetaBox) parse() error {
	return nil
}

// ***** Track Fragment ****************************************************

type TrafBox struct {
	*box
	Tfhd *TfhdBox
	Trun *TrunBox
	Tfdt *TfdtBox
	Meta *MetaBox
}

func (b *TrafBox) parse() error {
	var err error
	for subBox := range readBoxes(b.raw, b.Tag) {
		if subBox == nil {
			break
		}

		switch subBox.boxtype {
		case "tfhd":
			header := &TfhdBox{box: subBox}
			if err1 := header.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Tfhd = header
			b.AddSubBox(header)
		case "trun":
			header := &TrunBox{box: subBox}
			if err1 := header.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Trun = header
			b.AddSubBox(header)
		case "tfdt":
			header := &TfdtBox{box: subBox}
			if err1 := header.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Tfdt = header
			b.AddSubBox(header)
		case "meta":
			meta := &MetaBox{box: subBox}
			if err1 := meta.parse(); err1 != nil {
				err = kl.KWarn(klog.KlrWrapper, "%v", err1)
			}
			b.Meta = meta
			b.AddSubBox(meta)
		default:
			err = kl.KWarn(klog.KlrNotHandled, "Unknown Traf SubType: %s\n", subBox.Type())
			b.AddSubBox(subBox)
			subBox.typeNotDecoded = true

		}
	}
	return err
}
func (b *TrafBox) PrintDetail() {
	children := "   "
	cCount := b.GetSubBoxCount()
	if cCount > 0 {
		children = fmt.Sprintf("%2d ", cCount)
	}

	fmt.Printf("%-16s %-19s %7d", b.Tag.String(), b.Tag.Indent()+children+b.boxtype+" "+b.typeNotDecoded.String(), b.size)
	fmt.Printf("\n")

}
func (b *TrafBox) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}

// *********************************************************
// TrackFragmentHeaderBox
//
// note:  the flags of the ext have the following meaning:
// The following flags are defined in the tf_flags:
// 0x000001  base‐data‐offset‐present: indicates the presence of the base‐data‐offset field. This provides an explicit anchor
//           for the data offsets in each track run (see below). If not provided and if the default‐base‐is‐moof flag is not set,
//           the base‐data‐offset for the first track in the movie fragment is the position of the first byte of the
//           enclosing Movie Fragment Box, and for second and subsequent track fragments, the default is the end of the data defined
//           by the preceding track fragment. Fragments 'inheriting' their offset in this way must all use the same data‐
//           reference (i.e., the data for these tracks must be in the same file)
// 0x000002  sample‐description‐index‐present: indicates the presence of this field, which over‐rides,
//           in this fragment, the default set up in the Track Extends Box.
// 0x000008  default‐sample‐duration‐present
// 0x000010  default‐sample‐size‐present
// 0x000020  default‐sample‐flags‐present
// 0x010000  duration‐is‐empty: this indicates that the duration provided in either default‐sample‐duration, or by the default‐duration
//           in the Track Extends Box, is empty, i.e. that there are no samples for this time interval. It is an error to make a presentation
//           that has both edit lists in the Movie Box, and empty‐duration fragments.
// 0x020000  default‐base‐is‐moof: if base‐data‐offset‐present is 1, this flag is ignored. If base‐data‐ offset‐present is
//           zero, this indicates that the base‐data‐offset for this track fragment is the position of the first byte of the
//           enclosing Movie Fragment Box. Support for the default‐base‐is‐ moof flag is required under the ‘iso5’ brand, and it shall
//           not be used in brands or compatible brands earlier than iso5.
// NOTE The use of the default‐base‐is‐moof flag breaks the compatibility to earlier brands of the file format, because
//      it sets the anchor point for offset calculation differently than earlier. Therefore, the default‐base‐is‐moof flag
//      cannot be set when earlier brands are included in the File Type box.
//
type TfhdBox struct { // section 8.8.7.1
	*box
	track_ID                 uint32
	base_data_offset         uint64
	sample_description_index uint32
	default_sample_duration  uint32
	default_sample_size      uint32
	default_sample_flags     uint32
}

func (b *TfhdBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags
	offset := 4         // from decoding the FullBoxExt - version not used
	rawLen := len(b.raw)
	if rawLen-offset < 4 {
		return kl.KWarn(klog.KlrRanOutOfData, "TfhdBox.parse ran out of bits")
	}
	b.track_ID = binary.BigEndian.Uint32(b.raw[offset : offset+4])
	offset += 4 // from decoding the FullBoxExt - version not used
	if (b.flags[2] & 0x01) != 0 {
		if rawLen-offset < 8 {
			return kl.KWarn(klog.KlrRanOutOfData, "TfhdBox.parse ran out of bits")
		}
		b.base_data_offset = binary.BigEndian.Uint64(b.raw[offset : offset+8])
		offset += 8
	}
	if (b.flags[2] & 0x02) != 0 {
		if rawLen-offset < 4 {
			return kl.KWarn(klog.KlrRanOutOfData, "TfhdBox.parse ran out of bits")
		}
		b.sample_description_index = binary.BigEndian.Uint32(b.raw[offset : offset+4])
		offset += 4
	}
	if (b.flags[2] & 0x08) != 0 {
		if rawLen-offset < 4 {
			return kl.KWarn(klog.KlrRanOutOfData, "TfhdBox.parse ran out of bits")
		}
		b.default_sample_duration = binary.BigEndian.Uint32(b.raw[offset : offset+4])
		offset += 4
	}
	if (b.flags[2] & 0x10) != 0 {
		if rawLen-offset < 4 {
			return kl.KWarn(klog.KlrRanOutOfData, "TfhdBox.parse ran out of bits")
		}
		b.default_sample_size = binary.BigEndian.Uint32(b.raw[offset : offset+4])
		offset += 4
	}
	if (b.flags[2] & 0x20) != 0 {
		if rawLen-offset < 4 {
			return kl.KWarn(klog.KlrRanOutOfData, "TfhdBox.parse ran out of bits")
		}
		b.default_sample_flags = binary.BigEndian.Uint32(b.raw[offset : offset+4])
		offset += 4
	}
	return nil
}
func (b *TfhdBox) PrintDetail() {
	children := "   "
	fmt.Printf("%-16s %-19s %7d ", b.Tag.String(), b.Tag.Indent()+children+b.boxtype+" "+b.typeNotDecoded.String(), b.size)
	fmt.Printf("flg:%02x%02x%02x ", b.flags[0], b.flags[1], b.flags[2])
	fmt.Printf("baseDatOff:0x%x smplDscrIdx:%d defSmplDur:%d ", (b.base_data_offset), b.sample_description_index, b.default_sample_duration)
	fmt.Printf("defSamplSz:%d defSmplFlgs:%d ", b.default_sample_size, b.default_sample_flags)
	fmt.Printf("DurEmpty:%v defBaseIsMoof:%v ", (b.flags[0]&0x01) != 0, (b.flags[0]&0x02) != 0)
	fmt.Printf("\n")

}
func (b *TfhdBox) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}

// *********************************************************
// Track Run Boxes
type TrunSample struct {
	sample_duration                uint32
	sample_size                    uint32
	sample_flags                   uint32
	sample_composition_time_offset uint32 // is signed for version != 0
}
type TrunBox struct {
	*box               // extended full box with version and flags
	sample_count       uint32
	data_offset        int32 // option (Flag bit)
	first_sample_flags uint32
	rSamples           []TrunSample
}

func (b *TrunBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags
	offset := 4         // from decoding the FullBoxExt - version not used
	rawLen := len(b.raw)

	if rawLen-offset < 4 {
		return kl.KWarn(klog.KlrRanOutOfData, "TrunBox.parse ran out of bits fetching sample_count")
	}
	b.sample_count = binary.BigEndian.Uint32(b.raw[offset : offset+4])
	offset += 4

	if (b.flags[2] & 0x01) != 0 { // data-offset-present flag
		if rawLen-offset < 4 {
			return kl.KWarn(klog.KlrRanOutOfData, "TrunBox.parse ran out of bits fetching data_offset")
		}
		b.data_offset = int32(binary.BigEndian.Uint32(b.raw[offset : offset+4]))
		//kl.KTrace("TrunBox.parse: data_offset is present and is %d", b.data_offset)
		offset += 4 // from decoding the FullBoxExt - version not used
	}
	first_sample_flags_present := (b.flags[2] & 0x04) != 0
	if first_sample_flags_present {
		// when present indicaates:
		// - first_sample_flags is present in stream
		// - sample flags NOT present in loop
		if rawLen-offset < 4 {
			return kl.KWarn(klog.KlrRanOutOfData, "TrunBox.parse ran out of bits fetching first_sample_flags")
		}
		b.first_sample_flags = binary.BigEndian.Uint32(b.raw[offset : offset+4])
		offset += 4 // from decoding the FullBoxExt - version not used
	}

	// flags for in-loop data presenr
	sample_duration_present := (b.flags[1] & 0x01) != 0
	sample_size_present := (b.flags[1] & 0x02) != 0
	sample_flags_present := ((b.flags[1] & 0x04) != 0) && (!first_sample_flags_present)
	sample_composition_time_offset_present := (b.flags[1] & 0x08) != 0

	//for idx := 0; idx < int(b.sample_count); idx++ {
	for idx := 0; idx < int(b.sample_count); idx++ {
		sd := uint32(0)
		if sample_duration_present {
			if rawLen-offset < 4 {
				return kl.KWarn(klog.KlrRanOutOfData, "TrunBox.parse ran out of bits fetching sample_duration (%d of %d)", idx, b.sample_count)
			}
			sd = binary.BigEndian.Uint32(b.raw[offset : offset+4])
			offset += 4 // from decoding the FullBoxExt - version not used
		}

		ss := uint32(0)
		if sample_size_present {
			if rawLen-offset < 4 {
				return kl.KWarn(klog.KlrRanOutOfData, "TrunBox.parse ran out of bits fetching sample_size in loop (%d of %d)", idx, b.sample_count)
			}
			ss = binary.BigEndian.Uint32(b.raw[offset : offset+4])
			offset += 4 // from decoding the FullBoxExt - version not used
		}
		local_sfp := sample_flags_present && !first_sample_flags_present
		sf := uint32(0)
		if local_sfp {
			if rawLen-offset < 4 {
				return kl.KWarn(klog.KlrRanOutOfData, "TrunBox.parse ran out of bits fetching sample_flags in loop (%d of %d)", idx, b.sample_count)
			}
			sf = binary.BigEndian.Uint32(b.raw[offset : offset+4])
			offset += 4 // from decoding the FullBoxExt - version not used
		} else if (idx == 0) && first_sample_flags_present {
			sf = b.first_sample_flags
		}

		scto := uint32(0)
		if sample_composition_time_offset_present {
			if rawLen-offset < 4 {
				return kl.KWarn(klog.KlrRanOutOfData, "TrunBox.parse ran out of bytes(need 4 have %d) fetching sample_composition_time_offset in loop (%d of %d)", rawLen-offset, idx, b.sample_count)
			}
			scto = binary.BigEndian.Uint32(b.raw[offset : offset+4])
			offset += 4 // from decoding the FullBoxExt - version not used
		}
		// if sd == ss || sf == scto {
		// 	fmt.Printf("Hello WOrld\n")
		// }
		ts := TrunSample{sample_duration: sd, sample_size: ss, sample_flags: sf, sample_composition_time_offset: scto}
		b.rSamples = append(b.rSamples, ts)
	}
	return nil
}

func (b *TrunBox) PrintDetail() {
	children := "   "
	fmt.Printf("%-16s %-19s %7d ", b.Tag.String(), b.Tag.Indent()+children+b.boxtype+" "+b.typeNotDecoded.String(), b.size)
	fmt.Printf("flg:%02x%02x%02x smplCnt:%d dataOffset:%d firstSampleFlags:0x%08x ", b.flags[0], b.flags[1], b.flags[2],
		b.sample_count, b.data_offset, b.first_sample_flags)
	fmt.Printf("\n")
	if false {
		localCount := int(b.sample_count)
		if false {
			localCount = 6
		}
		if b.sample_count > 0 {
			fmt.Printf("%-16s %-9s %17s ", "", "", "Sample Recs")
			for idx := 0; idx < localCount; idx++ {
				fmt.Printf("%d: Dur:%d Siz:%d Flags:0x%x CompositionTimeOffs:%d",
					idx, b.rSamples[idx].sample_duration, b.rSamples[idx].sample_size, b.rSamples[idx].sample_flags, b.rSamples[idx].sample_composition_time_offset)
				fmt.Printf("\n")
				if idx != localCount-1 {
					fmt.Printf("%-16s %-19s %7s ", "", "", "")
				}
			}
			if localCount != int(b.sample_count) {
				fmt.Printf("%-16s %-19s %7s     ........\n", "", "", "")
			}
		}
	}

}
func (b *TrunBox) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}

// *********************************************************
//TrackFragmentBaseMediaDecodeTimeBox
type TfdtBox struct {
	*box
	baseMediaDecodeTime uint64 // in units of media's timescale (which is ticksPerSecond)
}

func (b *TfdtBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] => version and flags
	offset := 4         // from decoding the FullBoxExt
	if b.version == 0 {
		b.baseMediaDecodeTime = uint64(binary.BigEndian.Uint32(b.raw[offset : offset+4]))
		//offset += 4
	} else if b.version == 1 {
		b.baseMediaDecodeTime = binary.BigEndian.Uint64(b.raw[offset : offset+8])
		//offset += 8
	}
	return nil
}

func (b *TfdtBox) PrintDetail() {
	fmt.Printf("%-16s %-19s %7d ", b.Tag.String(), b.Tag.Indent()+"   "+b.boxtype+" "+b.typeNotDecoded.String(), b.size)
	fmt.Printf("baseMediaDecodeTime: %d ", b.baseMediaDecodeTime)
	fmt.Printf("\n")

}
func (b *TfdtBox) PrintRecursive() {
	b.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}
