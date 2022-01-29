package CollageCreator

import (
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
)

// Holds a pair of floats representing dimensions or coordinates.
type Dims struct {
	x, y float64
}

// Creates a new 'Dims' object.
func NewDims(x, y float64) Dims { return Dims{x, y} }

// Gets the X coordinate from a 'Dims' object.
func (d Dims) X() float64 { return d.x }

// Sets the X coordinate in a 'Dims' object.
func (d *Dims) SetX(val float64) { d.x = val }

// Gets the Y coordinate from a 'Dims' object.
func (d Dims) Y() float64 { return d.y }

// Sets the Y coordinate in a 'Dims' object.
func (d *Dims) SetY(val float64) { d.y = val }

// Gets the coordinate in a 'Dims' object indicated by the parameter: X if even, Y if odd.
func (d Dims) Dim(i int) float64 {
	if i%2 == 0 {
		return d.x
	} else {
		return d.y
	}
}
func toInt(f float64) int {
	return int(f)
}
func toIntP(f float64) int {
	return int(math.Round(f))
}

// Sets the coordinate in a 'Dims' object indicated by the parameter: X if even, Y if odd.
func (d *Dims) SetDim(i int, val float64) {
	if i%2 == 0 {
		d.x = val
	} else {
		d.y = val
	}
}

// A front-end to 'ParseDims' that logs a fatal error if the string does not parse.
func MustParseDims(arg string) Dims {
	rv, err := ParseDims(arg)
	if err != nil {
		log.Fatal(err)
	}
	return rv
}

// Parses a string into a 'Dims' object. The string must consist of either one
// non-negative integer, or two non-negative integers separated by a comma or 'x'.
// Returns the parsed object, and an error if the string is malformed.
func ParseDims(arg string) (d Dims, err error) {
	var dims_regex *regexp.Regexp = regexp.MustCompile(`^([0-9]+)?([,x]([0-9]*))?$`)
	if dims_regex.MatchString(arg) && len(arg) != 0 && ((arg[0] == '(') == (arg[len(arg)-1] == ')')) {
		sm := dims_regex.FindStringSubmatch(arg)
		x, errx := strconv.Atoi(sm[1])
		y, erry := strconv.Atoi(sm[3])
		if errx == nil && erry == nil {
			d = NewDims(float64(x), float64(y))
		} else if errx == nil {
			if sm[2] == "" {
				d = NewDims(float64(x), float64(x))
			} else {
				d = NewDims(float64(x), 0)
			}
		} else if erry == nil {
			d = NewDims(0, float64(y))
		}
		err = nil
		return
	}
	d = NewDims(0, 0)
	err = errors.New("Malformed dimension string: '" + arg + "'")
	return
}

// Units supported by the 'Geometry' type.
type GeometryUnits int

const (
	Pixels GeometryUnits = iota
	Percent
)

// A unit-bearing image measurement.
type GeometryDimension struct {
	N float64
	U GeometryUnits
}

// Types of scaling supported by the 'Geometry' type.
type GeometryScaling int

const (
	// Always scale, no matter the size of the image to be scaled.
	ScaleAlways GeometryScaling = iota
	// Scale only if the image is larger than the given size.
	ScaleDownOnly
	// Scale only if the image is smaller than the given size.
	ScaleUpOnly
)

// Represents the geometry of an image operation such as cropping or scaling, as parsed from an ImageMagick-like "geometry" string.
type Geometry struct {
	// For cropping, holds the width of the crop-box; for scaling, holds the horizontal dimension of the scaled image.
	width GeometryDimension
	// For cropping, holds the height of the crop-box; for scaling, holds the vertical dimension of the scaled image.
	height GeometryDimension
	// For cropping, holds the horizontal offset of the upper left corner of the crop-box.
	x GeometryDimension
	// For cropping, holds the vertical offset of the upper left corner of the crop-box.
	y GeometryDimension
	// Holds the aspect ratio
	//aspectRatio float64
	// When this flag is set, a scaling operation will preserve the original aspect ratio of the image being scaled.
	// When it is not set, the image will be scaled to the exact dimensions provided regardless of aspect ratio.
	preserveAspectRatio bool
	// The scaling conditions to be used.
	scaling GeometryScaling
}

const (
	floatregex                 string = `([0-9]+(\.[0-9]+)?)?`
	geom_regex                 string = `^` + floatregex + `(x` + floatregex + `)?([+-]` + floatregex + `)?([+-]` + floatregex + `)?([%]?)([!]?)([<>]?)$`
	geom_regex_width           int    = 1
	geom_regex_height          int    = 4
	geom_regex_x               int    = 6
	geom_regex_y               int    = 9
	geom_regex_scaleunits      int    = 12
	geom_regex_preserve_aspect int    = 13
	geom_regex_scaling         int    = 14
)

func parseFloatDim(arg string) (rv float64, err error) {
	if arg[0] == '+' {
		rv, err = strconv.ParseFloat(arg[1:], 64)
	} else {
		rv, err = strconv.ParseFloat(arg, 64)
	}
	return
}

