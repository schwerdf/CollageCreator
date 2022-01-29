package CollageCreator

import (
	"flag"
	"image"
	"log"
	"os"
)

const (
	Raster_PreloadImages string = "Raster_PreloadImages"
)

// An ImageInfo implementation that stores only the filename and dimension
// of an image and not all the pixel data.
type ImageInfo_placeholder struct {
	id       ImageIdentifier
	fileName string
	dims     Dims
}

func (iip ImageInfo_placeholder) ImageId() ImageIdentifier {
	return iip.id
}

func (iip ImageInfo_placeholder) FileName() string {
	return iip.fileName
}

func (iip ImageInfo_placeholder) DimensionsOf() Dims {
	return iip.dims
}

func (iip ImageInfo_placeholder) ImageData() interface{} {
	reader, err := os.Open(string(iip.fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	rv, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return rv
}

// An ImageInfo implementation that stores all of an image's pixel data.
type ImageInfo_impl struct {
	id       ImageIdentifier
	fileName string
	img      image.Image
}

func (iii ImageInfo_impl) ImageId() ImageIdentifier {
	return iii.id
}

func (iii ImageInfo_impl) FileName() string {
	return iii.fileName
}

func (iii ImageInfo_impl) DimensionsOf() Dims {
	return NewDims(float64(iii.img.Bounds().Max.X), float64(iii.img.Bounds().Max.Y))
}

func (iii ImageInfo_impl) ImageData() interface{} {
	return iii.img
}

// Loads an image for inclusion in an ImageLayout. If 'preload' is set, the
// entire image is loaded into memory; if not, only the header is read
// to obtain the dimensions.
func LoadImage(id ImageIdentifier, fileName string, preload bool) ImageInfo {
	reader, err := os.Open(string(fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	if preload {
		rv, _, err := image.Decode(reader)
		if err != nil {
			log.Fatal(err)
		}
		return ImageInfo_impl{id, fileName, rv}
	} else {
		rvC, _, err := image.DecodeConfig(reader)
		if err != nil {
			log.Fatal(err)
		}
		return ImageInfo_placeholder{id, fileName, NewDims(float64(rvC.Width), float64(rvC.Height))}
	}
}

type InputImageReader_Raster_CustomParameters struct {
	preload bool
}

func InputImageReader_Raster_Init() InputImageReader_Raster {
	return InputImageReader_Raster{new(InputImageReader_Raster_CustomParameters)}
}

// An InputImageReader that reads raster images in any format supported by the Go
// 'image' library.
type InputImageReader_Raster struct {
	p *InputImageReader_Raster_CustomParameters
}

func (iicio InputImageReader_Raster) RegisterCustomParameters(parameters *Parameters) bool {
	flag.BoolVar(&(iicio.p.preload), "1", false, "Preload all images, rather than loading dimensions at the start and data as necessary")
	return true
}

func (iicio InputImageReader_Raster) ParseCustomParameters(parameters *Parameters) bool {
	parameters.SetOther(Raster_PreloadImages, iicio.p.preload)
	return true
}

func (iicio InputImageReader_Raster) ReadInputImages(parameters *Parameters) (il ImageLayout, err error) {
	return readInputImages_Raster(parameters)
}

func readInputImages_Raster(parameters *Parameters) (il ImageLayout, err error) {
	files := parameters.InFiles()
	rv := ImageLayout_impl{data: new(imageLayout_data)}
	preload := parameters.OtherBool(Raster_PreloadImages)
	rv.data.size = NewDims(0, 0)
	rv.data.parameters = parameters
	rv.data.images = make([]ImageIdentifier, len(files))
	rv.data.imageInfo = make(map[ImageIdentifier]ImageInfo)
	rv.data.dimensions = make(map[ImageIdentifier]Dims)
	rv.data.cropping = make(map[ImageIdentifier]Geometry)
	rv.data.scaling = make(map[ImageIdentifier]Geometry)
	rv.data.positions = make(map[ImageIdentifier]Dims)
	for i, file := range files {
		rv.data.images[i] = ImageIdentifier(i)
		rv.data.imageInfo[rv.data.images[i]] = LoadImage(rv.data.images[i], file, preload)
		rv.data.dimensions[rv.data.images[i]] = rv.data.imageInfo[rv.data.images[i]].DimensionsOf()
	}
	il, err = rv, nil
	return
}
