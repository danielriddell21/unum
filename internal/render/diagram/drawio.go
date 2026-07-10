package diagram

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"

	"oss.terrastruct.com/d2/d2target"
)

type mxFile struct {
	XMLName xml.Name  `xml:"mxfile"`
	Host    string    `xml:"host,attr"`
	Diagram mxDiagram `xml:"diagram"`
}

type mxDiagram struct {
	Name  string       `xml:"name,attr"`
	ID    string       `xml:"id,attr"`
	Model mxGraphModel `xml:"mxGraphModel"`
}

type mxGraphModel struct {
	Grid   int    `xml:"grid,attr"`
	Root   mxRoot `xml:"root"`
	Arrows int    `xml:"arrows,attr"`
}

type mxRoot struct {
	Cells []mxCell `xml:"mxCell"`
}

type mxCell struct {
	ID       string      `xml:"id,attr"`
	Value    string      `xml:"value,attr,omitempty"`
	Style    string      `xml:"style,attr,omitempty"`
	Vertex   string      `xml:"vertex,attr,omitempty"`
	Edge     string      `xml:"edge,attr,omitempty"`
	Parent   string      `xml:"parent,attr,omitempty"`
	Source   string      `xml:"source,attr,omitempty"`
	Target   string      `xml:"target,attr,omitempty"`
	Geometry *mxGeometry `xml:"mxGeometry,omitempty"`
}

type mxGeometry struct {
	X        int    `xml:"x,attr,omitempty"`
	Y        int    `xml:"y,attr,omitempty"`
	Width    int    `xml:"width,attr,omitempty"`
	Height   int    `xml:"height,attr,omitempty"`
	Relative string `xml:"relative,attr,omitempty"`
	As       string `xml:"as,attr"`
}

func marshalDrawio(cells []mxCell) ([]byte, error) {
	cells = append([]mxCell{
		{ID: "0"},
		{ID: "1", Parent: "0"},
	}, cells...)
	doc := mxFile{
		Host: "unum",
		Diagram: mxDiagram{
			Name:  "Page-1",
			ID:    "unum-diagram",
			Model: mxGraphModel{Grid: 1, Root: mxRoot{Cells: cells}},
		},
	}
	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal drawio: %w", err)
	}
	return append([]byte(xml.Header), out...), nil
}

func drawioFromD2(d *d2target.Diagram) ([]byte, error) {
	cells := make([]mxCell, 0, len(d.Shapes)+len(d.Connections))
	for _, s := range d.Shapes {
		label := s.Label
		if label == "" {
			label = s.ID
		}
		cells = append(cells, mxCell{
			ID:     s.ID,
			Value:  label,
			Style:  "rounded=0;whiteSpace=wrap;html=1;",
			Vertex: "1",
			Parent: "1",
			Geometry: &mxGeometry{
				X: s.Pos.X, Y: s.Pos.Y, Width: s.Width, Height: s.Height, As: "geometry",
			},
		})
	}
	for i, c := range d.Connections {
		cells = append(cells, mxCell{
			ID:       fmt.Sprintf("edge-%d", i),
			Value:    c.Label,
			Style:    "edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;",
			Edge:     "1",
			Parent:   "1",
			Source:   c.Src,
			Target:   c.Dst,
			Geometry: &mxGeometry{Relative: "1", As: "geometry"},
		})
	}
	return marshalDrawio(cells)
}

var svgSizeRE = regexp.MustCompile(`<svg[^>]*\bwidth="(\d+)(?:\.\d+)?(?:px)?"[^>]*\bheight="(\d+)(?:\.\d+)?(?:px)?"`)

func DrawioFromSVG(svg []byte) ([]byte, error) {
	width, height := 640, 480
	if m := svgSizeRE.FindSubmatch(svg); m != nil {
		if w, err := strconv.Atoi(string(m[1])); err == nil && w > 0 {
			width = w
		}
		if h, err := strconv.Atoi(string(m[2])); err == nil && h > 0 {
			height = h
		}
	}
	dataURI := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(svg)
	cells := []mxCell{{
		ID:       "image-0",
		Style:    "shape=image;verticalLabelPosition=bottom;labelBackgroundColor=#ffffff;verticalAlign=top;imageAspect=1;aspect=fixed;image=" + dataURI + ";",
		Vertex:   "1",
		Parent:   "1",
		Geometry: &mxGeometry{X: 0, Y: 0, Width: width, Height: height, As: "geometry"},
	}}
	return marshalDrawio(cells)
}
