package bmff

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pkg/errors"

	"efmt"
)

// Box is defined as "object‐oriented building block defined by a unique type identifier and length".
// Used in MP4 Containers and referenced as an "atom" in some specifications including the first definition of MP4.
//
/*  box parsing: byteAligned box(size(4bbyte),boxType(4byte),{largesize(8byte)},{extended_type(16)}
    if size == 1 then parse/use largesize
    if size == 0 then box extends to EOF
    if boxtype == 'uuid' then parse extended_type
*/
//  mb 2018_0627: add usertype to box.
// mb 2018_0627: change largesize to int64 (from uint64): uint64 inherently wrong for comparison

type star bool

func (i star) String() string {
	if i == true {
		return "*"
	}
	return " "
}

type box struct {
	Tag       *efmt.Ntag // for structured identification and printing
	boxtype   string     // from 4 byte field
	usertype  string     // if boxtype == 'uuid' then this is that uuid
	size      uint32     // size includes all header data starting at firt byte (boxtype)
	largesize int64      // if size == 1 then use this 'largesize' for size
	boxExt_s             // this embedded field embodies the "Full Box extension"... available for all boxes
	raw       []byte

	// container Vars boxes typically don't act as containers and also decoders
	typeNotDecoded star // flag that we don't know how to parse this
	readIdx        int
	writeIdx       int
	subBox         []Box
}

type boxExt_s struct {
	isFullBox bool // is the ext In use
	version   uint8
	flags     [3]byte
}

// Generic Box interface
type Box interface {
	Type() string
	Size() int64
	Raw() []byte // when subBoxes exist, raw returns an empty array

	GetSubBoxCount() int
	ResetSubBox()
	GetSubBox() (Box, error)
	AddSubBox(Box)

	PrintDetail()
	PrintRecursive()
	Output(io.Writer, int) (writeCount int, err error)
}

// helper function to parse the FullBox Extension
// note:  Parsing is pulled from the raw payload... size is not adjusted
func (b *box) parseFullBoxExt() {
	b.isFullBox = true
	b.version = b.raw[0]
	copy(b.flags[0:3], b.raw[1:4])
}

// recursive function to print out the box type, size and substructure of a box
func (b *box) PrintDetail() {
	children := "   "
	cCount := b.GetSubBoxCount()
	if cCount > 0 {
		children = fmt.Sprintf("%2d ", cCount)
	}
	fmt.Printf("%-16s %-19s %7d\n", b.Tag.String(), b.Tag.Indent()+children+b.boxtype+" "+b.typeNotDecoded.String(), b.size)
}

