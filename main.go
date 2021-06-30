package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {

	const w int = 512
	const h int = 512

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("GO Raytracer", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(w), int32(h), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGB888, sdl.TEXTUREACCESS_STREAMING,
		int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	pixels, pitch, err := texture.Lock(nil)
	if err != nil {
		panic(err)
	}

	bpp := pitch / w // bytes-per-pixel
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			ind := (j * pitch) + (i * bpp)
			r := float64(i) / float64(w-1)
			g := float64(j) / float64(h-1)
			b := 0.25

			ir := uint8(255.99 * r)
			ig := uint8(255.99 * g)
			ib := uint8(255.99 * b)
			pixels[ind] = ib   // B
			pixels[ind+1] = ig // G
			pixels[ind+2] = ir // R
		}
	}

	texture.Update(nil, pixels, pitch)
	texture.Unlock()

	renderer.Clear()
	renderer.Copy(texture, nil, nil)
	renderer.Present()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quit")
				running = false
				break
			}
		}
	}
}
