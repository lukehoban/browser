// Package svg provides scanline rasterization for SVG shapes.
package svg

import (
	"image/color"
	"sort"
)

// Rasterizer provides polygon rasterization functionality.
type Rasterizer struct {
	Width  int
	Height int
}

// FillPolygon fills a polygon using the scanline algorithm.
// This implements a basic polygon fill algorithm suitable for simple shapes.
//
// Algorithm: For each horizontal scanline, find intersections with polygon edges,
// sort them, and fill between pairs of intersections.
//
// Reference: Computer Graphics: Principles and Practice (Foley et al.)
// Chapter on polygon fill algorithms.
func (r *Rasterizer) FillPolygon(points [][2]float64, fillFunc func(x, y int, col color.RGBA), col color.RGBA) {
	if len(points) < 3 {
		return
	}
	
	// Find bounding box
	minY, maxY := points[0][1], points[0][1]
	for _, p := range points {
		if p[1] < minY {
			minY = p[1]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}
	
	// Scanline algorithm
	for scanY := int(minY); scanY <= int(maxY); scanY++ {
		// Find intersections with polygon edges
		var intersections []float64
		
		for i := 0; i < len(points); i++ {
			j := (i + 1) % len(points)
			y1, y2 := points[i][1], points[j][1]
			x1, x2 := points[i][0], points[j][0]
			
			// Skip horizontal edges (y1 == y2) to avoid division by zero
			if y1 == y2 {
				continue
			}
			
			// Check if edge crosses scanline (excluding horizontal edges)
			if (y1 <= float64(scanY) && float64(scanY) < y2) || (y2 <= float64(scanY) && float64(scanY) < y1) {
				// Calculate x intersection (safe since y1 != y2)
				t := (float64(scanY) - y1) / (y2 - y1)
				x := x1 + t*(x2-x1)
				intersections = append(intersections, x)
			}
		}
		
		// Sort intersections
		sort.Float64s(intersections)
		
		// Fill between pairs of intersections
		for i := 0; i+1 < len(intersections); i += 2 {
			x1 := int(intersections[i])
			x2 := int(intersections[i+1])
			
			for x := x1; x <= x2; x++ {
				// Bounds checking
				if x >= 0 && x < r.Width && scanY >= 0 && scanY < r.Height {
					fillFunc(x, scanY, col)
				}
			}
		}
	}
}

// TransformPoints transforms points from viewBox coordinates to target dimensions.
// SVG 1.1 §7.7: The viewBox attribute defines a coordinate system transformation.
func TransformPoints(points [][2]float64, viewBox []float64, targetWidth, targetHeight int) [][2]float64 {
	if viewBox == nil || len(viewBox) != 4 {
		// No transformation needed if no viewBox
		return points
	}
	
	scaleX := float64(targetWidth) / (viewBox[2] - viewBox[0])
	scaleY := float64(targetHeight) / (viewBox[3] - viewBox[1])
	
	transformed := make([][2]float64, len(points))
	for i, p := range points {
		transformed[i] = [2]float64{
			(p[0] - viewBox[0]) * scaleX,
			(p[1] - viewBox[1]) * scaleY,
		}
	}
	
	return transformed
}
