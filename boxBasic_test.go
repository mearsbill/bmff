package bmff

import (
	"bytes"
	"efmt"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// here test the functionality of the basic box.
// this include paring and outputing.... pretty well tested for the basic box
func TestBasicBoxes(t *testing.T) {

	tests := []struct {
		name    string
		data    []byte
		want    []Box
		output  bool
		wantErr bool
	}{
		{"Test small box parse", []byte{0, 0, 0, 16, 't', 'e', 's', 't', 1, 2, 3, 4, 5, 6, 7, 8}, nil, false, false},
		{"Test large box parse", []byte{0, 0, 0, 1, 't', 'e', 's', 't',
			0, 0, 0, 0, 0, 0, 0, 24, 1, 2, 3, 4, 5, 6, 7, 8}, nil, false, false},
		{"Test bad boxtype", []byte{0, 0, 0, 16, 'a', 'b', 'c', 'd', 1, 2, 3, 4, 5, 6, 7, 8}, nil, false, true},
		{"Test small output match", []byte{0, 0, 0, 16, 't', 'e', 's', 't', 1, 2, 3, 4, 5, 6, 7, 8}, nil, true, false},
		{"Test large  output match", []byte{0, 0, 0, 1, 't', 'e', 's', 't', 0, 0, 0, 0, 0, 0, 0, 24,
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p',
			1, 2, 3, 4, 5, 6, 7, 8}, nil, true, false},
		{"Test small uuid output match", []byte{0, 0, 0, 32, 'u', 'u', 'i', 'd',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p',
			1, 2, 3, 4, 5, 6, 7, 8}, nil, true, false},
		{"Test large uuid output match", []byte{0, 0, 0, 1, 'u', 'u', 'i', 'd', 0, 0, 0, 0, 0, 0, 0, 40,
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p',
			1, 2, 3, 4, 5, 6, 7, 8}, nil, true, false},
	}
	for idx, tt := range tests {
		reader := bytes.NewReader(tt.data) // read from the test data instead of a file/
		t.Run(tt.name, func(t *testing.T) {
			nt := efmt.NewNtag()
			switch {
			case idx == 0:
				b, bErr := NewBox(reader, nt)
				if bErr != nil {
					t.Errorf("#%d: %s: NewBox() error = %v", idx, tt.name, bErr)
					return
				}
				if b.size != 16 {
					t.Errorf("#%d: %s: NewBox() bad size=%d", idx, tt.name, b.size)
					return
				}
				if b.boxtype != "test" {
					t.Errorf("#%d: %s: NewBox() bad decod=%+v", idx, tt.name, b)
				}
			case idx == 1:
				b, _ := NewBox(reader, nt)
				if (b.size != 1) || (b.largesize != 24) {
					t.Errorf("#%d: %s: NewBox() bad size=%d largesize=%d", idx, tt.name, b.size, b.largesize)
					return
				}
				if b.boxtype != "test" {
					t.Errorf("#%d: %s: NewBox() bad decod=%+v", idx, tt.name, b)
				}
			case idx == 2:
				b, _ := NewBox(reader, nt)
				if b.size != 16 {
					t.Errorf("#%d: %s: NewBox() bad size=%d", idx, tt.name, b.size)
					return
				}
				if b.boxtype != "test" && !tt.wantErr {
					t.Errorf("#%d: %s: NewBox() bad decod=%+v", idx, tt.name, b)
				}
			case (idx >= 3) && (idx <= 6.):
				// testing Output matching input.
				// make an output buffer and use it as the output stream for the output mehtod
				// wierd behavior not conforming the docs.... so I have to use Truncate to get
				// the output buffer to behave
				buf := make([]byte, 128, 128)
				wB := bytes.NewBuffer(buf)
				wB.Truncate(0)

				b, _ := NewBox(reader, nt)
				wCnt, err := b.Output(wB)
				if err != nil {
					t.Errorf("#%d: %s: Output error: %v", idx, tt.name, err)
					return
				}
				testSize := b.largesize
				if b.size != 1 {
					testSize = int64(b.size)
				}
				if wCnt != int(testSize) {
					t.Errorf("#%d: %s: Size is %d, wanted %d", idx, tt.name, wCnt, testSize)
				}
				for i := 0; i < wCnt; i++ {
					if buf[i] != tt.data[i] {
						t.Errorf("#%d: %s: MisMatch: Idx:%d Want:%v Got:%v", i, tt.name, i, tt.data, buf[0:wCnt])
						return
					}
				}
			default:
				fmt.Printf("Test case not handled\n")
			}
		})
	}

}

func readerFromFixture(t *testing.T, path string) io.Reader {
	t.Helper()

	f, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		t.Fatalf("failed to open %s file for read: %v", path, err)
	}
	return f
}
func writerFromFixture(t *testing.T, path string) io.Writer {
	t.Helper()

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		t.Fatalf("failed to open %s file for write: %v", path, err)
	}
	return f
}

func TestFileReadWrite(t *testing.T) {
	type args struct {
		src io.Reader
		dst io.Writer
	}

	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Test box parsing",
			args: args{
				src: readerFromFixture(t, filepath.Join("testdata", "01_simple.mp4")),
				dst: writerFromFixture(t, filepath.Join("testdata", "01_simple_output.mp4")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := Parse(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			log.Printf(`
Ftyp: %+v\n
Moov: %+v\n
`, f.Ftyp, f.Moov)
		})
	}
}

//func Fixture(t *testing.T, path string) io.Reader {
// 	t.Helper()
//
// 	f, err := os.OpenFile(path, os.O_RDONLY, 0400)
// 	if err != nil {
// 		t.Fatalf("failed to open %s file: %v", path, err)
// 	}
//
// 	return f
// }
//
// func TestParse(t *testing.T) {
// 	type args struct {
// 		src io.Reader
// 	}
//
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []Box
// 		wantErr bool
// 	}{
// 		{
// 			name: "Test box parsing",
// 			args: args{
// 				src: readerFromFixture(t, filepath.Join("testdata", "01_simple.mp4")),
// 			},
// 			want: []Box{
// 				&box{
// 					boxtype: "ftyp",
// 					size:    int32(18),
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			f, err := Parse(tt.args.src)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			log.Printf(`
// Ftyp: %+v\n
// Moov: %+v\n
// `, f.Ftyp, f.Moov)
// 		})
// 	}
// }
