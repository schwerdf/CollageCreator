package CollageCreator

import (
	"fmt"
	"math"
	"sort"
)

func isWithin(x, y Dims) bool {
	x1, x2 := x.X(), x.Y()
	y1, y2 := y.X(), y.Y()

	return x1 <= y1 || x2 <= y2
}

// A comparator for ImageIdentifiers, to be used in sorting images.
type IISBy func(lhs, rhs *ImageIdentifier) bool

type IISorter struct {
	images []ImageIdentifier
	by     IISBy
}

func (by IISBy) Sort(ids []ImageIdentifier) {
	is := &IISorter{images: ids, by: by}
	sort.Sort(is)
}

func (iis *IISorter) Len() int {
	return len(iis.images)
}

func (iis *IISorter) Swap(i, j int) {
	iis.images[i], iis.images[j] = iis.images[j], iis.images[i]
}

func (iis *IISorter) Less(i, j int) bool {
	return iis.by(&iis.images[i], &iis.images[j])
}

func getDimensionRange(imageLayout ImageLayout) (minDim, maxDim float64) {
	minDim, maxDim = math.Inf(1), math.Inf(-1)
	for _, img := range imageLayout.Images(false) {
		if math.Min(imageLayout.DimensionsOf(img).X(), imageLayout.DimensionsOf(img).Y()) < minDim {
			minDim = math.Min(imageLayout.DimensionsOf(img).X(), imageLayout.DimensionsOf(img).Y())
		}
		if math.Max(imageLayout.DimensionsOf(img).X(), imageLayout.DimensionsOf(img).Y()) > maxDim {
			maxDim = math.Max(imageLayout.DimensionsOf(img).X(), imageLayout.DimensionsOf(img).Y())
		}
	}
	return
}

func getDimensionsRange(imageLayout ImageLayout, withPadding bool) (minima, maxima, sums Dims) {
	minWidth, maxWidth, sumWidth, minHeight, maxHeight, sumHeight := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	for _, img := range imageLayout.Images(false) {
		imgDims := imageLayout.DimensionsOf(img)
		if withPadding {
			padding := Padding(imageLayout, img)
			imgDims.SetX(imgDims.X() + 2*padding.X())
			imgDims.SetY(imgDims.Y() + 2*padding.Y())
		}
		sumWidth += imgDims.X()
		sumHeight += imgDims.Y()
		if minWidth == 0 || imgDims.X() < minWidth {
			minWidth = imgDims.X()
		}
		if maxWidth == 0 || imgDims.X() > maxWidth {
			maxWidth = imgDims.X()
		}
		if minHeight == 0 || imgDims.Y() < minHeight {
			minHeight = imgDims.Y()
		}
		if maxHeight == 0 || imgDims.Y() > maxHeight {
			maxHeight = imgDims.Y()
		}
	}
	minima, maxima, sums = NewDims(minWidth, minHeight), NewDims(maxWidth, maxHeight), NewDims(sumWidth, sumHeight)
	return
}

func autoAspectRatio(imageLayout ImageLayout) float64 {
	aspectRatio, _ := imageLayout.Parameters().AspectRatio()
	if aspectRatio == 0.0 {
		arSum := 0.0
		for _, img := range imageLayout.Images(false) {
			imgDims := imageLayout.DimensionsOf(img)
			arSum += float64(imgDims.X()) / float64(imgDims.Y())
		}
		aspectRatio = arSum / float64(imageLayout.TotalImageCount())
	}
	return aspectRatio
}

func ProgressMonitor_Init() ProgressMonitor_impl {
	return ProgressMonitor_impl{}
}

type ProgressMonitor_impl struct{}

func (pmi ProgressMonitor_impl) RegisterCustomParameters(parameters *Parameters) bool {
	return true
}

func (pmi ProgressMonitor_impl) ParseCustomParameters(parameters *Parameters) bool {
	return true
}

func (pmi ProgressMonitor_impl) ReportMessage(msg string) {
	fmt.Println(msg)
}
func (pmi ProgressMonitor_impl) ReportPositioningFailure() {
	fmt.Println("Could not position")
}
func (pmi ProgressMonitor_impl) ReportPositioningSuccess() {
	fmt.Println("Positioned")
}
func (pmi ProgressMonitor_impl) ReportDims(msg string, dims Dims) {
	fmt.Printf("%s: (%f,%f)\n", msg, dims.X(), dims.Y())
}
func (pmi ProgressMonitor_impl) ReportRandomPositioningProgress(canvasSize Dims, currentCanvasTry int, maxCanvasTries int, currentPositionedCount int, totalImageCount int, currentImageTry int, maxImageTries int) {
	fmt.Printf("\r(%f,%f) - try %d/%d - Image %d/%d - try %d/%d       ", canvasSize.X(), canvasSize.Y(), currentCanvasTry, maxCanvasTries, currentPositionedCount, totalImageCount, currentImageTry, maxImageTries)
}
func (pmi ProgressMonitor_impl) ReportTileInOrderPositioningProgress(canvasSize Dims, badness tileInOrder_Badness) {
	fmt.Printf("\r(%.0f,%.0f) - U %.0f, S %.1f, A %f       ", canvasSize.X(), canvasSize.Y(), badness.emptySpace, badness.scaledownSum, badness.aspectRatioSkew)
}
func (pmi ProgressMonitor_impl) ReportBalanceProgress(step int, dimIndex int, iteration int, maxImbalance int) {
	fmt.Printf("\rBalance step %d(%d) iteration %d - max imbalance %d   ", step, dimIndex, iteration, maxImbalance)
}
func (pmi ProgressMonitor_impl) ReportBalanceCollision(img ImageIdentifier, fileName string, oldPos Dims) {
	fmt.Printf("Collision on alignment of image #%d (%s) from (%f,%f)\n", img, fileName, oldPos.X(), oldPos.Y())
}
func (pmi ProgressMonitor_impl) ReportBalancingSuccess() {
	fmt.Println("Balanced")
}
func (pmi ProgressMonitor_impl) ReportBalancingFailure() {
	fmt.Println("Failed to balance")
}
func (pmi ProgressMonitor_impl) ReportRenderingProgress(currentImage int, imageCount int) {
	fmt.Printf("\rRendering image %d / %d     ", currentImage, imageCount)
}
func (pmi ProgressMonitor_impl) ReportRenderingSuccess() {
	fmt.Println("Fully rendered")
}
func (pmi ProgressMonitor_impl) ReportRenderingFailure() {
	fmt.Println("Rendering failed")
}
func (pmi ProgressMonitor_impl) ReportOutputSuccess(fileName string) {
	fmt.Printf("Output to %s\n", fileName)
}
func (pmi ProgressMonitor_impl) ReportOutputFailure(fileName string) {
	fmt.Printf("Failure on output to %s\n", fileName)
}

func (pmi ProgressMonitor_impl) ReportRuntimeError(msg string, err error) {
	fmt.Printf("%s: %s\n", msg, err.Error())
}
