// Package svg provides a minimal SVG parser and rasterizer.
// Implements a subset of the SVG 1.1 specification sufficient for rendering
// simple vector graphics like icons and basic shapes.
//
// SVG 1.1 Specification: https://www.w3.org/TR/SVG11/
//
// Supported features:
// - Basic shapes via <path> element (SVG 1.1 §8.3)
// - Path data commands: moveto, lineto, horizontal/vertical lineto, closepath (SVG 1.1 §8.3.2-8.3.4)
// - Fill colors via fill attribute (SVG 1.1 §11.2)
// - Coordinate system via viewBox (SVG 1.1 §7.7)
// - Multiple path elements in a single SVG
package svg

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/lukehoban/browser/log"
)

// SVGPath represents a single path element with its fill color.
type SVGPath struct {
	Points    [][2]float64 // Polygon points from parsed path data
	FillColor color.RGBA
}

// ParsedSVG represents a parsed SVG with minimal data needed for rasterization.
type ParsedSVG struct {
	ViewBox    []float64    // [minX, minY, width, height]
	PathPoints [][2]float64 // Polygon points from first parsed path (for backward compatibility)
	FillColor  color.RGBA   // Fill color of first path (for backward compatibility)
	Paths      []SVGPath    // All paths in the SVG
}

// Parse parses SVG data and extracts the minimal information needed for rendering.
// SVG 1.1 §5: Document structure
func Parse(svgData []byte) (*ParsedSVG, error) {
	svgStr := string(svgData)

	// Extract viewBox dimensions for coordinate transformation
	// SVG 1.1 §7.7: The viewBox attribute
	viewBox := parseViewBox(svgStr)

	// Extract all path elements
	// SVG 1.1 §8.3: The 'path' element
	paths := extractAllPaths(svgStr)
	if len(paths) == 0 {
    log.Debug("SVG: no path data found")
		return nil, nil
	}

	result := &ParsedSVG{
		ViewBox: viewBox,
		Paths:   paths,
	}

	// For backward compatibility, set the first path's data
	if len(paths) > 0 {
		result.PathPoints = paths[0].Points
		result.FillColor = paths[0].FillColor
	}

	return result, nil
}

// parseViewBox extracts viewBox dimensions from SVG string.
// SVG 1.1 §7.7: The 'viewBox' attribute establishes a user coordinate system.
// Format: viewBox="min-x min-y width height"
func parseViewBox(svg string) []float64 {
	// Look for viewBox="x y width height"
	start := strings.Index(svg, "viewBox=\"")
	if start == -1 {
		start = strings.Index(svg, "viewBox='")
		if start == -1 {
			return nil
		}
		start += 9 // len("viewBox='")
	} else {
		start += 9 // len("viewBox=\"")
	}
	
	end := start
	for end < len(svg) && svg[end] != '"' && svg[end] != '\'' {
		end++
	}
	
	viewBoxStr := svg[start:end]
	parts := strings.Fields(viewBoxStr)
	if len(parts) != 4 {
		log.Debug("SVG: invalid viewBox format")
		return nil
	}
	
	result := make([]float64, 4)
	for i, p := range parts {
		val, err := strconv.ParseFloat(p, 64)
		if err != nil {
			log.Warnf("SVG: failed to parse viewBox coordinate: %v", err)
			return nil
		}
		result[i] = val
	}
	
	return result
}

