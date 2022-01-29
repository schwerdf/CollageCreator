// This package provides an extensible framework for automatically generating image collages.
//
// (C) 2021 August Schwerdfeger
package CollageCreator

import (
	"log"
	"math"
)

// A superinterface for any object that adds custom parameters to the CollageCreator CLI.
type CollageCreatorComponent interface {
	// Adds information to 'parameters' about any custom parameters accepted by this object.
	RegisterCustomParameters(parameters *Parameters) bool
	// Parses, using Go's 'flag' library, any custom parameters in 'parameters' accepted by this object.
	ParseCustomParameters(parameters *Parameters) bool
}

// A superinterface for any object that reports the progress of the collage creator.
type ProgressMonitor interface {
	CollageCreatorComponent
	// Reports an unstructured message.
	ReportMessage(msg string)
	// Reports an unstructured runtime error.
	ReportRuntimeError(msg string, err error)
	// Reports that positioning has failed.
	ReportPositioningFailure()
	// Reports that positioning has succeeded.
	ReportPositioningSuccess()
	// Reports a set of dimensions.
	ReportDims(msg string, dims Dims)
	// Reports the progress of rendering, in terms of images rendered and still to render.
	ReportRenderingProgress(currentImage int, imageCount int)
	// Reports that rendering was a success.
	ReportRenderingSuccess()
	// Reports that rendering was a failure.
	ReportRenderingFailure()
	// Reports that output to a given file was a success.
	ReportOutputSuccess(fileName string)
	// Reports that output to a given file was a failure.
	ReportOutputFailure(fileName string)
	// Reports the progress of the random layout method:
	// the canvas size being tried; the number of canvas sizes tried so far and still to be tried;
	// the number of images positioned and still to be positioned; the number of attempts made
	// and still to make to position the current image.
	ReportRandomPositioningProgress(canvasSize Dims, currentCanvasTry int, maxCanvasTries int, currentPositionedCount int, totalImageCount int, currentImageTry int, maxImageTries int)
	// Reports the progress of the random layout method.
	ReportTileInOrderPositioningProgress(canvasSize Dims, badness tileInOrder_Badness)
	// Reports the progress of the whitespace-balancing post-process of the random layout method:
	// how many balancing steps have been taken; whether horizontal or vertical whitespace
	// is being balanced; how many iterations of balancing have been taken in the current
	// step; the maximum amount of whitespace imbalance still existing.
	ReportBalanceProgress(step int, dimIndex int, iteration int, maxImbalance int)
	// Reports a collision or overlap between two positioned images -- normally should not happen.
	ReportBalanceCollision(img ImageIdentifier, fileName string, oldPos Dims)
	// Reports that the whitespace-balancing post-process of the random layout method was a success.
	ReportBalancingSuccess()
	// Reports that the whitespace-balancing post-process of the random layout method was a failure.
	ReportBalancingFailure()
}

// A superinterface for any object that reads images into a layout.
type InputImageReader interface {
	CollageCreatorComponent
	// Read input images as specified in 'parameters';
	// return the ImageLayout object 'il' on success or error on failure.
	ReadInputImages(parameters *Parameters) (il ImageLayout, err error)
}

// A superinterface for any object that initializes crop and scale settings for each input image
// before positioning starts (e.g., by scaling each image to a uniform size).
type DimensionInitializer interface {
	CollageCreatorComponent
	// Initialize the dimensions of each image in 'imageLayout'. Returns error on failure,
	// or 'il' on success, which may be, but is not guaranteed to be, a separate object
	// from 'imageLayout'.
	InitializeDimensions(imageLayout ImageLayout) (il ImageLayout, err error)
}

// A superinterface for any object that positions images on the collage.
type PositionCalculator interface {
	CollageCreatorComponent
	// Find a place on a canvas for each image in 'imageLayout'. Returns error on failure,
	// or 'il' on success, which may be, but is not guaranteed to be, a separate object
	// from 'imageLayout'.
	CalculatePositions(imageLayout ImageLayout) (il ImageLayout, err error)
}

// A superinterface for any object that renders a finished image layout to a ready-to-write output image.
type CollageRenderer interface {
	CollageCreatorComponent
	// Takes 'imageLayout' after it has been through a 'PositionCalculator' and produces an object
	// ready to write to output file(s), or error on failure.
	CreateCollageImage(imageLayout ImageLayout) (oi OutputImage, err error)
}

// A superinterface for any object that holds an image ready to write to a file.
type OutputImage interface {
	WriteToFile(fileName string, parameters *Parameters)
}

// A map holding "custom" parameters specific to one component.
type CustomParameters map[string]interface{}

