package render

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/style"
)

func TestDrawImage(t *testing.T) {
	// Create a test canvas
	c := NewCanvas(100, 100)
	c.Clear(color.RGBA{255, 255, 255, 255})

	// Create a simple test image (10x10 red square)
	testImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	red := color.RGBA{255, 0, 0, 255}
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			testImg.Set(x, y, red)
		}
	}

	// Draw the image at position (20, 20) scaled to 30x30
	c.DrawImage(testImg, 20, 20, 30, 30)

	// Check that pixels inside the drawn area are red
	if c.Pixels[25*100+25] != red {
		t.Errorf("expected red inside drawn image, got %v", c.Pixels[25*100+25])
	}

	// Check that pixels outside the drawn area are white
	white := color.RGBA{255, 255, 255, 255}
	if c.Pixels[10*100+10] != white {
		t.Errorf("expected white outside drawn image, got %v", c.Pixels[10*100+10])
	}
}

func TestDrawImageWithAlpha(t *testing.T) {
	// Create a test canvas with blue background
	c := NewCanvas(100, 100)
	blue := color.RGBA{0, 0, 255, 255}
	c.Clear(blue)

	// Create a semi-transparent red image
	testImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	semiRed := color.RGBA{255, 0, 0, 128} // 50% alpha
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			testImg.Set(x, y, semiRed)
		}
	}

	// Draw the image at position (20, 20) scaled to 10x10
	c.DrawImage(testImg, 20, 20, 10, 10)

	// Check that the pixel is blended (should be purple-ish)
	pixel := c.Pixels[25*100+25]
	// With 50% alpha blending of red over blue, we should get some purple
	if pixel.R < 50 || pixel.B < 50 {
		t.Errorf("expected blended color, got %v", pixel)
	}
}

func TestLoadImage(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a simple test PNG image
	testImgPath := filepath.Join(tmpDir, "test.png")
	testImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	red := color.RGBA{255, 0, 0, 255}
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			testImg.Set(x, y, red)
		}
	}

	// Save the test image
	f, err := os.Create(testImgPath)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, testImg); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}
	f.Close()

	// Create a canvas and load the image
	c := NewCanvas(100, 100)
	c.BaseDir = tmpDir

	img, err := c.LoadImage("test.png")
	if err != nil {
		t.Fatalf("failed to load image: %v", err)
	}

	// Check image dimensions
	bounds := img.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("expected 10x10 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Check that the image is cached
	img2, err := c.LoadImage("test.png")
	if err != nil {
		t.Fatalf("failed to load cached image: %v", err)
	}

	// Should be the same object from cache
	if img != img2 {
		t.Errorf("expected cached image to be the same object")
	}
}

func TestRenderImage(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a simple test PNG image
	testImgPath := filepath.Join(tmpDir, "test.png")
	testImg := image.NewRGBA(image.Rect(0, 0, 20, 20))
	red := color.RGBA{255, 0, 0, 255}
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			testImg.Set(x, y, red)
		}
	}

	// Save the test image
	f, err := os.Create(testImgPath)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	if err := png.Encode(f, testImg); err != nil {
		f.Close()
		t.Fatalf("failed to encode test image: %v", err)
	}
	f.Close()

	// Create a DOM node for the img element
	imgNode := dom.NewElement("img")
	imgNode.SetAttribute("src", "test.png")

	// Create a styled node
	styledNode := &style.StyledNode{
		Node: imgNode,
		Styles: map[string]string{
			"width":  "40px",
			"height": "40px",
		},
		Children: []*style.StyledNode{},
	}

	// Create a layout box for the image
	box := &layout.LayoutBox{
		BoxType:    layout.BlockBox,
		StyledNode: styledNode,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{
				X:      10,
				Y:      10,
				Width:  40,
				Height: 40,
			},
		},
		Children: []*layout.LayoutBox{},
	}

	// Render the layout box
	canvas := Render(box, 100, 100, tmpDir)

	// Check that the image was rendered (pixels should be red)
	pixel := canvas.Pixels[20*100+20] // Inside the rendered image area
	if pixel != red {
		t.Errorf("expected red pixel from rendered image, got %v", pixel)
	}

	// Check that pixels outside the image are white (background)
	white := color.RGBA{255, 255, 255, 255}
	outsidePixel := canvas.Pixels[5*100+5]
	if outsidePixel != white {
		t.Errorf("expected white pixel outside image, got %v", outsidePixel)
	}
}
