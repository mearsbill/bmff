package bmff

import (
	"testing"
)

func TestUint16_16_String(t *testing.T) {
	tests := []struct {
		name string
		x    Uint16_16
		want string
	}{
		{"1.0", Uint16_16(0x00010000), "1.000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.String(); got != tt.want {
				t.Errorf("Uint16_16.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt16_16_String(t *testing.T) {
	tests := []struct {
		name string
		x    Int16_16
		want string
	}{
		{"2.0", Int16_16(0x00010000), "1.000"},
		{"2.1", Int16_16(-1 << 16), "-1.000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.String(); got != tt.want {
				t.Errorf("Int16_16.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint16_16_UnmarshalBinary(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		x       Uint16_16
		args    args
		wantErr bool
		want    string
	}{
		{"3.1", Uint16_16(0), args{[]byte{0x00, 0x01, 0x00, 0x00}}, false, "1.000"},
		{"3.2", Uint16_16(0), args{[]byte{0x00, 0x01, 0x80, 0x00}}, false, "1.500"},
		{"3.3", Uint16_16(0), args{[]byte{0x01, 0x01, 0x20, 0x00}}, false, "257.125"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.x.UnmarshalBinary(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Uint16_16.UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.x.String() != tt.want {
				t.Errorf("Uint16_16.UnmarshalBinary() string = %v, want %v", tt.x.String(), tt.want)
			}
		})
	}
}
func TestInt16_16_UnmarshalBinary(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		x       Int16_16
		args    args
		wantErr bool
		want    string
	}{
		{"4.1", Int16_16(0), args{[]byte{0x00, 0x01, 0x40, 0x00}}, false, "1.250"},
		{"4.2", Int16_16(0), args{[]byte{0xff, 0xff, 0x80, 0x00}}, false, "-0.500"},
		{"4.3", Int16_16(0), args{[]byte{0x0ff, 0xfe, 0xe6, 0x80}}, false, "-1.100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.x.UnmarshalBinary(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Uint16_16.UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.x.String() != tt.want {
				t.Errorf("Uint16_16.UnmarshalBinary() string = %v, want %v", tt.x.String(), tt.want)
			}
		})
	}
}

func TestUint8_8_String(t *testing.T) {
	tests := []struct {
		name string
		x    Uint8_8
		want string
	}{
		{"5.0", Uint8_8(0x0100), "1.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.String(); got != tt.want {
				t.Errorf("Uint8_8.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint8_8_UnmarshalBinary(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		x       Uint8_8
		args    args
		wantErr bool
		want    string
	}{
		{"6.1", Uint8_8(0), args{[]byte{0x01, 0x00}}, false, "1.00"},
		{"6.2", Uint8_8(0), args{[]byte{0x01, 0x40}}, false, "1.25"},
		{"6.3", Uint8_8(0), args{[]byte{0x10, 0x11}}, false, "16.07"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.x.UnmarshalBinary(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Uint8_8.UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.x.String() != tt.want {
				t.Errorf("Uint16_16.UnmarshalBinary() string = %v, want %v", tt.x.String(), tt.want)
			}
		})
	}
}
func TestInt8_8_String(t *testing.T) {
	tests := []struct {
		name string
		x    Int8_8
		want string
	}{
		{"7.0", Int8_8(0x0100), "1.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.String(); got != tt.want {
				t.Errorf("Iint8_8.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt8_8_UnmarshalBinary(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		x       Int8_8
		args    args
		wantErr bool
		want    string
	}{
		{"8.1", Int8_8(0), args{[]byte{0x01, 0x00}}, false, "1.00"},
		{"8.2", Int8_8(0), args{[]byte{0xff, 0x00}}, false, "-1.00"},
		{"8.3", Int8_8(0), args{[]byte{0xff, 0xe6}}, false, "-0.10"},
		{"8.4", Int8_8(0), args{[]byte{0x01, 0x40}}, false, "1.25"},
		{"8.5", Int8_8(0), args{[]byte{0xfe, 0xc0}}, false, "-1.25"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.x.UnmarshalBinary(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Int8_8.UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.x.String() != tt.want {
				t.Errorf("Int16_16.UnmarshalBinary() string = %v, want %v", tt.x.String(), tt.want)
			}
		})
	}
}