func (b *box) PrintRecursive() {
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

func (b *box) Size() int64 {
	if b == nil {
		return 0
	}
	if b.size == 1 {
		return b.largesize
	}
	return int64(b.size)
}

func (b *box) Type() string {
	return b.boxtype
}

func (b *box) Raw() []byte {
	return b.raw
}

func (b *box) GetSubBoxCount() int {
	return len(b.subBox)
}

func (b *box) ResetSubBox() {
	b.readIdx = 0
}
func (b *box) GetSubBox() (Box, error) {
	if b.readIdx == b.writeIdx {
		return nil, fmt.Errorf("No box available")
	}
	b.readIdx++
	return b.subBox[b.readIdx-1], nil
}
func (b *box) AddSubBox(aBox Box) {
	b.subBox = append(b.subBox, aBox)
	b.writeIdx++
}

func (b *box) EncodeFullHeaderExt() (writeCount int) {
	if !b.isFullBox {
		return 0
	}
	b.raw[0] = b.version
	copy(b.raw[1:4], b.flags[:])
	return 4
}

func (b *box) outputHeader(w io.Writer) (writeCount int, err error) {
	// make a large header area and then shorten it for writing
	// assemble out header into this array and then output it
	oh := make([]byte, 128)

	// encode size, then boxType
	binary.BigEndian.PutUint32(oh[0:4], b.size)
	copy(oh[4:8], []byte(b.boxtype))
	oIdx := 8

	// optionally encode largeSize and usertype
	if b.size == 1 {
		binary.BigEndian.PutUint64(oh[oIdx:oIdx+8], uint64(b.largesize))
		oIdx += 8
	}
	if b.boxtype == "uuid" {
		copy(oh[oIdx:oIdx+16], []byte(b.usertype))
		oIdx += 16
	}

	// now optionally encode full header
	// ********  WARNING... This is actually consumed from the raw payload.... so it must be factored
	// out later
	if b.isFullBox {
		oh[oIdx] = b.version
		copy(oh[oIdx+1:oIdx+4], b.flags[:])
		oIdx += 4
	}
	return w.Write(oh[0:oIdx])

}

func (b *box) Output(w io.Writer, objDepth int) (writeCount int, err error) {
	// basic writer outputs only containers and raw
	// must use proper boxtype for other boxes

	// FIXME  should confirm the size for each box is sum of subboxes/payload plus header
	wCount, err := b.outputHeader(w)
	if err != nil {
		return wCount, err
	}
	if objDepth > 0 && b.GetSubBoxCount() != 0 {
		for _, subBox := range b.subBox {
			oC, bErr := subBox.Output(w, objDepth-1)
			if bErr != nil {
				bErr = fmt.Errorf("Got subbox.Output error: %v", err)
				return wCount, bErr
			}
			wCount += oC
		}
		return wCount, nil
	}

	if b.size == 0 {
		// shouln't be possible
		err = fmt.Errorf("We don't support size=0")
		return 0, err
	}

	// output the raw payload...  account for the extended header
	payloadIdx := 0
	if b.isFullBox {
		payloadIdx = 4
	}
	//fmt.Printf("Output P: %+v\n", b.raw)
	writeCnt, err := w.Write(b.raw[payloadIdx:])
	wCount += writeCnt
	//fmt.Printf("Output: %6d ", wCount)
	//b.pThis()
	return wCount, nil
}

// ****************************

//NewBox returns a raw parsing of a box from the input io.Reader
const boxHeaderSize = 8

func NewBox(src io.Reader, newtag *efmt.Ntag) (*box, error) {

	buf := make([]byte, boxHeaderSize)
	// read and parse first 8 bytes
	_, err := io.ReadFull(src, buf)
	if err != nil {
		return nil, errors.Wrap(err, "error reading buffer header")
	}
	s := binary.BigEndian.Uint32(buf[0:4])
	b := &box{
		boxtype: string(buf[4:8]),
		size:    s,
		Tag:     newtag.Clone(),
	}
	bufUsed := 8

	if s == 1 {
		// read in largesize and parseSdesChunk
		_, err := io.ReadFull(src, buf)
		if err != nil {
			return nil, errors.Wrap(err, "error reading buffer header large size")
		}
		b.largesize = int64(binary.BigEndian.Uint64(buf))
		bufUsed += 8
	}
	if b.boxtype == "uuid" {
		// read in largesize and parseSdesChunk
		buf1 := make([]byte, 16)
		_, err := io.ReadFull(src, buf1)
		if err != nil {
			return nil, errors.Wrap(err, "error reading buffer header uuid")
		}
		b.usertype = string(buf1)
		bufUsed += 16
	}
	if b.size == 0 {
		return nil, errors.Wrap(err, "NewBox: unable to handle size=0")
	}
	rawSize := b.Size() - int64(bufUsed)
	if rawSize > 0 {
		b.raw = make([]byte, rawSize)
		_, err = io.ReadFull(src, b.raw)
		if err != nil {
			return nil, errors.Wrap(err, "Error reading box data")
		}
		//fmt.Printf("%-16s %-16s %7d\n", b.Tag.String(), b.Tag.Indent()+b.boxtype, b.size)
		return b, nil
	}
	return nil, nil
}

// *********************************************************
