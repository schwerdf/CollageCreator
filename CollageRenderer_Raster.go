package CollageCreator

import (
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"golang.org/x/image/tiff"
)

// Produces output in the form of a JPEG, PNG, or TIFF file, using Jan Schlicht's "resize" package
// to handle scaling.
type OutputImage_image struct {
	img image.Image
}

func (oii OutputImage_image) WriteToFile(fileName string, parameters *Parameters) {

	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".jpg", ".jpeg":
		{
			fp, err := os.Create(fileName)
			if err != nil {
				log.Fatal(err)
			}
			defer fp.Close()
			opt := jpeg.Options{Quality: 95}
			err = jpeg.Encode(fp, oii.img, &opt)
			if err != nil {
				log.Fatal(err)
			}
			parameters.ProgressMonitor().ReportOutputSuccess(fileName)
		}
	case ".png":
		{
			fp, err := os.Create(fileName)
			if err != nil {
				log.Fatal(err)
			}
			defer fp.Close()
			err = png.Encode(fp, oii.img)
			if err != nil {
				log.Fatal(err)
			}
			parameters.ProgressMonitor().ReportOutputSuccess(fileName)
		}
	case ".tif", ".tiff":
		{
			fp, err := os.Create(fileName)
			if err != nil {
				log.Fatal(err)
			}
			defer fp.Close()
			opt := tiff.Options{
				Compression: tiff.Deflate,
				Predictor:   true,
			}
			err = tiff.Encode(fp, oii.img, &opt)
			if err != nil {
				log.Fatal(err)
			}
			parameters.ProgressMonitor().ReportOutputSuccess(fileName)
		}
	default:
		{
			log.Fatal("Unknown file extension: " + filepath.Ext(fileName))
		}
	}
}

func createCollageImage(imageLayout ImageLayout) image.Image {
	xAdd := 0.0
	yAdd := 0.0
	xSize := 0.0
	ySize := 0.0
	if imageLayout.CanvasSize() != (NewDims(0, 0)) {
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
	collageImage := image.NewNRGBA(image.Rect(0, 0, toInt(xSize), toInt(ySize)))

	i := 1
	for _, img := range imageLayout.Images(false) {
		imageLayout.Parameters().ProgressMonitor().ReportRenderingProgress(i, imageLayout.PositionedImageCount())
		info := imageLayout.ImageInfoOf(img)
		imgData := info.ImageData().(image.Image)
		dimensions := info.DimensionsOf()
		scaling := imageLayout.ScalingOf(img)
		if scaling.HasSize() {
			dimensions = scaling.Scale(dimensions)
			imgData = resize.Resize(uint(toIntP(dimensions.X())), uint(toIntP(dimensions.Y())), imgData, resize.Lanczos3)
		}
		cropping := imageLayout.CroppingOf(img)
		offset := Dims{0, 0}
		if cropping.HasOffset() {
			offset = cropping.Offset(dimensions)
			dimensions = cropping.Crop(dimensions)
		}
		position := imageLayout.PositionOf(img)
		positionRect := image.Rect(toInt(position.X()), toInt(position.Y()), toIntP(position.X()+dimensions.X()), toIntP(position.Y()+dimensions.Y()))
		draw.Draw(collageImage, positionRect, imgData, image.Point{toInt(offset.X()), toInt(offset.Y())}, draw.Src)
		i++
	}
	imageLayout.Parameters().ProgressMonitor().ReportRenderingSuccess()
	return collageImage
}

func CollageRenderer_Raster_Init() CollageRenderer_Raster {
	return CollageRenderer_Raster{}
}

// Produces output in the form of a JPEG, PNG, or TIFF file, using Jan Schlicht's "resize" package
// to handle scaling.
type CollageRenderer_Raster struct{}

func (icr CollageRenderer_Raster) RegisterCustomParameters(parameters *Parameters) bool {
	return true
}

func (ict CollageRenderer_Raster) ParseCustomParameters(parameters *Parameters) bool {
	return true
}

func (icr CollageRenderer_Raster) CreateCollageImage(imageLayout ImageLayout) (oi OutputImage, err error) {
	oi, err = OutputImage_image{createCollageImage(imageLayout)}, nil
	return
}
