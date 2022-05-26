package CollageCreator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Produces output in the form of an SVG file that links to all the input images.
type OutputImage_SVG struct {
	contents string
}

func (ois OutputImage_SVG) WriteToFile(fileName string, parameters *Parameters) {
	err := os.WriteFile(fileName, []byte(ois.contents), 0666)
	if err != nil {
		log.Fatal(err)
	}
	parameters.ProgressMonitor().ReportOutputSuccess(fileName)
}

func createCollageSVG(imageLayout ImageLayout) string {
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

	rv += fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n	<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\"\n		 width=\"%f\" height=\"%f\" viewBox=\"0 -%f %f %f\">\n<defs>\n</defs>\n  <rect x=\"0\" y=\"-%f\" width=\"%f\" height=\"%f\"/>\n", xSize, ySize, ySize, xSize, ySize, ySize, xSize, ySize)

	i := 1
	clipPaths := ""
	imageTags := ""
	for _, img := range imageLayout.Images(false) {
		imageLayout.Parameters().ProgressMonitor().ReportRenderingProgress(i, imageLayout.PositionedImageCount())
		position := imageLayout.PositionOf(img)
		cropping := imageLayout.CroppingOf(img)
		scaling := imageLayout.ScalingOf(img)
		imageInfo := imageLayout.ImageInfoOf(img)
		dimensions := imageInfo.DimensionsOf()
		if scaling.HasSize() {
			dimensions = scaling.Scale(dimensions)
		}
		imagePath, errI := filepath.Abs(imageLayout.ImageInfoOf(img).FileName())
		if errI != nil {
			imagePath = imageLayout.ImageInfoOf(img).FileName()
		}
		if cropping.HasOffset() {
			croppedDimensions := cropping.Crop(dimensions)
			offset := cropping.Offset(dimensions)
			clipPaths += fmt.Sprintf("  <clipPath id=\"clip%d\"><rect x=\"%f\" y=\"%f\" width=\"%f\" height=\"%f\"/></clipPath>\n", i, position.X(), -(ySize - position.Y()), croppedDimensions.X(), croppedDimensions.Y())
			imageTags += fmt.Sprintf("  <image x=\"%f\" y=\"%f\" width=\"%f\" height=\"%f\"  clip-path=\"url(#clip%d)\" xlink:href=\"file:///%s\"/>\n", position.X()-offset.X(), -(ySize-position.Y())+offset.Y(), dimensions.X(), dimensions.Y(), i, imagePath)
		} else {
			imageTags += fmt.Sprintf("  <image x=\"%f\" y=\"%f\" width=\"%f\" height=\"%f\" xlink:href=\"file:///%s\"/>\n", position.X(), -(ySize - position.Y()), dimensions.X(), dimensions.Y(), imagePath)
		}

		i++
	}
	rv += clipPaths
	rv += imageTags
	rv += "</svg>\n"
	imageLayout.Parameters().ProgressMonitor().ReportRenderingSuccess()
	return rv
}

func CollageRenderer_SVG_Init() CollageRenderer_SVG {
	return CollageRenderer_SVG{}
}

// Produces output in the form of an SVG file that links to all the input images.
type CollageRenderer_SVG struct{}

func (icr CollageRenderer_SVG) RegisterCustomParameters(parameters *Parameters) bool {
	return true
}

func (ict CollageRenderer_SVG) ParseCustomParameters(parameters *Parameters) bool {
	return true
}

func (icr CollageRenderer_SVG) CreateCollageImage(imageLayout ImageLayout) (oi OutputImage, err error) {
	oi, err = OutputImage_SVG{createCollageSVG(imageLayout)}, nil
	return
}
