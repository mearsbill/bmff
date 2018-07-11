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
	} else {
		return " "
	}
}

type box struct {
	Tag       *efmt.Ntag
	boxtype   string
	usertype  string // if boxtype == 'uuid' then this is that uuid
	size      uint32 //  size includes header
	largesize int64  // if size == 1 then use largesize for size
	boxExt_s
	raw []byte

	// container Vars boxes typically don't act as containers and also decoders
	unknown  star // we don't know how to parse this
	readIdx  int
	writeIdx int
	subBox   []*box
}

type boxExt_s struct {
	isFullBox bool // is the ext In use
	version   uint8
	flags     [3]byte
}

// helper function to parse the FullBox Extension
// note:  Parsing is pulled from the raw payload... size is not adjusted
func (b *box) parseFullBoxExt() {
	b.isFullBox = true
	b.version = b.raw[0]
	copy(b.flags[0:3], b.raw[1:4])
}

// Generic Box interface
type Box interface {
	Type() string
	Size() uint64
	Raw() []byte // when subBoxes exist, raw returns an empty array

	GetSubBoxCount() int
	ResetSubBox()
	GetSubBox() (*box, error)
	AddSubBox(*box)

	Output(io.Writer)
}

// recursive function to print out the box type, size and substructure of a box
func (b *box) pFunc() {
	fmt.Printf("%-16s %-16s %7d\n", b.Tag.String(), b.Tag.Indent()+b.boxtype+" "+b.unknown.String(), b.size)
	b.ResetSubBox()
	b1, err := b.GetSubBox()
	//fmt.Printf("First Next is %v\n",err)
	for err == nil {
		//fmt.Printf("Calling pfunc from pfunc for %+v\n",b1)
		b1.pFunc()
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
func (b *box) GetSubBox() (*box, error) {
	if b.readIdx == b.writeIdx {
		return nil, fmt.Errorf("No box available")
	}
	b.readIdx++
	return b.subBox[b.readIdx-1], nil
}
func (b *box) AddSubBox(aBox *box) {
	b.subBox = append(b.subBox, aBox)
	b.writeIdx++
}

func (b *box) Output(w io.Writer) (writeCount int, err error) {
	// basic writer outputs only containers and raw
	// must use proper boxtype for other boxes
	wCount := 0
	if b.GetSubBoxCount() != 0 {
		for _, subBox := range b.subBox {
			oC, bErr := subBox.Output(w)
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
		err = fmt.Errorf("Don't support size=0")
		return 0, err
	}

	// make a large header area and then shorten it for writing
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
	if b.isFullBox {
		oh[oIdx] = b.version
		copy(oh[oIdx+1:oIdx+4], b.flags[:])
		oIdx += 4
	}

	// output the header
	var writeCnt int
	//fmt.Printf("Output H: %+v\n", oh[0:oIdx])
	writeCnt, err = w.Write(oh[0:oIdx])
	if err != nil {
		return 0, err
	}
	wCount += writeCnt

	// output the raw payload.
	//fmt.Printf("Output P: %+v\n", b.raw)
	writeCnt, err = w.Write(b.raw)
	return writeCnt + wCount, nil

	// // perform basic raw output
	// err = binary.Write(w, binary.BigEndian, b.size)
	// if err != nil {
	// 	fmt.Printf("box.Output binary.Write got err: %v\n", err)
	// 	return 0, err
	// }
	// // if size is 1 then also output the largesize (8 bytes)
	// if b.size == 1 {
	// 	// largeSize
	// 	err = binary.Write(w, binary.BigEndian, b.largesize)
	// 	if err != nil {
	// 		fmt.Printf("box.Output binary.Write got err: %v\n", err)
	// 		return 0, err
	// 	}
	// }
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
