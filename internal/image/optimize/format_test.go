package optimize

import "testing"

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		want   Format
		wantOK bool
	}{
		{"jpeg", "jpeg", FormatJPEG, true},
		{"jpg alias", "jpg", FormatJPEG, true},
		{"png", "png", FormatPNG, true},
		{"gif", "gif", FormatGIF, true},
		{"webp", "webp", FormatWebP, true},
		{"tiff", "tiff", FormatTIFF, true},
		{"tif alias", "tif", FormatTIFF, true},
		{"bmp", "bmp", FormatBMP, true},
		{"uppercase", "PNG", FormatPNG, true},
		{"padded", "  jpeg  ", FormatJPEG, true},
		{"unknown", "avif", FormatUnknown, false},
		{"empty", "", FormatUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseFormat(tt.in)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("ParseFormat(%q) = (%v, %v), want (%v, %v)", tt.in, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestFormatStringExtMIME(t *testing.T) {
	tests := []struct {
		format          Format
		name, ext, mime string
	}{
		{FormatJPEG, "jpeg", ".jpg", "image/jpeg"},
		{FormatPNG, "png", ".png", "image/png"},
		{FormatGIF, "gif", ".gif", "image/gif"},
		{FormatWebP, "webp", ".webp", "image/webp"},
		{FormatTIFF, "tiff", ".tiff", "image/tiff"},
		{FormatBMP, "bmp", ".bmp", "image/bmp"},
		{FormatUnknown, "unknown", "", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.String(); got != tt.name {
				t.Errorf("String() = %q, want %q", got, tt.name)
			}
			if got := tt.format.Ext(); got != tt.ext {
				t.Errorf("Ext() = %q, want %q", got, tt.ext)
			}
			if got := tt.format.MIME(); got != tt.mime {
				t.Errorf("MIME() = %q, want %q", got, tt.mime)
			}
		})
	}
}

func TestFormatCanEncode(t *testing.T) {
	encodable := map[Format]bool{FormatJPEG: true, FormatPNG: true, FormatGIF: true, FormatWebP: true}
	all := []Format{FormatUnknown, FormatJPEG, FormatPNG, FormatGIF, FormatWebP, FormatTIFF, FormatBMP}

	for _, f := range all {
		if got := f.CanEncode(); got != encodable[f] {
			t.Errorf("%v.CanEncode() = %v, want %v", f, got, encodable[f])
		}
	}
}

func TestFormatLossy(t *testing.T) {
	if !FormatJPEG.Lossy() {
		t.Error("jpeg should report as lossy")
	}
	for _, f := range []Format{FormatPNG, FormatGIF, FormatWebP, FormatUnknown} {
		if f.Lossy() {
			t.Errorf("%v should not report as lossy", f)
		}
	}
}

func TestResolveOutputFormat(t *testing.T) {
	tests := []struct {
		name          string
		explicit, src Format
		want          Format
	}{
		{"explicit wins", FormatPNG, FormatJPEG, FormatPNG},
		{"falls back to encodable source", FormatUnknown, FormatJPEG, FormatJPEG},
		{"gif source kept", FormatUnknown, FormatGIF, FormatGIF},
		{"webp source kept", FormatUnknown, FormatWebP, FormatWebP},
		{"bmp source becomes png", FormatUnknown, FormatBMP, FormatPNG},
		{"unknown source becomes png", FormatUnknown, FormatUnknown, FormatPNG},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveOutputFormat(tt.explicit, tt.src); got != tt.want {
				t.Errorf("ResolveOutputFormat(%v, %v) = %v, want %v", tt.explicit, tt.src, got, tt.want)
			}
		})
	}
}
