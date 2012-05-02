// Some of this code (namely the code for computing the
// width of a string in a given font) was copied from
// code.google.com/p/freetype-go/freetype/ which includes
// the following copyright notice:
// Copyright 2010 The Freetype-Go Authors. All rights reserved.

package vecgfx

import (
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// importString is the import string expected for
	// this package.  It is used to find the font
	// directory included with the package source.
	importString = "plotinum/vecgfx"
)

var (
	// FontMap maps Postscript/PDF font names to compatible
	// free fonts (TrueType converted ghostscript fonts).
	// Fonts that are not in this map are not supported.
	FontMap = map[string]string{

		// At the moment, we use fonts from GNU's freefont
		// project.

		"Courier":             "NimbusMonL-Regu",
		"Courier-Bold":        "NimbusMonL-Bold",
		"Courier-Oblique":     "NimbusMonL-ReguObli",
		"Courier-BoldOblique": "NimbusMonL-BoldObli",

		"Helvetica":             "NimbusSanL-Regu",
		"Helvetica-Bold":        "NimbusSanL-Bold",
		"Helvetica-Oblique":     "NimbusSanL-ReguItal",
		"Helvetica-BoldOblique": "NimbusSanL-BoldItal",

		"Times-Roman":      "NimbusRomNo9L-Regu",
		"Times-Bold":       "NimbusRomNo9L-Medi",
		"Times-Italic":     "NimbusRomNo9L-ReguItal",
		"Times-BoldItalic": "NimbusRomNo9L-MediItal",
	}

	// loadedFonts is indexed by a font name and it
	// caches the associated *truetype.Font.
	loadedFonts = make(map[string]*truetype.Font)
)

// A Font represents one of the supported font
// faces.
type Font struct {
	// Size is the size of the font in points.
	Size float64

	// name is the name of this font.
	name string

	// This is a little bit of a hack, but the truetype
	// font is currently only needed to determine the
	// dimensions of strings drawn in this font.
	// The actual drawing of the strings is handled
	// separately by different back-ends:
	// Both Postscript and PDF are capable of drawing
	// their own fonts and draw2d loads its own copy of
	// the truetype fonts for its own output.
	//
	// This isn't a necessity--some future backend is
	// free to use this field--however it is a consequence
	// of the fact that the current backends were
	// developed independently of this package.

	// font is the truetype font pointer for this
	// font.
	font *truetype.Font
}

// MakeFont returns a font object.
// The font size is specified in points.
func MakeFont(name string, size float64) (font Font, err error) {
	font.Size = size
	font.name = name
	font.font, err = getFont(name)
	return
}

// Name returns the name of the font.
func (f *Font) Name() string {
	return f.name
}

// Font returns the corresponding truetype.Font.
func (f *Font) Font() *truetype.Font {
	return f.font
}

// SetName set the name of the font to name.  If an
// error is returned then the font is left unchanged.
func (f *Font) SetName(name string) error {
	font, err := getFont(name)
	if err != nil {
		return err
	}
	f.name = name
	f.font = font
	return nil
}

// getFont returns the truetype.Font for the given font
// name.
func getFont(name string) (*truetype.Font, error) {
	if f, ok := loadedFonts[name]; ok {
		return f, nil
	}

	n, ok := FontMap[name]
	if !ok {
		return nil, fmt.Errorf("No matching font: %s", name)
	}

	pkg, err := build.Import(importString, "", build.FindOnly)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(pkg.Dir, "fonts", n+".ttf")

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	font, err := freetype.ParseFont(bytes)
	if err != nil {
		loadedFonts[name] = font
	}
	return font, err
}

// FontExtents has font metric information.  All
// values are given in postscript points.
type FontExtents struct{
	// Ascent is the distance that the text
	// extends above the baseline.
	Ascent float64

	// Descent is the distance that the text
	// extends below the baseline.  The descent
	// is given as a negative value.
	Descent float64

	// Height is the distance from the lowest
	// descending point to the highest ascending
	// point.
	Height float64
}

// Height returns height of a string when rendered
// using this font.
func (f *Font) Extents() FontExtents {
	bounds := f.font.Bounds()
	scale := f.Size / float64(f.Font().UnitsPerEm())
	return FontExtents{
		Ascent: float64(bounds.YMax)*scale,
		Descent: float64(bounds.YMin)*scale,
		Height: float64(bounds.YMax-bounds.YMin)*scale,
	}
}

// Width returns width of a string when rendered
// using this font.
func (f *Font) Width(s string) float64 {
	// scale converts truetype.FUnit to float64
	scale := f.Size / float64(f.font.UnitsPerEm())

	width := 0
	prev, hasPrev := truetype.Index(0), false
	for _, rune := range s {
		index := f.font.Index(rune)
		if hasPrev {
			width += int(f.font.Kerning(prev, index))
		}
		width += int(f.font.HMetric(index).AdvanceWidth)
		prev, hasPrev = index, true
	}
	return float64(width) * scale
}
