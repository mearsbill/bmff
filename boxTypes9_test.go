package bmff

import (
	"bytes"
	"efmt"
	"fmt"
	"testing"
)

// here test the functionality of the basic box.
// this include paring and outputing.... pretty well tested for the basic box
func TestBoxTypes2(t *testing.T) {

	tests := []struct {
		name    string
		data    []byte
		want    []Box
		output  bool
		wantErr bool
	}{
		{"Test Emsg Sizes", []byte{0, 0, 0, 41, 'e', 'm', 's', 'g', 0, 0, 0, 0, 'u', 'r', 'i', 0, 'v', 'a', 'l', 'u', 'e', 0,
			0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4, 'x', 'y', 'z'}, nil, false, false},
		{"Test Emsg Input/Output Match", []byte{0, 0, 0, 41, 'e', 'm', 's', 'g', 0, 0, 0, 0, 'u', 'r', 'i', 0, 'v', 'a', 'l', 'u', 'e', 0,
			0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4, 'x', 'y', 'z'}, nil, false, false},
		{"Test Emsg Parse/Encode Match", []byte{0, 0, 0, 41, 'e', 'm', 's', 'g', 0, 0, 0, 0, 'u', 'r', 'i', 0, 'v', 'a', 'l', 'u', 'e', 0,
			0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4, 'x', 'y', 'z'}, nil, false, false},
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
				if int(b.size) != (len(tt.data)) {
					t.Errorf("#%d: %s: NewBox() bad box size=%d vs %d", idx, tt.name, (len(tt.data)), b.size)
					t.Errorf("#%d: %s: NewBox() bad raw size=%d vs %d", idx, tt.name, (len(tt.data) - 8), len(b.raw))
					return
				}
				if b.boxtype != "emsg" {
					t.Errorf("#%d: %s: NewBox() bad decod=%+v", idx, tt.name, b)
				}
			case (idx >= 1) && (idx <= 6):
				// testing Output matching input.
				// make an output buffer and use it as the output stream for the output mehtod
				// wierd behavior not conforming the docs.... so I have to use Truncate to get
				// the output buffer to behave
				buf := make([]byte, 128, 128)
				wB := bytes.NewBuffer(buf)
				wB.Truncate(0)
				b, _ := NewBox(reader, nt)
				eb := EmsgBox{box: b}
				fmt.Printf("Emsg Parse Test %d: ", idx)
				eb.parse()
				// eb.PrintDetail()
				eb.Encode()
				wCnt, err := eb.Output(wB, 0)
				if err != nil {
					t.Errorf("#%d: %s: Output error: %v", idx, tt.name, err)
					return
				}
				if wCnt != len(tt.data) {
					t.Errorf("#%d: OutputSize:%d eMsgSize:%d\n", idx, wCnt, len(tt.data))
				}
				for i := 0; i < wCnt; i++ {
					if buf[i] != tt.data[i] {
						t.Errorf("#%d: %s: MisMatch: Idx:%d Want:%v Got:%v", idx, tt.name, i, tt.data, buf[0:wCnt])
						return
					}
				}
			default:
				fmt.Printf("Test case not handled\n")
			}
		})
	}

}

// func compareFiles(r, w *os.File) (firstDiff int, err error) {
// 	r.Seek(0, 0)
// 	w.Seek(0, 0)
// 	cCnt := 0
// 	readSize := 8192 * 4
// 	b1 := make([]byte, readSize)
// 	b2 := make([]byte, readSize)
// 	for {
// 		rCount, rErr := r.Read(b1)
// 		if rErr != nil {
// 			return cCnt, fmt.Errorf("Compare read error(r) @ offset %d: %v", cCnt, rErr)
// 			return
// 		}
// 		wCount, wErr := w.Read(b2)
// 		if wErr != nil {
// 			return cCnt, fmt.Errorf("Compare read error(3) @ offset %d: %v", cCnt, wErr)
// 			return
// 		}
// 		if wCount != rCount {
// 			return wCount, fmt.Errorf("File lengths differ.. r:%d  w:%d", cCnt+rCount, cCnt+wCount)
// 		}
// 		if rCount > 0 {
// 			for idx := 0; idx < rCount; idx++ {
// 				if b1[idx] != b2[idx] {
// 					return cCnt + idx, fmt.Errorf("Files differ (r:%02x vs w:%02x) starting at offset %d",
// 						b1[idx], b2[idx], cCnt+idx)
// 				}
// 			}
// 		}
// 		cCnt += rCount
// 		if rCount == readSize {
// 			return cCnt, nil
// 		}
// 	}
// }
//
// func TestSmallFile(t *testing.T) {
// 	type args struct {
// 		src string
// 		dst string
// 	}
//
// 	tests := []struct {
// 		name     string
// 		src      string
// 		dst      string
// 		maxDepth int
// 	}{
// 		{
// 			name:     "Test Read/Write Matching at all box depths",
// 			src:      filepath.Join("testdata", "01_simple.mp4"),
// 			dst:      filepath.Join("testdata", "01_simple_output.mp4"),
// 			maxDepth: 6,
// 		},
// 		{
// 			name:     "Test Read/Write Matching depth=1",
// 			src:      "/Users/bmears/videoClips/seaworld.mp4",
// 			dst:      "/tmp/searworld_out.mp4",
// 			maxDepth: 6,
// 		},
// 		{
// 			name:     "Test Read/Write Matching depth=1",
// 			src:      "/Users/bmears/videoClips/chunk_ctaudio_cfm4s_ridp0aa0br88560_cs46262195568_w1932408732_mpd.m4s",
// 			dst:      "/tmp/chunk_ctaudio_cfm4s_ridp0aa0br88560_cs46262195568_w1932408732_mpd_out.mp4",
// 			maxDepth: 6,
// 		},
// 		{
// 			name:     "Test Read/Write Matching depth=1",
// 			src:      "/Users/bmears/videoClips/chunk_ctvideo_cfm4s_ridp0va0br41991_cs86746737600_w1932408732_mpd.m4s",
// 			dst:      "/tmp/chunk_ctvideo_cfm4s_ridp0va0br41991_cs86746737600_w1932408732_mpd_out.mp4",
// 			maxDepth: 6,
// 		},
// 	}
// 	// use these vars across tests ... some are carried between sequential tests
// 	var bF0 *File_s
// 	var fileSize int
// 	for idx, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			switch {
// 			case idx >= 0 && idx <= 99:
// 				for depth := 1; depth <= tt.maxDepth; depth++ {
// 					// open source and destination files
// 					rF, err := os.OpenFile(tt.src, os.O_RDONLY, 0400)
// 					if err != nil {
// 						t.Fatalf("failed to open %s file for read: %v", tt.src, err)
// 					}
// 					wF, err := os.OpenFile(tt.dst, os.O_RDWR|os.O_CREATE, 0600)
// 					if err != nil {
// 						t.Fatalf("failed to open %s file for write: %v", tt.dst, err)
// 					}
// 					bF0, err = Parse(rF)
// 					if err != nil {
// 						t.Errorf("Box Parse() error = %v", err)
// 						return
// 					}
// 					fileSize, err = bF0.Output(wF, depth) // just the toplevel files
//
// 					_, cErr := compareFiles(rF, wF)
// 					if cErr != nil {
// 						t.Errorf("File compare error = %v ", cErr)
// 						return
// 					}
// 					rF.Close()
// 					wF.Close()
// 				}
//
// 			default:
// 				fmt.Printf("Unhandled Test case %d", idx)
// 			}
// 		})
// 	}
// }
