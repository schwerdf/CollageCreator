package CollageCreator

import "math"

// An identifier assigned to an input image as it is read in.
type ImageIdentifier int

// Holds information about an input image.
type ImageInfo interface {
	ImageId() ImageIdentifier
	FileName() string
	DimensionsOf() Dims
	ImageData() interface{}
}

// Represents the layout of the collage: the positions and scaled/cropped dimensions of each constituent image.
type ImageLayout interface {
	// Most of the content of an ImageLayout is stored as a reference. This returns true if that is a nil reference.
	IsNil() bool
	// Creates a full copy of this ImageLayout.
	Duplicate() ImageLayout
	// The parameters assigned to this ImageLayout.
	Parameters() *Parameters
	// The total number of constituent images in this ImageLayout.
	TotalImageCount() int
	// The number of images in this ImageLayout that have been positioned in the collage.
	PositionedImageCount() int
	// Gets the size of the output image being generated.
	CanvasSize() Dims
	// Sets the size of the output image being generated.
	SetCanvasSize(d Dims)
	// Gets the list of images in this ImageLayout, a reference or a copy per the parameter.
	Images(copy bool) []ImageIdentifier
	// Gets the 'ImageInfo' for the given image.
	ImageInfoOf(img ImageIdentifier) ImageInfo
	// Clears all the position information set in this ImageLayout.
	ClearPositions() ImageLayout
	// Clears all the scaling information set in this ImageLayout.
	ClearDimensions() ImageLayout
	// Gets the position information for the given image.
	PositionOf(img ImageIdentifier) Dims
	// Sets the position information for the given image. Returns an ImageLayout including the new position
	// (which may or may not be the same object as the input ImageLayout) and, if the image as positioned
	// collided with another image, a pointer to that image.
	SetPosition(img ImageIdentifier, position Dims) (rv ImageLayout, collidedWith *ImageIdentifier)
	// Gets the dimensions of the given image as it will be placed on the canvas.
	DimensionsOf(img ImageIdentifier) Dims
	// Gets the cropping information for the given image.
	CroppingOf(img ImageIdentifier) Geometry
	// Sets the cropping information for the given image. Returns an ImageLayout including the new cropping information
	// (which may or may not be the same object as the input ImageLayout) and, if the image as positioned
	// collided with another image, a pointer to that image.
	SetCropping(img ImageIdentifier, geom Geometry) (rv ImageLayout, collidedWith *ImageIdentifier)
	// Gets the scaling information for the given image.
	ScalingOf(img ImageIdentifier) Geometry
	// Sets the scaling information for the given image. Returns an ImageLayout including the new scaling information
	// (which may or may not be the same object as the input ImageLayout) and, if the image as positioned
	// collided with another image, a pointer to that image.
	SetScaling(img ImageIdentifier, geom Geometry) (rv ImageLayout, collidedWith *ImageIdentifier)
	// Tests whether an image, positioned in the ImageLayout, collides with any others.
	TestCollision(newImage ImageIdentifier) *ImageIdentifier
}

type ImageLayout_impl struct {
	data *imageLayout_data
}

type imageLayout_data struct {
	size       Dims
	parameters *Parameters
	images     []ImageIdentifier
	imageInfo  map[ImageIdentifier]ImageInfo
	dimensions map[ImageIdentifier]Dims
	scaling    map[ImageIdentifier]Geometry
	cropping   map[ImageIdentifier]Geometry
	positions  map[ImageIdentifier]Dims
}

// Creates a "nil" image layout object (used to indicate an error).
func CreateNilImageLayout() ImageLayout {
	return ImageLayout_impl{data: nil}
}

// Most of the content of an ImageLayout is stored as a reference. This returns true if that is a nil reference.
func (iLay ImageLayout_impl) IsNil() bool {
	return iLay.data == nil
}

func (iLay ImageLayout_impl) Duplicate() ImageLayout {
	rv := ImageLayout_impl{data: new(imageLayout_data)}
	rv.data.size = iLay.data.size
	rv.data.parameters = iLay.data.parameters
	rv.data.images = iLay.data.images
	rv.data.imageInfo = iLay.data.imageInfo
	rv.data.dimensions = map[ImageIdentifier]Dims{}
	for img, info := range iLay.data.dimensions {
		rv.data.dimensions[img] = info
	}
	rv.data.scaling = map[ImageIdentifier]Geometry{}
	for img, info := range iLay.data.scaling {
		rv.data.scaling[img] = info
	}
	rv.data.cropping = map[ImageIdentifier]Geometry{}
	for img, info := range iLay.data.cropping {
		rv.data.cropping[img] = info
	}
	rv.data.positions = map[ImageIdentifier]Dims{}
	for img, info := range iLay.data.positions {
		rv.data.positions[img] = info
	}
	return rv
}

func (iLay ImageLayout_impl) Parameters() *Parameters {
	return iLay.data.parameters
}

func (iLay ImageLayout_impl) TotalImageCount() int {
	return len(iLay.data.imageInfo)
}

func (iLay ImageLayout_impl) PositionedImageCount() int {
	return len(iLay.data.positions)
}

func (iLay ImageLayout_impl) CanvasSize() Dims {
	return iLay.data.size
}

func (iLay ImageLayout_impl) SetCanvasSize(d Dims) {
	iLay.data.size = d
}