// extractAllPaths extracts all path elements from an SVG string.
// SVG 1.1 §8.3: The 'path' element
func extractAllPaths(svg string) []SVGPath {
	var paths []SVGPath
	remaining := svg

	for {
		// Find the next <path element
		pathStart := strings.Index(remaining, "<path")
		if pathStart == -1 {
			break
		}

		// Find the end of this path element
		pathEnd := strings.Index(remaining[pathStart:], "/>")
		if pathEnd == -1 {
			// Try finding </path>
			pathEnd = strings.Index(remaining[pathStart:], "</path>")
			if pathEnd == -1 {
				break
			}
			pathEnd += 7 // len("</path>")
		} else {
			pathEnd += 2 // len("/>")
		}

		pathElement := remaining[pathStart : pathStart+pathEnd]

		// Extract d attribute from this path element
		pathData := extractPathDataFromElement(pathElement)
		if pathData != "" {
			// Extract fill color from this path element
			fillColor := extractFillColorFromElement(pathElement)

			// Parse path commands to get polygon points
			points := parsePath(pathData)
			if len(points) >= 3 {
				paths = append(paths, SVGPath{
					Points:    points,
					FillColor: fillColor,
				})
			}
		}

		// Move past this path element
		remaining = remaining[pathStart+pathEnd:]
	}

	return paths
}

// extractPathDataFromElement extracts the 'd' attribute from a single path element.
func extractPathDataFromElement(pathElement string) string {
	// Look for d="..." in path element
	start := strings.Index(pathElement, " d=\"")
	if start == -1 {
		start = strings.Index(pathElement, " d='")
		if start == -1 {
			return ""
		}
		start += 4 // len(" d='")
	} else {
		start += 4 // len(" d=\"")
	}

	end := start
	for end < len(pathElement) && pathElement[end] != '"' && pathElement[end] != '\'' {
		end++
	}

	return pathElement[start:end]
}

// extractFillColorFromElement extracts the fill attribute from a single element.
func extractFillColorFromElement(element string) color.RGBA {
	// Look for fill="..."
	start := strings.Index(element, "fill=\"")
	if start == -1 {
		start = strings.Index(element, "fill='")
		if start == -1 {
			return color.RGBA{0, 0, 0, 255} // default black per SVG 1.1 §11.2
		}
		start += 6 // len("fill='")
	} else {
		start += 6 // len("fill=\"")
	}

	end := start
	for end < len(element) && element[end] != '"' && element[end] != '\'' {
		end++
	}

	fillStr := element[start:end]
	return parseColor(fillStr)
}

