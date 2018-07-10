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

func Parse(src io.Reader) (*File, error) {
	f := &File{}
	r := bufio.NewReader(src)

	topTag := efmt.NewNtag()

readloop:
	for {
		b, err := NewBox(r, topTag)
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
			f.Ftyp = fb
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
		case "styp":
			sb := &StypBox{box: b}
			if err := sb.parse(); err != nil {
				return nil, err
			}
			f.Styp = sb
		case "sidx":
			sb := &SidxBox{box: b}
			if err := sb.parse(); err != nil {
				return nil, err
			}
			f.Sidx = sb
		default:
			fmt.Printf("%s: @Top.. Unknown Type:%s\n", b.Tag.String(), b.Type())
		}
		f.AllBoxes = append(f.AllBoxes, b)

	}

	return f, nil
}