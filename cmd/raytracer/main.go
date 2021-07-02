package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
	"github.com/gabrielfvale/go-raytracer/pkg/obj"
	"github.com/gabrielfvale/go-raytracer/pkg/tracer"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {

	const aspect float64 = 16.0 / 9.0
	const width int = 640
	const height int = int(float64(width) / aspect)
	const samples int = 100
	fmt.Println(width, height)

	frame := tracer.NewFrame(width, height, aspect)
	objects := obj.NewList(
		obj.NewSphere(geom.NewVec3(0, 0, -1), 0.5),
		obj.NewSphere(geom.NewVec3(0, -100.5, -1), 100),
	)

	/* Begin SDL startup */
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
	/* End SDL startup */

	pixels, pitch, err := texture.Lock(nil)
	if err != nil {
		panic(err)
	}

	log.Printf("Started rendering")
	start := time.Now()

	frame.Render(pixels, pitch, objects, 8)

	elapsed := time.Since(start)
	log.Printf("Rendering took %s", elapsed)

	texture.Update(nil, pixels, pitch)
	texture.Unlock()

	renderer.Clear()
	renderer.CopyEx(texture, nil, nil, 0, nil, sdl.FLIP_VERTICAL)
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