// parsePath parses SVG path data and returns polygon points.
// SVG 1.1 §8.3.2: Path data consists of commands (letters) followed by parameters (numbers).
//
// Supported commands:
// - M/m: moveto (SVG 1.1 §8.3.2)
// - L/l: lineto (SVG 1.1 §8.3.3)
// - H/h: horizontal lineto (SVG 1.1 §8.3.3)
// - V/v: vertical lineto (SVG 1.1 §8.3.3)
// - Z/z: closepath (SVG 1.1 §8.3.4)
func parsePath(pathData string) [][2]float64 {
	var points [][2]float64
	var currentX, currentY float64
	var startX, startY float64
	
	// Tokenize the path data
	pathData = strings.TrimSpace(pathData)
	i := 0
	
	for i < len(pathData) {
		// Skip whitespace
		for i < len(pathData) && (pathData[i] == ' ' || pathData[i] == ',' || pathData[i] == '\t' || pathData[i] == '\n' || pathData[i] == '\r') {
			i++
		}
		if i >= len(pathData) {
			break
		}
		
		cmd := pathData[i]
		i++
		
		switch cmd {
		case 'M': // Absolute moveto - SVG 1.1 §8.3.2
			nums := extractNumbers(pathData, &i)
			if len(nums) >= 2 {
				currentX = nums[0]
				currentY = nums[1]
				startX = currentX
				startY = currentY
				points = append(points, [2]float64{currentX, currentY})
				
				// After moveto, subsequent coordinate pairs are implicit lineto commands
				// SVG 1.1 §8.3.2: "If a moveto is followed by multiple pairs of coordinates,
				// the subsequent pairs are treated as implicit lineto commands."
				for j := 2; j+1 < len(nums); j += 2 {
					currentX = nums[j]
					currentY = nums[j+1]
					points = append(points, [2]float64{currentX, currentY})
				}
			}
			
		case 'm': // Relative moveto - SVG 1.1 §8.3.2
			nums := extractNumbers(pathData, &i)
			if len(nums) >= 2 {
				currentX += nums[0]
				currentY += nums[1]
				startX = currentX
				startY = currentY
				points = append(points, [2]float64{currentX, currentY})
				
				// After moveto, subsequent coordinate pairs are implicit lineto commands
				for j := 2; j+1 < len(nums); j += 2 {
					currentX += nums[j]
					currentY += nums[j+1]
					points = append(points, [2]float64{currentX, currentY})
				}
			}
			
		case 'L': // Absolute lineto - SVG 1.1 §8.3.3
			nums := extractNumbers(pathData, &i)
			for j := 0; j+1 < len(nums); j += 2 {
				currentX = nums[j]
				currentY = nums[j+1]
				points = append(points, [2]float64{currentX, currentY})
			}
			
		case 'l': // Relative lineto - SVG 1.1 §8.3.3
			nums := extractNumbers(pathData, &i)
			for j := 0; j+1 < len(nums); j += 2 {
				currentX += nums[j]
				currentY += nums[j+1]
				points = append(points, [2]float64{currentX, currentY})
			}

		case 'H': // Absolute horizontal lineto - SVG 1.1 §8.3.3
			nums := extractNumbers(pathData, &i)
			for _, x := range nums {
				currentX = x
				points = append(points, [2]float64{currentX, currentY})
			}

		case 'h': // Relative horizontal lineto - SVG 1.1 §8.3.3
			nums := extractNumbers(pathData, &i)
			for _, dx := range nums {
				currentX += dx
				points = append(points, [2]float64{currentX, currentY})
			}

		case 'V': // Absolute vertical lineto - SVG 1.1 §8.3.3
			nums := extractNumbers(pathData, &i)
			for _, y := range nums {
				currentY = y
				points = append(points, [2]float64{currentX, currentY})
			}

		case 'v': // Relative vertical lineto - SVG 1.1 §8.3.3
			nums := extractNumbers(pathData, &i)
			for _, dy := range nums {
				currentY += dy
				points = append(points, [2]float64{currentX, currentY})
			}

		case 'Z', 'z': // closepath - SVG 1.1 §8.3.4
			// Close the current subpath by drawing a straight line from
			// the current point to the initial point of the current subpath
			if len(points) > 0 && (currentX != startX || currentY != startY) {
				points = append(points, [2]float64{startX, startY})
			}
		}
	}
	
	return points
}

// extractNumbers extracts consecutive numbers from path data.
// SVG 1.1 §8.3.1: Path data can contain numbers separated by whitespace or commas.
func extractNumbers(pathData string, i *int) []float64 {
	var numbers []float64
	
	for *i < len(pathData) {
		// Skip whitespace and commas
		for *i < len(pathData) && (pathData[*i] == ' ' || pathData[*i] == ',' || pathData[*i] == '\t' || pathData[*i] == '\n' || pathData[*i] == '\r') {
			*i++
		}
		
		// Check if we hit a command letter
		if *i < len(pathData) && ((pathData[*i] >= 'A' && pathData[*i] <= 'Z') || (pathData[*i] >= 'a' && pathData[*i] <= 'z')) {
			break
		}
		
		if *i >= len(pathData) {
			break
		}
		
		// Extract number (handle negative numbers that may appear without space)
		start := *i
		if *i < len(pathData) && (pathData[*i] == '-' || pathData[*i] == '+') {
			*i++
		}
		
		hasDigit := false
		for *i < len(pathData) && ((pathData[*i] >= '0' && pathData[*i] <= '9') || pathData[*i] == '.') {
			hasDigit = true
			*i++
		}
		
		if hasDigit {
			num, err := strconv.ParseFloat(pathData[start:*i], 64)
			if err == nil {
				numbers = append(numbers, num)
			}
		} else {
			break
		}
	}
	
	return numbers
}

