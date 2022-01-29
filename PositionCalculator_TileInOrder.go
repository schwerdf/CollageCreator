package CollageCreator

import (
	"errors"
	"flag"
	"fmt"
	"math"
)

const (
	TileInOrder_ExactOrder string = "TileInOrder_ExactOrder"
	TileInOrder_Columns    string = "TileInOrder_Columns"
)

type badnessComparator interface {
	hasAspectRatio() bool
	prioritizeAspectRatio() bool
	badnessLT(b1, b2 tileInOrder_Badness) bool
}

type badnessComparator_LT_impl struct {
	hasAspectRatioF        bool
	prioritizeAspectRatioF bool
}

func (bci badnessComparator_LT_impl) hasAspectRatio() bool {
	return bci.hasAspectRatioF
}

func (bci badnessComparator_LT_impl) prioritizeAspectRatio() bool {
	return bci.prioritizeAspectRatioF
}

func (bci badnessComparator_LT_impl) badnessLT(b1, b2 tileInOrder_Badness) bool {
	if bci.prioritizeAspectRatio() {
		if b1.aspectRatioSkew != b2.aspectRatioSkew {
			return math.Abs(b1.aspectRatioSkew) < math.Abs(b2.aspectRatioSkew)
		} else if b1.emptySpace != b2.emptySpace {
			return b1.emptySpace < b2.emptySpace
		} else {
			return b1.scaledownSum < b2.scaledownSum
		}
	} else {
		if b1.emptySpace != b2.emptySpace {
			return b1.emptySpace < b2.emptySpace
		} else if b1.aspectRatioSkew != b2.aspectRatioSkew {
			return math.Abs(b1.aspectRatioSkew) < math.Abs(b2.aspectRatioSkew)
		} else {
			return b1.scaledownSum < b2.scaledownSum
		}
	}
}

type tileInOrder_Badness struct {
	emptySpace      float64
	scaledownSum    float64
	aspectRatioSkew float64
}

type tileLine struct {
	images   []ImageIdentifier
	fixedDim float64
	startsAt float64
}

type tileLayout struct {
	lines    []tileLine
	fixedDim float64
}

func (tl *tileLayout) append(line tileLine) {
	tl.lines = append(tl.lines, line)
	tl.fixedDim += line.fixedDim
}

// type tileLayoutSorter struct {
// 	tl *tileLayout
// }

// func (tls *tileLayoutSorter) Less(i, j int) bool {
// 	return (tls.tl.lines[i].fixedDim < tls.tl.lines[j].fixedDim)
// }

// func (tls *tileLayoutSorter) Swap(i, j int) {
// 	tls.tl.lines[i], tls.tl.lines[j] = tls.tl.lines[j], tls.tl.lines[i]
// }

// func (tls *tileLayoutSorter) Len() int {
// 	return len(tls.tl.lines)
// }

// func rebalance_line(imageLayout ImageLayout, line tileLine) {

// }

// func rebalance(tl *tileLayout) {
// 	tls := &tileLayoutSorter{tl: tl}
// 	sort.Stable(tls)
// 	var newLines []tileLine = make([]tileLine, len(tl.lines))
// 	tl.lines = newLines
// }

func finalizeTiling_line(imageLayout ImageLayout, line tileLine, nextLineDim float64, badness *tileInOrder_Badness, fixedDim, varDim int) (il ImageLayout, newLinePosition float64, err error) {
	currentLayout := imageLayout
	nextImageDim := line.startsAt
	for _, img := range line.images {
		imgDims := currentLayout.DimensionsOf(img)
		newImgDims := NewDims(0, 0)
		newImgDims.SetDim(fixedDim, imgDims.Dim(fixedDim)*line.fixedDim/imgDims.Dim(varDim))
		newImgDims.SetDim(varDim, line.fixedDim)
		newScaling := MustParseGeometry(fmt.Sprintf("%fx%f!", newImgDims.X(), newImgDims.Y()))
		currentLayout, _ = currentLayout.SetScaling(img, newScaling)
		imgPadding := Padding(currentLayout, img)
		pos := NewDims(0, 0)
		pos.SetDim(fixedDim, nextImageDim+imgPadding.Dim(fixedDim))
		pos.SetDim(varDim, nextLineDim+imgPadding.Dim(varDim))
		currentLayout, _ = currentLayout.SetPosition(img, pos)
		nextImageDim += newImgDims.Dim(fixedDim) + 2*imgPadding.Dim(fixedDim)
	}
	il = currentLayout
	err = nil
	newLinePosition = nextLineDim + line.fixedDim + 2*Padding(currentLayout, line.images[0]).Dim(varDim)
	return
}

