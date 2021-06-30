package main

import (
	"fmt"

	"github.com/gabrielfvale/go-traytracer/pkg/geom"
	"github.com/gabrielfvale/go-traytracer/pkg/img"
	"github.com/veandco/go-sdl2/sdl"
)

func color(r geom.Ray) img.RGB {
	t := 0.5 * (r.Dir.Y() + 1.0)
	c1 := img.NewRGB(1.0, 1.0, 1.0).Scale(1.0 - t)
	c2 := img.NewRGB(0.5, 0.7, 1.0).Scale(t)
	return c1.Plus(c2)
}

func main() {

	const aspect float64 = 16 / 9
	const width int = 640
	const height int = int(float64(width) / aspect)
	fmt.Println(width, height)

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("GO Raytracer", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(width), int32(height), sdl.WINDOW_SHOWN)
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
		int32(width), int32(height))
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	pixels, pitch, err := texture.Lock(nil)
	if err != nil {
		panic(err)
	}

	// Camera
	viewportHeight := 2.0
	viewportWidth := aspect * viewportHeight
	focalLength := 1.0

	origin := geom.NewVec3(0.0, 0.0, 0.0)
	horizontal := geom.NewVec3(viewportWidth, 0, 0)
	vertical := geom.NewVec3(0, viewportHeight, 0)
	focalVec := geom.NewVec3(0, 0, focalLength)
	lowerLeft := origin.Minus(horizontal.Scale(0.5)).Minus(vertical.Scale(0.5)).Minus(focalVec)

	bpp := pitch / width // bytes-per-pixel
	for j := height - 1; j >= 0; j-- {
		for i := 0; i < width; i++ {
			ind := (j * pitch) + (i * bpp)

			u := float64(i) / float64(width-1)
			v := float64(j) / float64(height-1)

			r := geom.NewRay(
				origin,
				lowerLeft.Plus(horizontal.Scale(u)).Plus(vertical.Scale(v)).Minus(origin).Unit(),
			)
			pixelColor := color(r)
			img.WriteColor(ind, pixels, pixelColor)
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
