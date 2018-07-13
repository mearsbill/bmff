package bmff

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/pkg/errors"

	"efmt"
)

func readBoxes(buf []byte, tag *efmt.Ntag) <-chan *box {
	boxes := make(chan *box, 1000)
	//    fmt.Printf("%v\n",tag)
	newTag := tag.Clone()
	newTag.Push()
	//    fmt.Printf("%v\n",newTag)
	r := bytes.NewReader(buf)
	go func() {
		for eof := false; !eof; {
			b, err := NewBox(r, newTag)
			newTag.Next()
			if err != nil {
				switch errors.Cause(err) {
				case io.EOF:
					eof = true
				}
			}

			boxes <- b
		}
		close(boxes)
	}()
	return boxes
}

// highest level parser.
func Parse(src io.Reader) (*File_s, error) {
	fb := &box{}
	f := &File_s{box: fb}
	r := bufio.NewReader(src)

	topTag := efmt.NewNtag()
	var bx Box
	var bxFlag bool

readloop:
	for {
		b, err := NewBox(r, topTag)
		bxFlag = false
		topTag.Next()
		if err != nil {
			switch errors.Cause(err) {
			case io.EOF:
				if b == nil {
					break readloop
				}
			default:
				return nil, err
			}
		}
		switch b.boxtype {
		case "ftyp":
			fb := &FtypBox{box: b}
			if err := fb.parse(); err != nil {
				return nil, err
			}
			f.Ftyp, bx, bxFlag = fb, fb, true

		case "styp":
			sb := &StypBox{box: b}
			if err := sb.parse(); err != nil {
				return nil, err
			}
			f.Styp = sb
			bx = sb
			bxFlag = true
			//f.AllBoxes = append(f.AllBoxes, b)
		// case pdin
		//
		case "moov":
			mb := &MoovBox{box: b}
			if err := mb.parse(); err != nil {
				return nil, err
			}
			f.Moov = mb
		case "moof":
			moof := &MoofBox{box: b}
			if err := moof.parse(); err != nil {
				return nil, err
			}
			f.Moof = moof
		// case mfra
		//
		case "mdat":
			mdat := &MdatBox{box: b}
			if err := mdat.parse(); err != nil {
				return nil, err
			}
			f.Mdat = mdat
			// case free
			//
			// case skip
			//
		case "meta":
			meta := &MetaBox{box: b}
			if err := meta.parse(); err != nil {
				return nil, err
			}
			f.Meta = meta
			// case meco
			//
		case "sidx":
			sb := &SidxBox{box: b}
			if err := sb.parse(); err != nil {
				return nil, err
			}
			f.Sidx = sb
			bx = sb
			bxFlag = true
		default:
			fmt.Printf("%s: @Top.. Unknown Type:%s\n", b.Tag.String(), b.Type())
			b.typeNotDecoded = true
		}
		if bxFlag {
			f.AddSubBox(bx)
		} else {
			f.AddSubBox(b)
		}

	}

	return f, nil
}

// function to output the  contents of the file object
// objDepth = 1 means just the top level (zero will also work for this)
func (f *File_s) Output(w io.Writer, objDepth int) (byteCount int, err error) {
	// depth of zero means: go no deeper
	totalByteCount := 0
	for idx, bx := range f.subBox {
		boxByteCount, err := bx.Output(w, objDepth-1)
		totalByteCount += boxByteCount
		if err != nil {
			err = fmt.Errorf("#%d.. Output got error: %v", idx, err)
			return totalByteCount, err
		}
	}

	return totalByteCount, nil
}
