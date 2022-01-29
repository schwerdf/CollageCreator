package CollageCreator

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	Random_SeedNumber               string = "Random_SeedNumber"
	Random_MaxLayoutTries           string = "Random_MaxLayoutTries"
	Random_MaxTriesPerImage         string = "Random_MaxTriesPerImage"
	Random_SizeToleranceFactor      string = "Random_SizeToleranceFactor"
	Balancer_MaxBalanceIterations   string = "Balancer_MaxBalanceIterations"
	Balancer_BalanceToleranceFactor string = "Balancer_BalanceToleranceFactor"
)

func oneImageTry(imageLayout ImageLayout, parameters *Parameters, img *ImageIdentifier, maxDims Dims) (imlrv ImageLayout, positioned *ImageIdentifier) {
	maxX, maxY := maxDims.X(), maxDims.Y()
	posX := rand.Intn(int(math.Max(1, maxX-imageLayout.DimensionsOf(*img).X()-2*Padding(imageLayout, *img).X()) + Padding(imageLayout, *img).X()))
	posY := rand.Intn(int(math.Max(1, maxY-imageLayout.DimensionsOf(*img).Y()-2*Padding(imageLayout, *img).Y()) + Padding(imageLayout, *img).Y()))
	positionToTry := NewDims(float64(posX), float64(posY))
	imlrv, positioned = imageLayout.SetPosition(*img, positionToTry)
	return
}

func oneCanvasTry(imageLayout ImageLayout, parameters *Parameters, imagesInOrder *[]ImageIdentifier, try *int, maxDims Dims) bool {
	for _, img := range *imagesInOrder {
		var positioned *ImageIdentifier = nil
		i := 1
		positionedCount := imageLayout.PositionedImageCount() + 1
		for i <= parameters.OtherInt(Random_MaxTriesPerImage) {
			parameters.ProgressMonitor().ReportRandomPositioningProgress(maxDims, (*try)+1, parameters.OtherInt(Random_MaxLayoutTries), positionedCount, imageLayout.TotalImageCount(), i, parameters.OtherInt(Random_MaxTriesPerImage))
			imageLayout, positioned = oneImageTry(imageLayout, parameters, &img, maxDims)
			if positioned == nil {
				break
			}
			i++
		}
		if i > parameters.OtherInt(Random_MaxTriesPerImage) {
			return false
		}
	}
	return true
}

func calculatePositions_Random_inner(images ImageLayout, maxDims Dims) ImageLayout {
	imageLayout := images.Duplicate()
	parameters := imageLayout.Parameters()
	imageLayout.SetCanvasSize(maxDims)
	tries := 0
	imagesInOrder := imageLayout.Images(true)
	IISBy(func(lhs, rhs *ImageIdentifier) bool {
		return (imageLayout.DimensionsOf(*rhs).X() * imageLayout.DimensionsOf(*rhs).Y()) < (imageLayout.DimensionsOf(*lhs).X() * imageLayout.DimensionsOf(*lhs).Y())
	}).Sort(imagesInOrder)
	for tries < parameters.OtherInt(Random_MaxLayoutTries) && imageLayout.PositionedImageCount() < imageLayout.TotalImageCount() {
		success := oneCanvasTry(imageLayout, parameters, &imagesInOrder, &tries, maxDims)
		if !success {
			imageLayout = imageLayout.ClearPositions()
			tries++
		}
	}
	if tries == parameters.OtherInt(Random_MaxLayoutTries) {
		parameters.ProgressMonitor().ReportPositioningFailure()
		return CreateNilImageLayout()
	} else {
		parameters.ProgressMonitor().ReportPositioningSuccess()
		return imageLayout
	}
}