func finalizeTiling(imageLayout ImageLayout, tl tileLayout, badness *tileInOrder_Badness, fixedDim, varDim int) (il ImageLayout, newLinePosition float64, err error) {
	currentLayout := imageLayout
	nextLineDim := 0.0
	for _, line := range tl.lines {
		currentLayout, nextLineDim, _ = finalizeTiling_line(currentLayout, line, nextLineDim, badness, fixedDim, varDim)
	}
	il = currentLayout
	newLinePosition = nextLineDim
	err = nil
	return
}

func runOneTiling(imageLayout ImageLayout, imagesInOrder []ImageIdentifier, dim Dims) (il ImageLayout, badness tileInOrder_Badness, err error) {
	parameters := imageLayout.Parameters()
	aspectRatio, _ := parameters.AspectRatio()

	var fixedDim, varDim int
	if dim.Y() > 0 {
		fixedDim = 1
		varDim = 0
	} else {
		fixedDim = 0
		varDim = 1
	}

	maxFixedDim := dim.Dim(fixedDim)

	currentLayout := imageLayout
	currentLineStartIndex := 0
	currentMinVarDim := 0.0
	imagesAspect := 0.0
	relativePaddingAspect := 0.0
	absolutePadding := 0.0
	currentLinePosition := 0.0
	badness = tileInOrder_Badness{emptySpace: -1, scaledownSum: 0.0, aspectRatioSkew: 0.0}

	tl := tileLayout{lines: make([]tileLine, 0, len(imagesInOrder)), fixedDim: 0}
	for i, img := range imagesInOrder {
		imgDims := currentLayout.DimensionsOf(img)
		imgPadding := Padding(currentLayout, img)
		if currentMinVarDim == 0 || imgDims.Dim(varDim) < currentMinVarDim {
			currentMinVarDim = imgDims.Dim(varDim)
		}
		imagesAspect += imgDims.Dim(fixedDim) / imgDims.Dim(varDim)
		if PaddingIsRelative(currentLayout, img) {
			relativePaddingAspect += 2 * imgPadding.Dim(fixedDim) / imgDims.Dim(varDim)
		} else {
			absolutePadding += 2 * imgPadding.Dim(fixedDim)
		}
		currentLineWidth := (imagesAspect+relativePaddingAspect)*currentMinVarDim + absolutePadding
		if currentLineWidth >= maxFixedDim {
			currentMinVarDim = (maxFixedDim - absolutePadding) / (imagesAspect + relativePaddingAspect)
			line := tileLine{images: imagesInOrder[currentLineStartIndex : i+1], fixedDim: currentMinVarDim, startsAt: 0}
			tl.append(line)
			for _, jmg := range line.images {
				badness.scaledownSum += (currentLayout.DimensionsOf(jmg).Dim(varDim) / line.fixedDim) - 1.0
			}
			currentMinVarDim = 0.0
			currentLineStartIndex = i + 1
			imagesAspect = 0.0
			relativePaddingAspect = 0.0
			absolutePadding = 0.0
		}
	}
	if currentLineStartIndex != len(imagesInOrder) {
		currentLineWidth := (imagesAspect+relativePaddingAspect)*currentMinVarDim + absolutePadding
		badness.emptySpace = maxFixedDim - currentLineWidth
		centering := (maxFixedDim - currentLineWidth) / 2.0
		line := tileLine{images: imagesInOrder[currentLineStartIndex:], fixedDim: currentMinVarDim, startsAt: centering}
		tl.append(line)
		for _, jmg := range line.images {
			badness.scaledownSum += (currentLayout.DimensionsOf(jmg).Dim(varDim) / line.fixedDim) - 1.0
		}
		currentMinVarDim = 0
		imagesAspect = 0.0
		relativePaddingAspect = 0.0
		absolutePadding = 0.0
	} else {
		badness.emptySpace = 0
	}

	currentLayout, currentLinePosition, _ = finalizeTiling(currentLayout, tl, &badness, fixedDim, varDim)
	canvasSize := NewDims(0, 0)
	canvasSize.SetDim(fixedDim, maxFixedDim)
	canvasSize.SetDim(varDim, currentLinePosition)
	currentLayout.SetCanvasSize(canvasSize)
	var currentImageAspect float64 = math.Log2(canvasSize.X() / canvasSize.Y())
	badness.aspectRatioSkew = currentImageAspect - math.Log2(aspectRatio)
	il, err = currentLayout, nil
	return
}