func (iLay ImageLayout_impl) Images(copy bool) []ImageIdentifier {
	if !copy {
		return iLay.data.images
	}
	rv := make([]ImageIdentifier, len(iLay.data.imageInfo))
	i := 0
	for _, img := range iLay.data.images {
		rv[i] = img
		i++
	}
	return rv
}

func (iLay ImageLayout_impl) ImageInfoOf(img ImageIdentifier) ImageInfo {
	return iLay.data.imageInfo[img]
}

func (iLay ImageLayout_impl) ClearDimensions() ImageLayout {
	iLay.data.cropping = map[ImageIdentifier]Geometry{}
	iLay.data.scaling = map[ImageIdentifier]Geometry{}
	for id := range iLay.data.imageInfo {
		iLay.data.cropping[id] = EmptyGeometry()
		iLay.data.scaling[id] = EmptyGeometry()
	}

	iLay.data.dimensions = map[ImageIdentifier]Dims{}
	for id, info := range iLay.data.imageInfo {
		iLay.data.dimensions[id] = info.DimensionsOf()
	}
	return iLay
}

func (iLay ImageLayout_impl) ClearPositions() ImageLayout {
	iLay.data.positions = map[ImageIdentifier]Dims{}
	return iLay
}

func (iLay ImageLayout_impl) PositionOf(img ImageIdentifier) Dims {
	return iLay.data.positions[img]
}

func (iLay ImageLayout_impl) SetPosition(img ImageIdentifier, position Dims) (rv ImageLayout, collidedWith *ImageIdentifier) {
	iLay.data.positions[img] = position
	collidedWith = iLay.TestCollision(img)
	rv = iLay
	return
}

func (iLay ImageLayout_impl) DimensionsOf(img ImageIdentifier) Dims {
	return iLay.data.dimensions[img]
}

func (iLay ImageLayout_impl) ScalingOf(img ImageIdentifier) Geometry {
	return iLay.data.scaling[img]
}

func (iLay ImageLayout_impl) SetScaling(img ImageIdentifier, geom Geometry) (rv ImageLayout, collidedWith *ImageIdentifier) {
	iLay.data.scaling[img] = geom
	iLay.data.dimensions[img] = ScaleAndCrop(iLay.data.imageInfo[img].DimensionsOf(), iLay.data.cropping[img], iLay.data.scaling[img])
	_, in := iLay.data.positions[img]
	if in {
		collidedWith = nil
	} else {
		collidedWith = iLay.TestCollision(img)
	}
	rv = iLay
	return
}

func (iLay ImageLayout_impl) CroppingOf(img ImageIdentifier) Geometry {
	return iLay.data.cropping[img]
}

func (iLay ImageLayout_impl) SetCropping(img ImageIdentifier, geom Geometry) (rv ImageLayout, collidedWith *ImageIdentifier) {
	iLay.data.cropping[img] = geom
	iLay.data.dimensions[img] = ScaleAndCrop(iLay.data.imageInfo[img].DimensionsOf(), iLay.data.cropping[img], iLay.data.scaling[img])
	_, in := iLay.data.positions[img]
	if in {
		collidedWith = nil
	} else {
		collidedWith = iLay.TestCollision(img)
	}
	rv = iLay
	return
}

// Calculates the padding to be maintained around the given image.
func Padding(iLay ImageLayout, img ImageIdentifier) Dims {
	return iLay.Parameters().Padding().Scale(iLay.DimensionsOf(img))
}

// Determines whether the padding to be maintained around the given
// image is in relative or absolute units.
func PaddingIsRelative(iLay ImageLayout, img ImageIdentifier) bool {
	padding := iLay.Parameters().Padding()
	return (padding.HasWidth() && padding.width.U == Percent) ||
		(padding.HasHeight() && padding.height.U == Percent)
}

func overlap_inner(pos1, pad1, pos2, pad2, dim1, dim2 Dims) Dims {
	x1 := pos1.X() - pad1.X()
	x1p := x1 + dim1.X() + pad1.X()
	x2 := pos2.X() - pad2.X()
	x2p := x2 + dim2.X() + pad2.X()
	y1 := pos1.Y() - pad1.Y()
	y1p := y1 + dim1.Y() + pad1.Y()
	y2 := pos2.Y() - pad2.Y()
	y2p := y2 + dim2.Y() + pad2.Y()

	return NewDims(math.Min(x1p-x2, x2p-x1), math.Min(y1p-y2, y2p-y1))
}

func Overlap(iLay ImageLayout, img1, img2 ImageIdentifier) Dims {
	pos1 := iLay.PositionOf(img1)
	pad1 := Padding(iLay, img1)
	pos2 := iLay.PositionOf(img2)
	pad2 := Padding(iLay, img2)
	dim1 := iLay.DimensionsOf(img1)
	dim2 := iLay.DimensionsOf(img2)

	return overlap_inner(pos1, pad1, pos2, pad2, dim1, dim2)
}

func (iLay ImageLayout_impl) TestCollision(newImage ImageIdentifier) *ImageIdentifier {
	for img := range iLay.data.positions {
		if img != newImage && !isWithin(Overlap(iLay, img, newImage), NewDims(0, 0)) {
			return &img
		}
	}
	return nil
}