func calculatePositions_Random(imageLayout ImageLayout) ImageLayout {
	parameters := imageLayout.Parameters()
	seed, valid := parameters.Other(Random_SeedNumber)
	if !valid {
		seed = time.Now().UnixNano()
	}
	rand.Seed(seed.(int64))
	parameters.ProgressMonitor().ReportMessage(fmt.Sprintf("Seed for random number generator: %d", seed))
	minDim, maxDim := getDimensionRange(imageLayout)
	targetWidth := math.Max(maxDim, math.Sqrt(float64(imageLayout.TotalImageCount()))*minDim)
	maxWidthP := 2 * math.Sqrt(float64(imageLayout.TotalImageCount())) * maxDim
	aspectRatio := autoAspectRatio(imageLayout)
	var minX, maxX float64
	if parameters.MinCanvasSize() != (NewDims(0, 0)) {
		minX = math.Max(parameters.MinCanvasSize().X(), parameters.MinCanvasSize().Y()/aspectRatio)
	} else {
		minX = targetWidth
	}
	if parameters.MaxCanvasSize() != (NewDims(0, 0)) {
		if parameters.MaxCanvasSize().X() == 0 {
			maxX = parameters.MaxCanvasSize().Y() / aspectRatio
		} else {
			maxX = math.Min(parameters.MaxCanvasSize().X(), parameters.MaxCanvasSize().Y()/aspectRatio)
		}
	} else {
		maxX = maxWidthP
	}
	var bestSoFar ImageLayout = CreateNilImageLayout()
	midpoint := 0.0
	bestSoFar = calculatePositions_Random_inner(imageLayout, NewDims(minX, minX/aspectRatio))
	if bestSoFar.IsNil() {
		bestSoFar = calculatePositions_Random_inner(imageLayout, NewDims(maxX, maxX/aspectRatio))
		if bestSoFar.IsNil() {
			return bestSoFar
		}
		for {
			midpoint = math.Round(minX + (maxX-minX)/2)
			if midpoint == minX || midpoint == maxX || float64(maxX-minX)/float64(targetWidth) < parameters.OtherFloat(Random_SizeToleranceFactor) {
				break
			}
			layout := calculatePositions_Random_inner(imageLayout, NewDims(midpoint, midpoint/aspectRatio))
			if layout.IsNil() {
				minX = midpoint
			} else {
				bestSoFar = layout
				bestSoFar.SetCanvasSize(NewDims(midpoint, midpoint/aspectRatio))
				maxX = midpoint
			}
		}
	}
	if !bestSoFar.IsNil() {
		parameters.ProgressMonitor().ReportDims("Final bounding box", bestSoFar.CanvasSize())
		if parameters.OtherInt(Balancer_MaxBalanceIterations) > 0 {
			bestSoFar = Balance(bestSoFar)
		}
	}
	return bestSoFar
}

type positionCalculator_Random_Parameters struct {
	randomSeed             int64
	canvasTries            int
	imageTries             int
	sizeToleranceFactor    float64
	maxBalanceIterations   int
	balanceToleranceFactor float64
}

// A PositionCalculator that places images randomly on the canvas.
type PositionCalculator_Random struct {
	p *positionCalculator_Random_Parameters
}

func PositionCalculator_Random_Init() PositionCalculator_Random {
	return PositionCalculator_Random{new(positionCalculator_Random_Parameters)}
}

func (pcr PositionCalculator_Random) RegisterCustomParameters(parameters *Parameters) bool {
	flag.Int64Var(&(pcr.p.randomSeed), "random-seed", -1, "(Random placement algorithm) Use this seed for generating random numbers (-1 to use a time-based seed)")
	flag.IntVar(&(pcr.p.canvasTries), "canvas-tries", 25, "(Random placement algorithm) Try a specific canvas size this many times")
	flag.IntVar(&(pcr.p.imageTries), "image-tries", 100, "(Random placement algorithm) Try to place a specific image this many times")
	flag.Float64Var(&(pcr.p.sizeToleranceFactor), "size-tolerance", 0.1, "(Random placement algorithm) Stop when the final canvas has gotten within this factor of the target size")
	flag.IntVar(&(pcr.p.maxBalanceIterations), "balance", 4, "(Random placement algorithm) Number of iterations to balance spacing (0 to skip balancing)")
	flag.Float64Var(&(pcr.p.balanceToleranceFactor), "balance-tolerance", 0.01, "(Random placement algorithm) Tolerance factor for imbalances in spacing")
	return true
}

func (pcr PositionCalculator_Random) ParseCustomParameters(parameters *Parameters) bool {
	if pcr.p.randomSeed > -1 {
		parameters.SetOther(Random_SeedNumber, pcr.p.randomSeed)
	}
	parameters.SetOther(Random_MaxLayoutTries, pcr.p.canvasTries)
	parameters.SetOther(Random_MaxTriesPerImage, pcr.p.imageTries)
	parameters.SetOther(Random_SizeToleranceFactor, pcr.p.sizeToleranceFactor)
	parameters.SetOther(Balancer_MaxBalanceIterations, pcr.p.maxBalanceIterations)
	parameters.SetOther(Balancer_BalanceToleranceFactor, pcr.p.balanceToleranceFactor)
	return true
}

func (pcr PositionCalculator_Random) CalculatePositions(imageLayout ImageLayout) (il ImageLayout, err error) {
	il, err = calculatePositions_Random(imageLayout), nil
	return
}
