package bmff

import (
	"encoding/binary"
	"fmt"
)

// type box struct {
// 	Tag       *efmt.Ntag // for structured identification and printing
// 	boxtype   string     // from 4 byte field
// 	usertype  string     // if boxtype == 'uuid' then this is that uuid
// 	size      uint32     // size includes all header data starting at firt byte (boxtype)
// 	largesize int64      // if size == 1 then use this 'largesize' for size
// 	boxExt_s             // this embedded field embodies the "Full Box extension"... available for all boxes
// 	raw       []byte
//
// 	// container Vars boxes typically don't act as containers and also decoders
// 	typeNotDecoded star // flag that we don't know how to parse this
// 	readIdx        int
// 	writeIdx       int
// 	subBox         []Box
// }
//
// type boxExt_s struct {
// 	isFullBox bool // is the ext In use
// 	version   uint8
// 	flags     [3]byte
// }
//
// // Generic Box interface
// type Box interface {
// 	Type() string
// 	Size() int64
// 	Raw() []byte // when subBoxes exist, raw returns an empty array
//
// 	GetSubBoxCount() int
// 	ResetSubBox()
// 	GetSubBox() (Box, error)
// 	AddSubBox(Box)
//
// 	PrintDetail()
// 	PrintRecursive()
// 	Output(io.Writer, int) (writeCount int, err error)
// }

// func (b *box) Output(w io.Writer, objDepth int) (writeCount int, err error) {
// 	// basic writer outputs only containers and raw
// 	// must use proper boxtype for other boxes
// 	wCount, err := b.outputHeader(w)
// 	if err != nil {
// 		return wCount, err
// 	}
// 	if objDepth > 0 && b.GetSubBoxCount() != 0 {
// 		for _, subBox := range b.subBox {
// 			oC, bErr := subBox.Output(w, objDepth-1)
// 			if bErr != nil {
// 				bErr = fmt.Errorf("Got subbox.Output error: %v", err)
// 				return wCount, bErr
// 			}
// 			wCount += oC
// 		}
// 		return wCount, nil
// 	}
// 	if b.size == 0 {
// 		// shouln't be possible
// 		err = fmt.Errorf("We don't support size=0")
// 		return 0, err
// 	}
//
// 	// output the raw payload...  account for the extended header
// 	payloadIdx := 0
// 	if b.isFullBox {
// 		payloadIdx = 4
// 	}
// 	//fmt.Printf("Output P: %+v\n", b.raw)
// 	writeCnt, err := w.Write(b.raw[payloadIdx:])
// 	wCount += writeCnt
// 	//fmt.Printf("Output: %6d ", wCount)
// 	//b.pThis()
// 	return wCount, nil
// }

// ***********************   EmsgBox ***********************
/*
aligned(8) class DASHEventMessageBox extends FullBox(‘emsg’, version = 0, flags = 0){
   string            scheme_id_uri;
   string            value;
   unsigned int(32)  timescale;
   unsigned int(32)  presentation_time_delta;
   unsigned int(32)  event_duration;
   unsigned int(32)  id;
   unsigned int(8)   message_data[];
} }

DASH:  ISO_23009-1  Section 5.10.3.3.4 Semantics
scheme_id_uri:
    Identifies the message scheme. The semantics and syntax of the message_data[] are defined
    by the owner of the scheme identified. The string may use URN or URL syntax. When a URL is used,
    it is recommended to also contain a month-date in the form mmyyyy; the assignment of the URL must have been authorized by the owner of the domain name in that URL on or very close to that date. A URL may resolve to an Internet location, and a location that does resolve may store a specification of the message scheme.
value:
    Specifies the value for the event. The value space and semantics must be defined by the
    owners of the scheme identified in the scheme_id_uri field.
timescale:
    provides the timescale, in ticks per second, for the time and duration fields within this box
presentation_time_delta:
    Provides the Media Presentation time delta of the media presentation time of the event and the
    earliest presentation time in this segment. If the segment index is present, then the earliest
    presentation time is determined by the field earliest_presentation_time of the first 'sidx' box.
    If the segment index is not present, the earliest presentation time is determined as the earliest presentation
    time of any access unit in the media segment. The timescale is provided in the timescale field
event_duration:
    Provides the duration of event in media presentation time. The timescale is indicated in the timescale field.
    The value 0xFFFF indicates an unknown duration.
id:
    A field identifying this instance of the message. Messages with equivalent semantics shall have
    the same value, i.e. processing of any one event message box with the same id is sufficient.
message_data:
    Body of the message, which fills the remainder of the message box. This may be empty depending
    on the above information. The syntax and semantics of this field must be defined by the owner
    of the scheme identified in the scheme_id_uri field.



*/
type EmsgBox struct { // is a Fullbox, version=0, flags = 9
	*box
	scheme_id_uri           string
	value                   string
	timescale               uint32 // bigEndian
	presentation_time_delta uint32 // bigEndian
	event_duration          uint32 // bigEndian
	id                      uint32 // bigEndian
	message_data            string
}

