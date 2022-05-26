package CollageCreator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Produces output in the form of a Bash script that, when run, will call ImageMagick to produce
// the output collage.
type OutputImage_ImageMagickScript struct {
	contents string
}

func (ois OutputImage_ImageMagickScript) WriteToFile(fileName string, parameters *Parameters) {
	err := os.WriteFile(fileName, []byte(ois.contents), 0666)
	if err != nil {
		log.Fatal(err)
	}
	parameters.ProgressMonitor().ReportOutputSuccess(fileName)
}

func shellScriptDefang(str string) string {
	return "'" + strings.ReplaceAll(str, "'", "'\"'\"'") + "'"
}

func createCollageImageMagickScript(imageLayout ImageLayout) string {
	rv := ""
	xAdd := 0.0
	yAdd := 0.0
	xSize := 0.0
	ySize := 0.0
	if imageLayout.CanvasSize() != NewDims(0, 0) {
		xSize, ySize = imageLayout.CanvasSize().X(), imageLayout.CanvasSize().Y()
	} else {
		for _, img := range imageLayout.Images(false) {
			if xAdd+imageLayout.PositionOf(img).X()+imageLayout.DimensionsOf(img).X() > xSize {
				xSize = xAdd + imageLayout.PositionOf(img).X() + imageLayout.DimensionsOf(img).X()
			}
			if yAdd+imageLayout.PositionOf(img).Y()+imageLayout.DimensionsOf(img).Y() > ySize {
				ySize = yAdd + imageLayout.PositionOf(img).Y() + imageLayout.DimensionsOf(img).Y()
			}
		}
	}

	rv += "#!/bin/bash\n\n"
	rv += "OUTFILE=$1\n"
	rv += "[ -z $OUTFILE ] && OUTFILE='Collage.jpg'\n"
	rv += "[ -z $IM_CONVERT_BIN ] && IM_CONVERT_BIN='convert'\n"
	rv += "[ -z $IM_COMPOSITE_BIN ] && IM_COMPOSITE_BIN='composite'\n"
	rv += fmt.Sprintf("convert 'xc:black[%dx%d!]' -colorspace sRGB -type truecolor \"$OUTFILE\"\n\n", toInt(xSize), toInt(ySize))

	i := 1
	// TODO: This is extremely slow as it writes the entire canvas image once for each image placed onto it.
	// Find a way to carve the image up into smaller "tiles" that can be composed and put onto the output image as one unit.
	for _, img := range imageLayout.Images(false) {
		imageLayout.Parameters().ProgressMonitor().ReportRenderingProgress(i, imageLayout.PositionedImageCount())
		info := imageLayout.ImageInfoOf(img)
		dimensions := info.DimensionsOf()
		scaling := imageLayout.ScalingOf(img)
		imageFile, err := filepath.Abs(imageLayout.ImageInfoOf(img).FileName())
		if err != nil {
			imageFile = imageLayout.ImageInfoOf(img).FileName()
		}
		rv += fmt.Sprintf("\"$IM_CONVERT_BIN\" %s", shellScriptDefang(imageFile))
		// -scale %dx%d! -crop %dx%d+%d+%d - | \"$IM_COMPOSITE_BIN\" -compose atop -geometry +%d+%d - \"$OUTFILE\" \"$OUTFILE\"\n", toInt(dimensions.X()), toInt(dimensions.Y()))
		if scaling.HasSize() {
			dimensions = scaling.Scale(dimensions)
			rv += fmt.Sprintf(" -scale \"%dx%d!\"", toIntP(dimensions.X()), toIntP(dimensions.Y()))
		}
		cropping := imageLayout.CroppingOf(img)
		offset := Dims{0, 0}
		if cropping.HasOffset() {
			offset = cropping.Offset(dimensions)
			dimensions = cropping.Crop(dimensions)
			rv += fmt.Sprintf(" -crop \"%dx%d+%d+%d\"", toIntP(dimensions.X()), toIntP(dimensions.Y()), toIntP(offset.X()), toIntP(offset.Y()))
		}
		position := imageLayout.PositionOf(img)
		rv += fmt.Sprintf(" - | \"$IM_COMPOSITE_BIN\" -compose atop -geometry \"+%d+%d\" - \"$OUTFILE\"  -colorspace sRGB -type truecolor \"$OUTFILE\"\n", toIntP(position.X()), toIntP(position.Y()))
		i++
	}
	imageLayout.Parameters().ProgressMonitor().ReportRenderingSuccess()
	return rv
}

func CollageRenderer_ImageMagickScript_Init() CollageRenderer_ImageMagickScript {
	return CollageRenderer_ImageMagickScript{}
}

// Produces output in the form of a Bash script that, when run, will call ImageMagick to produce
// the output collage.
type CollageRenderer_ImageMagickScript struct{}

func (icr CollageRenderer_ImageMagickScript) RegisterCustomParameters(parameters *Parameters) bool {
	return true
}

func (ict CollageRenderer_ImageMagickScript) ParseCustomParameters(parameters *Parameters) bool {
	return true
}

func (icr CollageRenderer_ImageMagickScript) CreateCollageImage(imageLayout ImageLayout) (oi OutputImage, err error) {
	oi, err = OutputImage_ImageMagickScript{createCollageImageMagickScript(imageLayout)}, nil
	return
}