func runTilingCheckBadness(imageLayout ImageLayout, imagesInOrder []ImageIdentifier, minDim, maxDim Dims, bComp badnessComparator, bestBadness *tileInOrder_Badness, bestSoFar *ImageLayout, dim float64) (ilOut ImageLayout, badness tileInOrder_Badness, bestBadnessChanged bool, err error) {
	createColumns := imageLayout.Parameters().OtherBool(TileInOrder_Columns)
	bestBadnessChanged = false
	var fixedDimD Dims
	if createColumns {
		fixedDimD = NewDims(0, dim)
	} else {
		fixedDimD = NewDims(dim, 0)
	}
	ilOut, badness, err = runOneTiling(imageLayout.Duplicate(), imagesInOrder, fixedDimD)
	if err != nil {
		return
	}
	x, y := ilOut.CanvasSize().X(), ilOut.CanvasSize().Y()
	if x >= minDim.X() && x <= maxDim.X() &&
		y >= minDim.Y() && y <= maxDim.Y() {
		if bComp.badnessLT(badness, *bestBadness) {
			*bestBadness = badness
			bestBadnessChanged = true
			*bestSoFar = ilOut
		}
	}
	return
}

type fringeEntry struct {
	dim     float64
	delta   float64
	badness tileInOrder_Badness
}

type fringeEntryComparator interface {
	fringeEntryLT(f1, f2 fringeEntry) bool
}

type fringeEntryComparator_LT_impl struct {
	bComp badnessComparator
}

func (fComp fringeEntryComparator_LT_impl) fringeEntryLT(f1, f2 fringeEntry) bool {
	if fComp.bComp.badnessLT(f1.badness, f2.badness) {
		return true
	} else if fComp.bComp.badnessLT(f2.badness, f1.badness) {
		return false
	} else {
		return f1.delta < f2.delta
	}
}

func findFirstGreaterThan(a []fringeEntry, length int, newMember fringeEntry, fComp fringeEntryComparator) int {
	min, max := 0, length-1
	for min <= max {
		midpoint := (min + max) / 2
		if a[midpoint] == newMember {
			return midpoint + 1
		}
		if fComp.fringeEntryLT(a[midpoint], newMember) {
			min = midpoint + 1
		} else {
			max = midpoint - 1
		}
	}
	return min
}

func insertPreservingSort(a []fringeEntry, length *int, newMember fringeEntry, fComp fringeEntryComparator) {
	index := findFirstGreaterThan(a, *length, newMember, fComp)
	copy(a[index+1:*length+1], a[index:*length])
	a[index] = newMember
	(*length)++
}

func findMinimumByBisection(imageLayout ImageLayout, imagesInOrder []ImageIdentifier, minDim, maxDim Dims, bComp badnessComparator, bestBadness *tileInOrder_Badness, bestSoFar *ImageLayout, fixedDim int) (ilOut ImageLayout, badness tileInOrder_Badness, err error) {
	delta := math.Floor(math.Log2(maxDim.Dim(fixedDim)-minDim.Dim(fixedDim))) - 5
	if delta < 1 {
		for j := minDim.Dim(fixedDim); j <= maxDim.Dim(fixedDim); j++ {
			ilOut, _, _, err = runTilingCheckBadness(imageLayout, imagesInOrder, minDim, maxDim, bComp, bestBadness, bestSoFar, j)
			if err != nil {
				return
			}
			imageLayout.Parameters().ProgressMonitor().ReportTileInOrderPositioningProgress(ilOut.CanvasSize(), *bestBadness)
		}
		ilOut = *bestSoFar
		badness = *bestBadness
		err = nil
		return
	}

	var badnesses []fringeEntry = make([]fringeEntry, int(maxDim.Dim(fixedDim)))
	var badnessesLength int = 0
	var badnessChanged bool
	var lastBadnessChanged int = 0
	fComp := fringeEntryComparator_LT_impl{bComp: bComp}
	delta = math.Exp2(delta)
	for j := minDim.Dim(fixedDim) + delta; j <= maxDim.Dim(fixedDim); j += delta {
		ilOut, badness, _, err = runTilingCheckBadness(imageLayout, imagesInOrder, minDim, maxDim, bComp, bestBadness, bestSoFar, j)
		if err != nil {
			return
		}
		insertPreservingSort(badnesses, &badnessesLength, fringeEntry{dim: j, delta: delta, badness: badness}, fComp)
		imageLayout.Parameters().ProgressMonitor().ReportTileInOrderPositioningProgress(ilOut.CanvasSize(), *bestBadness)
	}

	for i := 1; badnessesLength > 0; i++ {
		nextEntry := badnesses[0]
		copy(badnesses[0:badnessesLength-1], badnesses[1:badnessesLength])
		badnessesLength--
		nextEntry.delta /= 2
		if nextEntry.delta >= 1 {
			newEntry := fringeEntry{dim: nextEntry.dim - nextEntry.delta, delta: nextEntry.delta}
			ilOut, newEntry.badness, badnessChanged, err = runTilingCheckBadness(imageLayout, imagesInOrder, minDim, maxDim, bComp, bestBadness, bestSoFar, newEntry.dim)
			if err != nil {
				return
			}
			if badnessChanged {
				lastBadnessChanged = i
			} else if (*bestBadness).emptySpace == 0 && (i-lastBadnessChanged) > 2*int(maxDim.Dim(fixedDim))/len(imagesInOrder) && bestBadness != nil {
				break
			}

			if newEntry.badness.emptySpace < nextEntry.badness.emptySpace {
				insertPreservingSort(badnesses, &badnessesLength, newEntry, fComp)
				insertPreservingSort(badnesses, &badnessesLength, nextEntry, fComp)
			}
			imageLayout.Parameters().ProgressMonitor().ReportTileInOrderPositioningProgress(ilOut.CanvasSize(), *bestBadness)
		}
	}

	ilOut = *bestSoFar
	badness = *bestBadness
	err = nil
	return
}

