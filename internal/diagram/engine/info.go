package engine

func SVGSize(svg []byte) (width, height int) {
	if m := svgSizeRE.FindSubmatch(svg); m != nil {
		return atoiOr(m[1], 0), atoiOr(m[2], 0)
	}
	if m := svgViewBoxRE.FindSubmatch(svg); m != nil {
		return atoiOr(m[1], 0), atoiOr(m[2], 0)
	}
	return 0, 0
}

func (d *D2Diagram) NumShapes() int {
	return len(d.target.Shapes)
}

func (d *D2Diagram) NumConnections() int {
	return len(d.target.Connections)
}