// Holds parameters to CollageCreator.
type Parameters struct {
	data *parameters_data
}

// Initializes a new instance of Parameters.
func Parameters_init() Parameters {
	params := Parameters{data: new(parameters_data)}
	params.data.inFiles = []string{}
	params.data.outFile = ""
	params.data.others = CustomParameters{}
	params.data.aspectRatio = EmptyGeometry()
	params.data.maxCanvasSize = NewDims(0, 0)
	params.data.minCanvasSize = NewDims(0, 0)
	params.data.padding = EmptyGeometry()
	return params
}

// Gets the list of pathnames representing input images.
func (p Parameters) InFiles() []string {
	return p.data.inFiles
}

// Sets the list of pathnames representing input images.
func (p *Parameters) SetInFiles(inFiles []string) {
	p.data.inFiles = inFiles
}

// Gets the pathname where the output is to be placed.
func (p Parameters) OutFile() string {
	return p.data.outFile
}

// Sets the pathname where the output is to be placed.
func (p *Parameters) SetOutFile(outFile string) {
	p.data.outFile = outFile
}

// Gets the minimum acceptable size of the final output image (0 signifying no limit).
func (p Parameters) MinCanvasSize() Dims {
	return p.data.minCanvasSize
}

// Sets the minimum acceptable size of the final output image (0 signifying no limit).
func (p *Parameters) SetMinCanvasSize(minCanvasSize Dims) {
	p.data.minCanvasSize = minCanvasSize
}

// Gets the maximum acceptable size of the final output image (0 signifying no limit).
func (p Parameters) MaxCanvasSize() Dims {
	return p.data.maxCanvasSize
}

// Gets the maximum acceptable size of the final output image (0 signifying no limit).
func (p *Parameters) SetMaxCanvasSize(maxCanvasSize Dims) {
	p.data.maxCanvasSize = maxCanvasSize
}

// Gets the target aspect ratio for the final output image, in Geometry form.
func (p Parameters) AspectRatioGeometry() Geometry {
	return p.data.aspectRatio
}

// Gets the target aspect ratio for the final output image, in the form of a
// float and a boolean indicating whether this number is a preference or a
// strict requirement.
func (p Parameters) AspectRatio() (ratio float64, strict bool) {
	geom := p.AspectRatioGeometry()
	if geom.height.N == 0 || math.IsNaN(geom.height.N) {
		ratio = 0.0
	} else {
		ratio = geom.width.N / geom.height.N
	}
	strict = geom.preserveAspectRatio
	return
}

// Gets the target aspect ratio for the final output image.
func (p *Parameters) SetAspectRatio(aspectRatio Geometry) {
	p.data.aspectRatio = aspectRatio
}

// Gets the Geometry governing padding around images during placement.
func (p Parameters) Padding() Geometry {
	return p.data.padding
}

// Sets the Geometry governing padding around images during placement.
func (p *Parameters) SetPadding(padding Geometry) {
	p.data.padding = padding
}

// Gets the ProgressMonitor used to report status.
func (p Parameters) ProgressMonitor() ProgressMonitor {
	return p.data.progressMonitor
}

// Sets the ProgressMonitor used to report status.
func (p *Parameters) SetProgressMonitor(progressMonitor ProgressMonitor) {
	p.data.progressMonitor = progressMonitor
}

// Gets the InputImageReader to be used for this run.
func (p Parameters) InputImageReader() InputImageReader {
	return p.data.inputImageReader
}

// Sets the InputImageReader to be used for this run.
func (p *Parameters) SetInputImageReader(inputImageReader InputImageReader) {
	p.data.inputImageReader = inputImageReader
}

// Gets the DimensionInitializer to be used for this run.
func (p Parameters) DimensionInitializer() DimensionInitializer {
	return p.data.dimensionInitializer
}

// Sets the DimensionInitializer to be used for this run.
func (p *Parameters) SetDimensionInitializer(dimensionInitializer DimensionInitializer) {
	p.data.dimensionInitializer = dimensionInitializer
}

// Gets the PositionCalculator to be used for this run.
func (p Parameters) PositionCalculator() PositionCalculator {
	return p.data.positionCalculator
}

// Sets the PositionCalculator to be used for this run.
func (p *Parameters) SetPositionCalculator(positionCalculator PositionCalculator) {
	p.data.positionCalculator = positionCalculator
}

// Gets the CollageRenderer to be used for this run.
func (p Parameters) CollageRenderer() CollageRenderer {
	return p.data.collageRenderer
}

// Sets the CollageRenderer to be used for this run.
func (p *Parameters) SetCollageRenderer(collageRenderer CollageRenderer) {
	p.data.collageRenderer = collageRenderer
}

