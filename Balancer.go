// This file contains auxiliary methods that run an interactive "balancing"
// algorithm over an existing image layout, putting each image in the center
// of the rectangle of blank space surrounding it.
package CollageCreator

import (
	"log"
	"math"
)

type imbalance struct {
	i             float64
	minBoundImage *ImageIdentifier
	maxBoundImage *ImageIdentifier
	minBound      float64
	maxBound      float64
}

type imbalances map[ImageIdentifier]*imbalance

func getImbalance(iLay ImageLayout, i ImageIdentifier, dimIndex int, imb imbalances) {
	iDim := iLay.PositionOf(i).Dim(dimIndex)
	imb[i].minBound = 0.0
	imb[i].maxBound = iLay.CanvasSize().Dim(dimIndex) - iLay.DimensionsOf(i).Dim(dimIndex)
	for _, j := range iLay.Images(false) {
		if i == j {
			continue
		}
		jDim := iLay.PositionOf(j).Dim(dimIndex)
		ov := Overlap(iLay, i, j)
		var oDim float64
		var oOpp float64
		if dimIndex == 0 {
			oDim = ov.X()
			oOpp = ov.Y()
		} else {
			oDim = ov.Y()
			oOpp = ov.X()
		}
		if oOpp >= 0 && oDim <= 0 {
			if iDim >= jDim && (iDim+oDim) > imb[i].minBound {
				imb[i].minBoundImage = &j
				imb[i].minBound = iDim + oDim
			} else if iDim <= jDim && (iDim-oDim) < imb[i].maxBound {
				imb[i].maxBoundImage = &j
				imb[i].maxBound = iDim - oDim
			}
		}
	}
	imb[i].i = imb[i].minBound + (imb[i].maxBound-imb[i].minBound)/2 - iDim
}

func getImbalances(iLay ImageLayout, dimIndex int, imb imbalances) {
	for _, i := range iLay.Images(false) {
		getImbalance(iLay, i, dimIndex, imb)
	}
}

func getNeighboringImbalances(iLay ImageLayout, lastBalanced ImageIdentifier, dimIndex int, imb imbalances) {
	getImbalance(iLay, lastBalanced, dimIndex, imb)
	if minBI := imb[lastBalanced].minBoundImage; minBI != nil {
		getImbalance(iLay, *minBI, dimIndex, imb)
	}
	if maxBI := imb[lastBalanced].minBoundImage; maxBI != nil {
		getImbalance(iLay, *maxBI, dimIndex, imb)
	}
}

// Run an iterative "balancing" algorithm on an existing image layout that attempts
// to center each image within the rectangle of blank space around it.
func Balance(iLay ImageLayout) ImageLayout {
	imb := make(imbalances)
	for _, img := range iLay.Images(false) {
		imb[img] = new(imbalance)
		imb[img].minBoundImage = nil
		imb[img].maxBoundImage = nil
	}
	imagesInOrder := iLay.Images(true)
	dimIndex := 0
	for balIt := 0; balIt < 2*iLay.Parameters().OtherInt(Balancer_MaxBalanceIterations); balIt++ {
		iterations := 0
		var i *ImageIdentifier = nil
		for {
			if i == nil {
				getImbalances(iLay, dimIndex, imb)
			} else {
				getNeighboringImbalances(iLay, *i, dimIndex, imb)
			}
			IISBy(func(lhs, rhs *ImageIdentifier) bool {
				return math.Abs(float64(imb[*rhs].i)) < math.Abs(float64(imb[*lhs].i))
			}).Sort(imagesInOrder)
			i = &imagesInOrder[0]
			gap := float64(imb[*i].maxBound - imb[*i].minBound)
			neededImbalance := iLay.Parameters().OtherFloat(Balancer_BalanceToleranceFactor) * gap
			iLay.Parameters().ProgressMonitor().ReportBalanceProgress((balIt/2)+1, dimIndex, iterations+1, int(math.Abs(float64(imb[*i].i))))
			if math.Abs(float64(imb[*i].i)) <= neededImbalance {
				break
			}
			pos := iLay.PositionOf(*i)
			pos.SetDim(dimIndex, pos.Dim(dimIndex)+imb[*i].i)
			_, result := iLay.SetPosition(*i, pos)
			if result != nil {
				iLay.Parameters().ProgressMonitor().ReportBalanceCollision(*i, iLay.ImageInfoOf(*i).FileName(), pos)
				log.Fatal()
			}
			iterations++
		}
		if dimIndex == 0 {
			dimIndex = 1
		} else {
			dimIndex = 0
		}
		iLay.Parameters().ProgressMonitor().ReportBalancingSuccess()
		if iterations == 0 {
			break
		}
	}
	return iLay
}