func findMinimumByBinarySearch(imageLayout ImageLayout, imagesInOrder []ImageIdentifier, minDim, maxDim Dims, bComp badnessComparator, bestBadness *tileInOrder_Badness, bestSoFar *ImageLayout, fixedDim int) (ilOut ImageLayout, badness tileInOrder_Badness, err error) {
	min, max := minDim.Dim(fixedDim), maxDim.Dim(fixedDim)
	bCompAspect := badnessComparator_LT_impl{hasAspectRatioF: bComp.hasAspectRatio(), prioritizeAspectRatioF: true}
	badness = tileInOrder_Badness{aspectRatioSkew: math.Inf(0)}
	for (max - min) > 1 {
		midpoint := math.Round((min + max) / 2)
		ilOut, badness, _, err = runTilingCheckBadness(imageLayout, imagesInOrder, minDim, maxDim, bCompAspect, bestBadness, bestSoFar, midpoint)
		if err != nil {
			return
		}
		imageLayout.Parameters().ProgressMonitor().ReportTileInOrderPositioningProgress(ilOut.CanvasSize(), badness) //*bestBadness)
		if badness.aspectRatioSkew > 0 {
			max = midpoint
		} else {
			min = midpoint
		}
	}
	if !bComp.prioritizeAspectRatio() {
		start := (*bestSoFar).CanvasSize().Dim(fixedDim)
		for j := 0.0; start-j > minDim.Dim(fixedDim) && start+j < maxDim.Dim(fixedDim) && bestBadness.emptySpace > 0; j++ {
			ilOut, badness, _, err = runTilingCheckBadness(imageLayout, imagesInOrder, minDim, maxDim, bComp, bestBadness, bestSoFar, start-j)
			if err != nil {
				return
			}
			imageLayout.Parameters().ProgressMonitor().ReportTileInOrderPositioningProgress(ilOut.CanvasSize(), badness)
			ilOut, badness, _, err = runTilingCheckBadness(imageLayout, imagesInOrder, minDim, maxDim, bComp, bestBadness, bestSoFar, start+j)
			if err != nil {
				return
			}
			imageLayout.Parameters().ProgressMonitor().ReportTileInOrderPositioningProgress(ilOut.CanvasSize(), badness)
		}
	}
	ilOut = *bestSoFar
	badness = *bestBadness
	err = nil
	return
}

