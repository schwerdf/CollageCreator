package CollageCreator

import (
	"flag"
	"math"
	"strings"
)

const (
	Uniform_Cropping   string = "Uniform_Cropping"
	Uniform_Scaling    string = "Uniform_Scaling"
	Uniform_ScaleToMin string = "Uniform_ScaleToMin"
)

// The simplest 'DimensionInitializer': sends all images through as-is.
type DimensionInitializer_Original struct{}

func (dio DimensionInitializer_Original) RegisterCustomParameters(parameters *Parameters) bool {
	return true
}

func (dio DimensionInitializer_Original) ParseCustomParameters(parameters *Parameters) bool {
	return true
}

func (dio DimensionInitializer_Original) InitializeDimensions(imageLayout ImageLayout) (il ImageLayout, err error) {
	il, err = imageLayout, nil
	return
}

type DimensionInitializer_Uniform_CustomParameters struct {
	cropping   string
	scaling    string
	scaleToMin string
}

type DimensionInitializer_Uniform_ScaleToMinParameter struct {
	x bool
	y bool
}

func DimensionInitializer_Uniform_Init() DimensionInitializer_Uniform {
	return DimensionInitializer_Uniform{new(DimensionInitializer_Uniform_CustomParameters)}
}

// A 'DimensionInitializer' that applies uniform cropping and scaling rules, specified as ImageMagick geometry strings,
// to all input images.
type DimensionInitializer_Uniform struct {
	p *DimensionInitializer_Uniform_CustomParameters
}

func (dio DimensionInitializer_Uniform) RegisterCustomParameters(parameters *Parameters) bool {
	flag.StringVar(&(dio.p.cropping), "crop", "", "Crop all images according to this geometry before processing")
	flag.StringVar(&(dio.p.scaling), "scale", "", "Scale all images according to this geometry before processing")
	flag.StringVar(&(dio.p.scaleToMin), "scale-to-min", "", "Scale all images to the dimensions of the smallest")
	return true
}

func (dio DimensionInitializer_Uniform) ParseCustomParameters(parameters *Parameters) bool {
	if dio.p.cropping == "" {
		parameters.SetOther(Uniform_Cropping, EmptyGeometry())
	} else {
		geometry, err := ParseGeometry(dio.p.cropping)
		if err != nil {
			parameters.ProgressMonitor().ReportMessage(err.Error())
			return false
		}
		parameters.SetOther(Uniform_Cropping, geometry)
	}
	if dio.p.scaleToMin != "" {
		scaler := strings.ToLower(dio.p.scaleToMin)
		switch scaler {
		case "xy":
			{
				parameters.SetOther(Uniform_ScaleToMin, DimensionInitializer_Uniform_ScaleToMinParameter{x: true, y: true})
			}
		case "y":
			{
				parameters.SetOther(Uniform_ScaleToMin, DimensionInitializer_Uniform_ScaleToMinParameter{x: false, y: true})
			}
		case "x":
			{
				parameters.SetOther(Uniform_ScaleToMin, DimensionInitializer_Uniform_ScaleToMinParameter{x: true, y: false})
			}
		default:
			{
				parameters.ProgressMonitor().ReportMessage("-scale-to-min value must be 'x', 'y', or 'xy'")
			}
		}
	}
	if dio.p.scaling == "" {
		parameters.SetOther(Uniform_Scaling, EmptyGeometry())
	} else {
		geometry, err := ParseGeometry(dio.p.scaling)
		if err != nil {
			parameters.ProgressMonitor().ReportMessage(err.Error())
			return false
		}
		parameters.SetOther(Uniform_Scaling, geometry)
	}
	return true
}

func (dio DimensionInitializer_Uniform) InitializeDimensions(imageLayout ImageLayout) (il ImageLayout, err error) {
	cropping := imageLayout.Parameters().OtherGeometry(Uniform_Cropping)
	scalingToMin, valid := imageLayout.Parameters().Other(Uniform_ScaleToMin)
	scaling := imageLayout.Parameters().OtherGeometry(Uniform_Scaling)
	if valid {
		switch scalingToMinO := scalingToMin.(type) {
		case DimensionInitializer_Uniform_ScaleToMinParameter:
			{
				minWidth := math.Inf(1)
				minHeight := math.Inf(1)
				for _, img := range imageLayout.Images(false) {
					if scalingToMinO.x {
						minWidth = math.Min(minWidth, imageLayout.ImageInfoOf(img).DimensionsOf().x)
					}
					if scalingToMinO.y {
						minHeight = math.Min(minHeight, imageLayout.ImageInfoOf(img).DimensionsOf().y)
					}
				}
				if minWidth < math.Inf(1) {
					scaling.width = GeometryDimension{minWidth, Pixels}
				}
				if minHeight < math.Inf(1) {
					scaling.height = GeometryDimension{minHeight, Pixels}
				}
			}
		default:
		}
	}
	for _, img := range imageLayout.Images(false) {
		imageLayout.SetCropping(img, cropping)
		imageLayout.SetScaling(img, scaling)
	}
	il, err = imageLayout, nil
	return
}
