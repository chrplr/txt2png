// txt2png
// Convert text to PNG image
//
// Usage:
// txt2png -text "JOJO" -fontfile /usr/share/fonts/truetype/liberation/LiberationSerif-Regular.ttf -dpi 72 -hinting none -size 125 -whiteonblack
//
// Original code: https://github.com/chrplr/txt2png
//
// License: GPL-3.0

package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "/usr/share/fonts/truetype/liberation/LiberationSerif-Regular.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	fontSize = flag.Float64("size", 125, "font size in points")
	wonb     = flag.Bool("whiteonblack", false, "white text on a black background")
	text     = flag.String("text", "JOJO", "text to render")
)

func main() {
	flag.Parse()

	fmt.Printf("Loading fontfile %q\n", *fontfile)
	fontBytes, err := os.ReadFile(*fontfile)
	if err != nil {
		log.Fatalf("Error reading font file: %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Fatalf("Error parsing font: %v", err)
	}

	// Calculate image dimensions based on text length
	// Each character is allocated a 250px wide slot
	const slotWidth = 250
	const imageHeight = 200
	width := len(*text) * slotWidth
	if width == 0 {
		width = slotWidth
	}
	rgba := image.NewRGBA(image.Rect(0, 0, width, imageHeight))

	// Define colors based on the 'wonb' flag
	fg, bg := image.Image(image.Black), image.Image(image.White)
	rulerColor := color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
	if *wonb {
		fg, bg = image.White, image.Black
		rulerColor = color.RGBA{0x44, 0x44, 0x44, 0xff}
	}

	// Fill background
	draw.Draw(rgba, rgba.Bounds(), bg, image.Point{}, draw.Src)

	// Draw vertical guidelines at each slot boundary
	for i := 0; i < len(*text); i++ {
		x := i * slotWidth
		for y := 0; y < imageHeight; y++ {
			rgba.Set(x, y, rulerColor)
		}
	}

	// Initialize Freetype context
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)

	switch *hinting {
	case "full":
		c.SetHinting(font.HintingFull)
	default:
		c.SetHinting(font.HintingNone)
	}

	// Create a font face to measure glyph advance widths
	opts := truetype.Options{
		Size: *fontSize,
		DPI:  *dpi,
	}
	face := truetype.NewFace(f, &opts)

	// Render each character centered in its slot
	for i, r := range *text {
		advance, ok := face.GlyphAdvance(r)
		if !ok {
			log.Printf("Warning: failed to get glyph advance for %q", r)
			continue
		}

		// Convert fixed-point advance to pixels
		glyphWidthPx := int(float64(advance) / 64)
		fmt.Printf("Char: %q, Width: %dpx\n", r, glyphWidthPx)

		// Calculate horizontal position to center glyph in slot
		xPos := i*slotWidth + (slotWidth/2 - glyphWidthPx/2)
		// 128 is the hardcoded baseline from the original code
		pt := freetype.Pt(xPos, 128)

		if _, err := c.DrawString(string(r), pt); err != nil {
			log.Printf("Error drawing %q: %v", r, err)
		}
	}

	// Save the resulting image to out.png
	const outFileName = "out.png"
	outFile, err := os.Create(outFileName)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outFile.Close()

	bWriter := bufio.NewWriter(outFile)
	if err := png.Encode(bWriter, rgba); err != nil {
		log.Fatalf("Error encoding PNG: %v", err)
	}

	if err := bWriter.Flush(); err != nil {
		log.Fatalf("Error flushing buffer: %v", err)
	}

	fmt.Printf("Successfully wrote %s\n", outFileName)
}