// Gets a component-specific custom parameter by name, returning the value and a
// boolean that is false if the parameter has not been set.
func (p Parameters) Other(name string) (param interface{}, valid bool) {
	param, valid = p.data.others[name]
	return
}

// Gets by name a component-specific system parameter known to be a float,
// logging a fatal error if it is of any other type.
func (p Parameters) OtherFloat(name string) float64 {
	paramI, valid := p.data.others[name]
	if !valid {
		log.Fatal("Invalid parameter ", name)
	}
	switch paramI := paramI.(type) {
	case float64:
		return paramI
	default:
		log.Fatal("Mistyped parameter ", name)
	}
	return math.NaN()
}

// Gets by name a component-specific system parameter known to be an integer,
// logging a fatal error if it is of any other type.
func (p Parameters) OtherInt(name string) int {
	paramI, valid := p.data.others[name]
	if !valid {
		log.Fatal("Invalid parameter ", name)
	}
	switch paramI := paramI.(type) {
	case int:
		return paramI
	default:
		log.Fatal("Mistyped parameter ", name)
	}
	return -1
}

// Gets by name a component-specific system parameter known to be a boolean,
// logging a fatal error if it is of any other type.
func (p Parameters) OtherBool(name string) bool {
	paramI, valid := p.data.others[name]
	if !valid {
		log.Fatal("Invalid parameter ", name)
	}
	switch paramI := paramI.(type) {
	case bool:
		return paramI
	default:
		log.Fatal("Mistyped parameter ", name)
	}
	return false
}

// Gets by name a component-specific system parameter known to be a string,
// logging a fatal error if it is of any other type.
func (p Parameters) OtherString(name string) string {
	paramI, valid := p.data.others[name]
	if !valid {
		log.Fatal("Invalid parameter ", name)
	}
	switch paramI := paramI.(type) {
	case string:
		return paramI
	default:
		log.Fatal("Mistyped parameter ", name)
	}
	return ""
}

// Gets by name a component-specific system parameter known to be of type 'Dims',
// logging a fatal error if it is of any other type.
func (p Parameters) OtherDims(name string) Dims {
	paramI, valid := p.data.others[name]
	if !valid {
		log.Fatal("Invalid parameter ", name)
	}
	switch paramI := paramI.(type) {
	case Dims:
		return paramI
	default:
		log.Fatal("Mistyped parameter ", name)
	}
	return NewDims(0, 0)
}

// Gets by name a component-specific system parameter known to be of type 'Geometry',
// logging a fatal error if it is of any other type.
func (p Parameters) OtherGeometry(name string) Geometry {
	paramI, valid := p.data.others[name]
	if !valid {
		log.Fatal("Invalid parameter ", name)
	}
	switch paramI := paramI.(type) {
	case Geometry:
		return paramI
	default:
		log.Fatal("Mistyped parameter ", name)
	}
	return EmptyGeometry()
}

// Sets a component-specific parameter by name.
func (p Parameters) SetOther(name string, value interface{}) {
	p.data.others[name] = value
}

type parameters_data struct {
	inFiles              []string
	outFile              string
	minCanvasSize        Dims
	maxCanvasSize        Dims
	aspectRatio          Geometry
	padding              Geometry
	progressMonitor      ProgressMonitor
	inputImageReader     InputImageReader
	dimensionInitializer DimensionInitializer
	positionCalculator   PositionCalculator
	collageRenderer      CollageRenderer
	others               CustomParameters
}

// Runs the complete collage-creation process from reading input files to producing
// the output file; returns 0 if successful and nonzero if not.
func CreateCollage(parameters *Parameters) int {
	imageLayout, err := parameters.InputImageReader().ReadInputImages(parameters)
	if err != nil {
		parameters.ProgressMonitor().ReportRuntimeError("Error reading input images", err)
		return 1
	}
	imageLayout, err = parameters.DimensionInitializer().InitializeDimensions(imageLayout)
	if err != nil {
		parameters.ProgressMonitor().ReportRuntimeError("Error initializing dimensions", err)
		return 1
	}
	laidOut, err := parameters.PositionCalculator().CalculatePositions(imageLayout)
	if err != nil {
		parameters.ProgressMonitor().ReportRuntimeError("Error positioning images", err)
		return 1
	} else if laidOut.IsNil() {
		parameters.ProgressMonitor().ReportPositioningFailure()
		return 1
	}
	collageImage, err := parameters.CollageRenderer().CreateCollageImage(laidOut)
	if err != nil {
		parameters.ProgressMonitor().ReportRuntimeError("Error rendering collage", err)
		return 1
	}
	collageImage.WriteToFile(parameters.OutFile(), parameters)
	return 0
}
