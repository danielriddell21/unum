package optimize

import "strings"

type Format int

const (
	FormatUnknown Format = iota
	FormatJPEG
	FormatPNG
	FormatGIF
	FormatWebP
	FormatTIFF
	FormatBMP
)

func (f Format) String() string {
	switch f {
	case FormatJPEG:
		return "jpeg"
	case FormatPNG:
		return "png"
	case FormatGIF:
		return "gif"
	case FormatWebP:
		return "webp"
	case FormatTIFF:
		return "tiff"
	case FormatBMP:
		return "bmp"
	default:
		return "unknown"
	}
}

func (f Format) Ext() string {
	switch f {
	case FormatJPEG:
		return ".jpg"
	case FormatPNG:
		return ".png"
	case FormatGIF:
		return ".gif"
	case FormatWebP:
		return ".webp"
	case FormatTIFF:
		return ".tiff"
	case FormatBMP:
		return ".bmp"
	default:
		return ""
	}
}

func (f Format) MIME() string {
	switch f {
	case FormatJPEG:
		return "image/jpeg"
	case FormatPNG:
		return "image/png"
	case FormatGIF:
		return "image/gif"
	case FormatWebP:
		return "image/webp"
	case FormatTIFF:
		return "image/tiff"
	case FormatBMP:
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}

func (f Format) CanEncode() bool {
	return f == FormatJPEG || f == FormatPNG || f == FormatGIF || f == FormatWebP
}

func (f Format) Lossy() bool {
	return f == FormatJPEG
}

func ParseFormat(s string) (Format, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "jpeg", "jpg":
		return FormatJPEG, true
	case "png":
		return FormatPNG, true
	case "gif":
		return FormatGIF, true
	case "webp":
		return FormatWebP, true
	case "tiff", "tif":
		return FormatTIFF, true
	case "bmp":
		return FormatBMP, true
	}
	return FormatUnknown, false
}

func EncodableFormats() []Format {
	return []Format{FormatJPEG, FormatPNG, FormatGIF, FormatWebP}
}

func formatFromDecoded(name string) Format {
	f, _ := ParseFormat(name)
	return f
}