func printFloat(val float64) string {
	if whole, frac := math.Modf(val); frac < 1e-5 {
		return fmt.Sprintf("%d", int(whole))
	} else {
		return fmt.Sprintf("%f", val)
	}
}

// A front-end to 'ParseGeometry' that logs a fatal error if the string does not parse.
func MustParseGeometry(arg string) Geometry {
	rv, err := ParseGeometry(arg)
	if err != nil {
		log.Fatal(err)
	}
	return rv
}

// Generates an "empty" 'Geometry' object in which all numeric values are 'NaN'.
func EmptyGeometry() Geometry {
	return Geometry{width: GeometryDimension{math.NaN(), Pixels},
		height:              GeometryDimension{math.NaN(), Pixels},
		x:                   GeometryDimension{math.NaN(), Pixels},
		y:                   GeometryDimension{math.NaN(), Pixels},
		preserveAspectRatio: true,
		scaling:             ScaleAlways}
}

// Parses a string into a 'Geometry' object. The string must follow the format
// of an ImageMagick 'geometry' parameter.
// Returns the parsed object, and an error if the string is malformed.
func ParseGeometry(arg string) (geom Geometry, err error) {
	var geom_regex *regexp.Regexp = regexp.MustCompile(geom_regex)
	var ff float64
	if geom_regex.MatchString(arg) && len(arg) != 0 {
		geom = Geometry{}
		sm := geom_regex.FindStringSubmatch(arg)
		var units GeometryUnits
		if sm[geom_regex_scaleunits] == "%" {
			units = Percent
		} else {
			units = Pixels
		}
		if sm[geom_regex_width] != "" {
			ff, err = parseFloatDim(sm[geom_regex_width])
			if err != nil {
				return
			}
			geom.width = GeometryDimension{ff, units}
		} else {
			geom.width = GeometryDimension{math.NaN(), units}
		}
		if sm[geom_regex_height] != "" {
			ff, err = parseFloatDim(sm[geom_regex_height])
			if err != nil {
				return
			}
			geom.height = GeometryDimension{ff, units}
		} else {
			geom.height = GeometryDimension{math.NaN(), units}
		}
		if sm[geom_regex_x+1] != "" {
			ff, err = parseFloatDim(sm[geom_regex_x])
			if err != nil {
				return
			}
			geom.x = GeometryDimension{ff, units}
		} else {
			geom.x = GeometryDimension{math.NaN(), units}
		}
		if sm[geom_regex_y+1] != "" {
			ff, err = parseFloatDim(sm[geom_regex_y])
			if err != nil {
				return
			}
			geom.y = GeometryDimension{ff, units}
		} else {
			geom.y = GeometryDimension{math.NaN(), units}
		}
		geom.preserveAspectRatio = (sm[geom_regex_preserve_aspect] != "!")
		switch sm[geom_regex_scaling] {
		case "<":
			geom.scaling = ScaleUpOnly
		case ">":
			geom.scaling = ScaleDownOnly
		default:
			geom.scaling = ScaleAlways
		}
	} else {
		err = errors.New("Malformed geometry string: '" + arg + "'")
		return
	}
	err = nil
	return
}

// Converts a 'Geometry' object into a string.
func (geom Geometry) String() string {
	var rv string = ""
	if geom.HasWidth() {
		rv += printFloat(geom.width.N)
	}
	if geom.HasHeight() {
		rv += fmt.Sprintf("x%s", printFloat(geom.height.N))
	}
	if geom.HasX() {
		if geom.x.N >= 0.0 {
			rv += "+"
		}
		rv += printFloat(geom.x.N)
	}
	if geom.HasY() {
		if math.IsNaN(geom.x.N) {
			rv += "+"
		}
		if geom.y.N >= 0.0 {
			rv += "+"
		}
		rv += printFloat(geom.y.N)
	}
	switch geom.width.U {
	case Percent:
		rv += "%"
	}
	if !geom.PreserveAspectRatio() {
		rv += "!"
	}
	switch geom.scaling {
	case ScaleDownOnly:
		rv += ">"
	case ScaleUpOnly:
		rv += "<"
	default:
	}
	return rv
}

// Returns true if this 'Geometry' object explicitly specifies a width.
func (geom Geometry) HasWidth() bool {
	return !math.IsNaN(geom.width.N)
}

// Returns true if this 'Geometry' object explicitly specifies a height.
func (geom Geometry) HasHeight() bool {
	return !math.IsNaN(geom.height.N)
}

// Returns true if this 'Geometry' object explicitly specifies either a width or a height.
func (geom Geometry) HasSize() bool {
	return geom.HasWidth() || geom.HasHeight()
}

// Returns true if this 'Geometry' object explicitly specifies a horizontal offset.
func (geom Geometry) HasX() bool {
	return !math.IsNaN(geom.x.N)
}