func calculatePositions_TileInOrder(imageLayout ImageLayout) (il ImageLayout, err error) {
	minDim := imageLayout.Parameters().MinCanvasSize()
	minDimN, _ /*maxDimN*/, sumsN := getDimensionsRange(imageLayout, true)
	maxDim := imageLayout.Parameters().MaxCanvasSize()

	var bestBadness tileInOrder_Badness
	var bestSoFar ImageLayout
	if minDim.X() == 0 {
		minDim.SetX(minDimN.X() / 2)
	}
	if minDim.Y() == 0 {
		minDim.SetY(minDimN.Y() / 2)
	}
	if maxDim.X() == 0 {
		maxDim.SetX(sumsN.X())
	}
	if maxDim.Y() == 0 {
		maxDim.SetY(sumsN.Y())
	}
	aspectRatio, preserveAspectRatio := imageLayout.Parameters().AspectRatio()
	prioritizeAspectRatio := !preserveAspectRatio
	hasAspectRatio := true
	if aspectRatio == 0.0 {
		hasAspectRatio = false
		prioritizeAspectRatio = false
		var aspectRatioG Geometry
		aspectRatioG, err = ParseGeometry(fmt.Sprintf("%fx%f", sumsN.X()/sumsN.Y(), 1.0))
		if err != nil {
			return
		}
		imageLayout.Parameters().SetAspectRatio(aspectRatioG)
	}
	var bComp badnessComparator = badnessComparator_LT_impl{hasAspectRatioF: hasAspectRatio, prioritizeAspectRatioF: prioritizeAspectRatio}
	createColumns := imageLayout.Parameters().OtherBool(TileInOrder_Columns)
	var fixedDim int
	if createColumns {
		fixedDim = 1
	} else {
		fixedDim = 0
	}
	var imagesInOrder []ImageIdentifier
	if imageLayout.Parameters().OtherBool(TileInOrder_ExactOrder) {
		imagesInOrder = imageLayout.Images(false)
	} else {
		imagesInOrder = imageLayout.Images(true)
		IISBy(func(lhs, rhs *ImageIdentifier) bool {
			return (imageLayout.DimensionsOf(*lhs).Dim(fixedDim) < imageLayout.DimensionsOf(*rhs).Dim(fixedDim))
		}).Sort(imagesInOrder)
	}

	// TODO: Allow cropping.
	for _, img := range imagesInOrder {
		if imageLayout.CroppingOf(img).HasSize() || imageLayout.CroppingOf(img).HasOffset() {
			il = imageLayout
			err = errors.New("TileInOrder does not support cropping")
			return
		}
	}
	bestBadness = tileInOrder_Badness{emptySpace: math.Inf(0), scaledownSum: math.Inf(0), aspectRatioSkew: math.Inf(0)}
	if !bComp.hasAspectRatio() {
		_, _, err = findMinimumByBisection(imageLayout, imagesInOrder, minDim, maxDim, bComp, &bestBadness, &bestSoFar, fixedDim)
	} else {
		_, _, err = findMinimumByBinarySearch(imageLayout, imagesInOrder, minDim, maxDim, bComp, &bestBadness, &bestSoFar, fixedDim)
	}

	if bestBadness.emptySpace == math.Inf(0) {
		err = errors.New("could not find a usable canvas size within these limits")
	} else {
		imageLayout.Parameters().ProgressMonitor().ReportTileInOrderPositioningProgress(bestSoFar.CanvasSize(), bestBadness)
		imageLayout.Parameters().ProgressMonitor().ReportPositioningSuccess()
	}
	il = bestSoFar
	return
}

// A PositionCalculator that places images in a tiling pattern.
type PositionCalculator_TileInOrder struct {
	p *positionCalculator_TileInOrder_Parameters
}
type positionCalculator_TileInOrder_Parameters struct {
	exactOrder    bool
	createColumns bool
}

func PositionCalculator_TileInOrder_Init() PositionCalculator_TileInOrder {
	return PositionCalculator_TileInOrder{new(positionCalculator_TileInOrder_Parameters)}
}

func (pcr PositionCalculator_TileInOrder) RegisterCustomParameters(parameters *Parameters) bool {
	flag.BoolVar(&(pcr.p.exactOrder), "exact-order", false, "(TileInOrder placement algorithm) Put images in the collage in exact parameter order")
	flag.BoolVar(&(pcr.p.createColumns), "columns", false, "(TileInOrder placement algorithm) Put images in the collage in columns instead of rows")
	return true
}

func (pcr PositionCalculator_TileInOrder) ParseCustomParameters(parameters *Parameters) bool {
	parameters.SetOther(TileInOrder_ExactOrder, pcr.p.exactOrder)
	if parameters.Padding().HasSize() && parameters.Padding().width.U != Percent {
		p := parameters.Padding()
		p.preserveAspectRatio = true
		parameters.SetPadding(p)
	}
	parameters.SetOther(TileInOrder_Columns, pcr.p.createColumns)
	return true
}

func (pcr PositionCalculator_TileInOrder) CalculatePositions(imageLayout ImageLayout) (il ImageLayout, err error) {
	il, err = calculatePositions_TileInOrder(imageLayout)
	return
}