// recursive function to print out the box type, size and substructure of a box
func (b *EmsgBox) PrintDetail() {
	children := "   "
	cCount := b.GetSubBoxCount()
	if cCount > 0 {
		children = fmt.Sprintf("%2d ", cCount)
	}
	fmt.Printf("%-16s %-19s %7d ", b.Tag.String(), b.Tag.Indent()+children+b.boxtype+" "+b.typeNotDecoded.String(), b.size)
	fmt.Printf("SchemeIdUri: %s Value:\"%s\" timescale:%d presentationTimeDelta:%d eventDuration:%d id:%d messageData:%s\n",
		b.scheme_id_uri, b.value, b.timescale, b.presentation_time_delta, b.event_duration, b.id, b.message_data)
}

func (b *EmsgBox) PrintRecursive() {
	var bx Box
	bx = b
	bx.PrintDetail()
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	for err == nil {
		b1.PrintRecursive()
		b1, err = b.GetSubBox()
	}
}

func parseString(input []byte, start int) (result string, next int) {
	for idx, char := range input[start:] {
		if char == 0 {
			return string(input[start : start+idx]), start + idx + 1
		}
	}
	return "", start
}

// encode null terminated string and return output length
func encodeString(output []byte, startOffset int, oStr string) (length int) {
	strLen := len(oStr)
	copy(output[startOffset:startOffset+strLen], []byte(oStr))
	output[startOffset+strLen] = 0
	return strLen + 1
}

func (b *EmsgBox) parse() error {
	b.parseFullBoxExt() // consume [0:4] of the raw payload of the base box => version and flags
	offset := 4
	b.scheme_id_uri, offset = parseString(b.raw, offset)
	if b.scheme_id_uri == "" {
		return fmt.Errorf("In EmsgBox::parse string not terminated\n")
	}

	b.value, offset = parseString(b.raw, offset)
	if b.value == "" {
		return fmt.Errorf("In EmsgBox::parse string not terminated")
	}
	b.timescale = binary.BigEndian.Uint32(b.raw[offset : offset+4])
	b.presentation_time_delta = binary.BigEndian.Uint32(b.raw[offset+4 : offset+8])
	b.event_duration = binary.BigEndian.Uint32(b.raw[offset+8 : offset+12])
	b.id = binary.BigEndian.Uint32(b.raw[offset+12 : offset+16])
	b.message_data = string(b.raw[offset+16:])
	return nil
}

func (b *EmsgBox) Encode() (encodeSize int, er error) {
	estRawSize := len(b.scheme_id_uri) + len(b.value) + len(b.message_data) + 2 /* for null termination*/ + 16 /* 4 *uint32 */ + 4 /* fullBoxext */
	fmt.Printf("estRawSize:%d  uri:%d val:%d messg:%d \n", estRawSize, len(b.scheme_id_uri), len(b.value), len(b.message_data))
	b.raw = make([]byte, estRawSize, estRawSize+2)
	offset := b.EncodeFullHeaderExt()
	offset += encodeString(b.raw, offset, b.scheme_id_uri) // returns length including null
	offset += encodeString(b.raw, offset, b.value)         // returns length including null
	binary.BigEndian.PutUint32(b.raw[offset:offset+4], b.timescale)
	offset += 4
	binary.BigEndian.PutUint32(b.raw[offset:offset+4], b.presentation_time_delta)
	offset += 4
	binary.BigEndian.PutUint32(b.raw[offset:offset+4], b.event_duration)
	offset += 4
	binary.BigEndian.PutUint32(b.raw[offset:offset+4], b.id)
	offset += 4
	offset += copy(b.raw[offset:], []byte(b.message_data)) // no null termination for last string at end of message
	b.raw = b.raw[0:offset]
	return offset, nil

}
