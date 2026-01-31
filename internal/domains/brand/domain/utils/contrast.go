package utils

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// CalculateContrastRatio calculates WCAG contrast ratio between two colors
// Returns a value between 1 (no contrast) and 21 (maximum contrast)
// WCAG AA requires 4.5:1 for normal text, 3:1 for large text
func CalculateContrastRatio(color1, color2 string) float64 {
	lum1 := getRelativeLuminance(color1)
	lum2 := getRelativeLuminance(color2)

	// Ensure lighter color is numerator
	if lum1 < lum2 {
		lum1, lum2 = lum2, lum1
	}

	// Contrast ratio formula: (L1 + 0.05) / (L2 + 0.05)
	return (lum1 + 0.05) / (lum2 + 0.05)
}

// GetContrastingTextColor returns appropriate text color (white or dark) for a given background
// Returns "#ffffff" for dark backgrounds, "#1f2937" for light backgrounds
func GetContrastingTextColor(backgroundColor string) string {
	luminance := getRelativeLuminance(backgroundColor)
	
	// If background is dark (luminance < 0.5), use white text
	// If background is light (luminance >= 0.5), use dark text
	if luminance < 0.5 {
		return "#ffffff"
	}
	return "#1f2937"
}

// ValidateContrast validates if contrast meets minimum ratio
// Returns true if contrast is sufficient, false otherwise
func ValidateContrast(backgroundColor, textColor string, minRatio float64) bool {
	if minRatio <= 0 {
		minRatio = 4.5 // WCAG AA default
	}
	ratio := CalculateContrastRatio(backgroundColor, textColor)
	return ratio >= minRatio
}

// EnsureContrast ensures text color has sufficient contrast against background
// If contrast is insufficient and autoAdjust is true, returns adjusted color
// Otherwise returns original text color
func EnsureContrast(backgroundColor, textColor string, minRatio float64, autoAdjust bool) string {
	if minRatio <= 0 {
		minRatio = 4.5 // WCAG AA default
	}

	// Validate current contrast
	if ValidateContrast(backgroundColor, textColor, minRatio) {
		return textColor
	}

	// If auto-adjust is disabled, return original (caller should handle)
	if !autoAdjust {
		return textColor
	}

	// Auto-adjust: return contrasting color
	return GetContrastingTextColor(backgroundColor)
}

// getRelativeLuminance calculates the relative luminance of a color
// Returns a value between 0 (black) and 1 (white)
// Based on WCAG 2.1 formula: https://www.w3.org/WAI/GL/wiki/Relative_luminance
func getRelativeLuminance(color string) float64 {
	r, g, b := parseColor(color)
	
	// Convert RGB to linear values
	rs := linearizeRGB(r)
	gs := linearizeRGB(g)
	bs := linearizeRGB(b)
	
	// Calculate relative luminance
	// Formula: 0.2126*R + 0.7152*G + 0.0722*B
	return 0.2126*rs + 0.7152*gs + 0.0722*bs
}

// linearizeRGB converts RGB value to linear space for luminance calculation
func linearizeRGB(value float64) float64 {
	if value <= 0.03928 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}

// parseColor parses a color string (hex, rgb, rgba) and returns RGB values (0-1 range)
func parseColor(color string) (r, g, b float64) {
	color = strings.TrimSpace(color)
	color = strings.ToLower(color)

	// Handle hex colors (#RRGGBB or #RGB)
	if strings.HasPrefix(color, "#") {
		return parseHexColor(color)
	}

	// Handle rgb/rgba colors
	if strings.HasPrefix(color, "rgb") {
		return parseRGBColor(color)
	}

	// Default: assume black if parsing fails
	return 0, 0, 0
}

// parseHexColor parses hex color (#RRGGBB or #RGB)
func parseHexColor(hex string) (r, g, b float64) {
	hex = strings.TrimPrefix(hex, "#")
	
	// Handle short form (#RGB -> #RRGGBB)
	if len(hex) == 3 {
		hex = string(hex[0]) + string(hex[0]) + string(hex[1]) + string(hex[1]) + string(hex[2]) + string(hex[2])
	}

	if len(hex) != 6 {
		return 0, 0, 0
	}

	// Parse RGB values
	rVal, _ := strconv.ParseInt(hex[0:2], 16, 64)
	gVal, _ := strconv.ParseInt(hex[2:4], 16, 64)
	bVal, _ := strconv.ParseInt(hex[4:6], 16, 64)

	// Convert to 0-1 range
	return float64(rVal) / 255.0, float64(gVal) / 255.0, float64(bVal) / 255.0
}

// parseRGBColor parses rgb/rgba color string
func parseRGBColor(rgb string) (r, g, b float64) {
	// Extract numbers using regex
	re := regexp.MustCompile(`\d+\.?\d*`)
	matches := re.FindAllString(rgb, -1)

	if len(matches) < 3 {
		return 0, 0, 0
	}

	// Parse RGB values
	rVal, _ := strconv.ParseFloat(matches[0], 64)
	gVal, _ := strconv.ParseFloat(matches[1], 64)
	bVal, _ := strconv.ParseFloat(matches[2], 64)

	// Normalize to 0-1 range (assuming 0-255 input)
	if rVal > 1.0 || gVal > 1.0 || bVal > 1.0 {
		rVal /= 255.0
		gVal /= 255.0
		bVal /= 255.0
	}

	return rVal, gVal, bVal
}