// parseColor parses a color string to RGBA.
// Supports hex colors (#RGB, #RRGGBB) per SVG 1.1 §4.2.
func parseColor(value string) color.RGBA {
	value = strings.TrimSpace(strings.ToLower(value))
	
	// Handle hex colors
	if strings.HasPrefix(value, "#") {
		return parseHexColor(value)
	}
	
	// Default to black
	return color.RGBA{0, 0, 0, 255}
}

// parseHexColor parses a hex color string (#RGB or #RRGGBB).
// SVG 1.1 §4.2: Color syntax follows CSS2 specification.
func parseHexColor(hex string) color.RGBA {
	hex = strings.TrimPrefix(hex, "#")
	
	var r, g, b uint8
	
	switch len(hex) {
	case 3: // #RGB
		if rr, err := strconv.ParseUint(string(hex[0])+string(hex[0]), 16, 8); err == nil {
			r = uint8(rr)
		}
		if gg, err := strconv.ParseUint(string(hex[1])+string(hex[1]), 16, 8); err == nil {
			g = uint8(gg)
		}
		if bb, err := strconv.ParseUint(string(hex[2])+string(hex[2]), 16, 8); err == nil {
			b = uint8(bb)
		}
	case 6: // #RRGGBB
		if rr, err := strconv.ParseUint(hex[0:2], 16, 8); err == nil {
			r = uint8(rr)
		}
		if gg, err := strconv.ParseUint(hex[2:4], 16, 8); err == nil {
			g = uint8(gg)
		}
		if bb, err := strconv.ParseUint(hex[4:6], 16, 8); err == nil {
			b = uint8(bb)
		}
	}
	
	return color.RGBA{r, g, b, 255}
}

// IsSVG checks if the given data appears to be SVG content.
// Detection follows SVG 1.1 §5.1: An SVG document fragment is defined by an 'svg' element.
// Also checks for common XML declarations that precede SVG content.
// https://www.w3.org/TR/SVG11/struct.html#NewDocument
func IsSVG(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Limit search to first 1024 bytes for efficiency
	// SVG root element and XML declaration appear early in the file
	searchLen := len(data)
	if searchLen > 1024 {
		searchLen = 1024
	}
	content := string(data[:searchLen])
	trimmed := strings.TrimSpace(content)

	// Check for <svg element directly at the start
	if strings.HasPrefix(trimmed, "<svg") {
		return true
	}

	// Check for XML declaration followed by SVG content
	// SVG files often start with <?xml ... ?> before <svg>
	if strings.HasPrefix(trimmed, "<?xml") && strings.Contains(content, "<svg") {
		return true
	}

	return false
}

// IsSVGFile checks if a filename indicates an SVG file.
// Per W3C media type registration, SVG files use .svg extension.
// https://www.w3.org/TR/SVGTiny12/mimereg.html
func IsSVGFile(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".svg")
}

// Render renders the parsed SVG to a canvas using the provided fill function.
// This encapsulates the rendering loop for SVG paths, handling viewBox
// transformation and rasterization per SVG 1.1 §7.7 (viewBox) and §8.3 (paths).
func Render(svgData []byte, width, height int, fillFunc func(x, y int, col color.RGBA)) error {
	if width <= 0 || height <= 0 {
		return nil
	}

	// Parse the SVG
	parsed, err := Parse(svgData)
	if err != nil || parsed == nil {
		return err
	}

	// Set default viewBox if not specified
	viewBox := parsed.ViewBox
	if viewBox == nil {
		viewBox = []float64{0, 0, float64(width), float64(height)}
	}

	// Create rasterizer
	rasterizer := Rasterizer{Width: width, Height: height}

	// Render all paths in order (first path is background, subsequent paths are foreground)
	for _, path := range parsed.Paths {
		if len(path.Points) < 3 {
			continue
		}
		// Transform points from viewBox coordinates to target dimensions
		transformedPoints := TransformPoints(path.Points, viewBox, width, height)
		// Rasterize the polygon
		rasterizer.FillPolygon(transformedPoints, fillFunc, path.FillColor)
	}

	return nil
}