// Returns true if this 'Geometry' object explicitly specifies a vertical offset.
func (geom Geometry) HasY() bool {
	return !math.IsNaN(geom.y.N)
}

// Returns true if this 'Geometry' object explicitly specifies either a horizontal or a vertical offset.
func (geom Geometry) HasOffset() bool {
	return geom.HasX() || geom.HasY()
}

// Returns true if this 'Geometry' object specifies preservation of the aspect ratio.
func (geom Geometry) PreserveAspectRatio() bool {
	return geom.preserveAspectRatio
}

// Calculate the offset of this 'Geometry' object relative to an image of the given size.
// 'fullSize' is taken into account only when the 'Percent' relative units are used.
func (geom Geometry) Offset(fullSize Dims) Dims {
	switch geom.x.U {
	case Pixels:
		x := 0.0
		if geom.HasX() {
			x = math.Max(0, math.Min(geom.x.N, fullSize.X()))
		}
		y := 0.0
		if geom.HasY() {
			y = math.Max(0, math.Min(geom.y.N, fullSize.Y()))
		}
		return NewDims(x, y)
	case Percent:
		x := 0.0
		if geom.HasX() {
			x = math.Max(0, math.Min(fullSize.X()*geom.x.N/100.0, fullSize.X()))
		}
		y := 0.0
		if geom.HasY() {
			y = math.Max(0, math.Min(fullSize.Y()*geom.y.N/100.0, fullSize.Y()))
		}
		return NewDims(x, y)

	}
	return NewDims(0, 0)
}

// Calculate the size of an image cropped using this 'Geometry' object, relative to an
// image of the given size. 'fullSize' is taken into account only when the 'Percent'
// relative units are used.
func (geom Geometry) Crop(fullSize Dims) Dims {
	if !geom.HasOffset() {
		return fullSize
	}
	topLeft := geom.Offset(fullSize)
	var xDim, yDim float64
	switch geom.width.U {
	case Pixels:
		xDim = math.Min(geom.width.N, fullSize.X()-topLeft.X())
		yDim = math.Min(geom.height.N, fullSize.Y()-topLeft.Y())
	case Percent:
		xDim = math.Min(fullSize.X()*geom.width.N/100, fullSize.X()-topLeft.X())
		yDim = math.Min(fullSize.Y()*geom.height.N/100, fullSize.Y()-topLeft.Y())
	}
	return NewDims(xDim, yDim)
}

// Calculate the size of an image scaled using this 'Geometry' object, relative to an
// image of the given size. 'fullSize' is taken into account only when the 'Percent'
// relative units are used.
func (geom Geometry) Scale(fullSize Dims) Dims {
	var xDim, yDim float64
	var units GeometryUnits
	if geom.HasWidth() && geom.HasHeight() {
		if geom.width.U == geom.height.U {
			units = geom.width.U
		} else {
			return fullSize
		}
	} else if geom.HasWidth() {
		units = geom.width.U
	} else if geom.HasHeight() {
		units = geom.height.U
	} else {
		return fullSize
	}
	if geom.PreserveAspectRatio() {
		switch units {
		case Percent:
			if !geom.HasWidth() {
				xDim = fullSize.X() * geom.height.N / 100
				yDim = fullSize.Y() * geom.height.N / 100
			} else if !geom.HasHeight() {
				xDim = fullSize.X() * geom.width.N / 100
				yDim = fullSize.Y() * geom.width.N / 100
			} else {
				xDim = fullSize.X() * math.Min(geom.width.N, geom.height.N) / 100
				yDim = fullSize.Y() * math.Min(geom.width.N, geom.height.N) / 100
			}
		case Pixels:
			aspect := float64(fullSize.X()) / float64(fullSize.Y())
			if !geom.HasWidth() {
				xDim = aspect * geom.height.N
				yDim = geom.height.N
			} else if !geom.HasHeight() {
				xDim = geom.width.N
				yDim = geom.width.N / aspect
			} else {
				xDim = math.Min(geom.width.N, geom.height.N*aspect)
				yDim = math.Min(geom.height.N, geom.width.N/aspect)
			}
		}
	} else {
		switch units {
		case Percent:
			if geom.HasWidth() {
				xDim = fullSize.X() * geom.width.N / 100
			} else {
				xDim = fullSize.X()
			}
			if geom.HasHeight() {
				xDim = fullSize.Y() * geom.height.N / 100
			} else {
				yDim = fullSize.Y()
			}
		case Pixels:
			if geom.HasWidth() {
				xDim = geom.width.N
			} else {
				xDim = fullSize.X()
			}
			if geom.HasHeight() {
				yDim = geom.height.N
			} else {
				yDim = fullSize.Y()
			}
		}
	}
	return NewDims(xDim, yDim)
}

// Apply both the 'Scale' and 'Crop' methods to a given image, in that order.
func ScaleAndCrop(original Dims, cropping Geometry, scaling Geometry) Dims {
	return cropping.Crop(scaling.Scale(original))
}
