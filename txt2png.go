// txt2png
// Convert text to PNG image
//
// Usage:
// txt2png -text "JOJO" -fontfile ./LiberationMono-Regular.ttf -dpi 72 -hinting none -size 125 -whiteonblack -out out.png -slotwidth 250 -height 200
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
	dpi            = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile       = flag.String("fontfile", "./LiberationMono-Regular.ttf", "filename of the ttf font")
	hinting        = flag.String("hinting", "none", "none | full")
	fontSize       = flag.Float64("size", 125, "font size in points")
	wonb           = flag.Bool("whiteonblack", false, "white text on a black background")
	text           = flag.String("text", "TEST", "text to render")
	outFile        = flag.String("out", "out.png", "output PNG filename")
	slotWidth      = flag.Int("slotwidth", 120, "width of each character slot in pixels")
	imageHeight    = flag.Int("height", 120, "height of the image in pixels")
	showGuidelines = flag.Bool("guidelines", false, "draw vertical guidelines between character slots")
	verbose        = flag.Bool("verbose", false, "print informational messages to the console")
)

func main() {
	flag.Parse()

	f := loadFont(*fontfile, *verbose)

	fg, bg, rulerColor := getColors(*wonb)

	rgba := createImage(len(*text), *slotWidth, *imageHeight, bg, rulerColor, *showGuidelines)

	c := getFreeTypeContext(f, *dpi, *fontSize, rgba, fg, *hinting)

	renderText(c, f, *text, *slotWidth, *imageHeight, *dpi, *fontSize, *verbose)

	saveImage(*outFile, rgba)

	if *verbose {
		fmt.Printf("Successfully wrote %s\n", *outFile)
	}
}

func loadFont(path string, verb bool) *truetype.Font {
	if verb {
		fmt.Printf("Loading fontfile %q\n", path)
	}
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading font file: %v", err)
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Fatalf("Error parsing font: %v", err)
	}
	return f
}

func getColors(whiteOnBlack bool) (fg, bg image.Image, ruler color.Color) {
	fg, bg = image.Image(image.Black), image.Image(image.White)
	ruler = color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
	if whiteOnBlack {
		fg, bg = image.White, image.Black
		ruler = color.RGBA{0x44, 0x44, 0x44, 0xff}
	}
	return
}

func createImage(textLen, slotW, imgH int, bg image.Image, rulerColor color.Color, showGuidelines bool) *image.RGBA {
	width := textLen * slotW
	if width == 0 {
		width = slotW
	}
	rgba := image.NewRGBA(image.Rect(0, 0, width, imgH))
	draw.Draw(rgba, rgba.Bounds(), bg, image.Point{}, draw.Src)

	// Vertical guidelines
	if showGuidelines {
		for i := 0; i < textLen; i++ {
			x := i * slotW
			for y := 0; y < imgH; y++ {
				rgba.Set(x, y, rulerColor)
			}
		}
	}
	return rgba
}

func getFreeTypeContext(f *truetype.Font, dpi, size float64, dst *image.RGBA, src image.Image, hintingStr string) *freetype.Context {
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(f)
	c.SetFontSize(size)
	c.SetClip(dst.Bounds())
	c.SetDst(dst)
	c.SetSrc(src)

	switch hintingStr {
	case "full":
		c.SetHinting(font.HintingFull)
	default:
		c.SetHinting(font.HintingNone)
	}
	return c
}

func renderText(c *freetype.Context, f *truetype.Font, text string, slotW, imgH int, dpi, size float64, verb bool) {
	opts := truetype.Options{
		Size: size,
		DPI:  dpi,
	}
	face := truetype.NewFace(f, &opts)

	for i, r := range text {
		advance, ok := face.GlyphAdvance(r)
		if !ok {
			log.Printf("Warning: failed to get glyph advance for %q", r)
			continue
		}

		glyphWidthPx := int(float64(advance) / 64)
		if verb {
			fmt.Printf("Char: %q, Width: %dpx\n", r, glyphWidthPx)
		}

		xPos := i*slotW + (slotW/2 - glyphWidthPx/2)
		pt := freetype.Pt(xPos, imgH*2/3)

		if _, err := c.DrawString(string(r), pt); err != nil {
			log.Printf("Error drawing %q: %v", r, err)
		}
	}
}

func saveImage(path string, rgba *image.RGBA) {
	out, err := os.Create(path)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer out.Close()

	bWriter := bufio.NewWriter(out)
	if err := png.Encode(bWriter, rgba); err != nil {
		log.Fatalf("Error encoding PNG: %v", err)
	}

	if err := bWriter.Flush(); err != nil {
		log.Fatalf("Error flushing buffer: %v", err)
	}
}
