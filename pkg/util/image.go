package util

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func SaveToImage(name string, width, height int, pixels []byte) {
	pitch := len(pixels) / height
	bpp := pitch / width

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ind := (y * pitch) + (x * bpp)
			img.Set(x, y, color.NRGBA{
				R: pixels[ind+2],
				G: pixels[ind+1],
				B: pixels[ind],
				A: 255,
			})
		}
	}
	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("Image", name, "saved")
}
