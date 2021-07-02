package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
	"github.com/gabrielfvale/go-raytracer/pkg/tracer"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {

	const aspect float64 = 16.0 / 9.0
	const width int = 640
	const height int = int(float64(width) / aspect)
	var samples int
	fmt.Println(width, height)
	flag.IntVar(&samples, "s", 8, "")
	flag.Parse()

	frame := tracer.NewFrame(width, height, aspect)

	matGround := tracer.NewLambert(tracer.NewColor(0.8, 0.8, 0.0))
	matCenter := tracer.NewLambert(tracer.NewColor(0.7, 0.3, 0.3))
	matLeft := tracer.NewMetal(tracer.NewColor(0.8, 0.8, 0.8), 0.3)
	matRight := tracer.NewMetal(tracer.NewColor(0.8, 0.6, 0.2), 1.0)

	objects := tracer.NewList(
		tracer.NewSphere(geom.NewVec3(0.0, -100.5, -1.0), 100, matGround),
		tracer.NewSphere(geom.NewVec3(0.0, 0.0, -1.0), 0.5, matCenter),
		tracer.NewSphere(geom.NewVec3(-1.0, 0.0, -1.0), 0.5, matLeft),
		tracer.NewSphere(geom.NewVec3(1.0, 0.0, -1.0), 0.5, matRight),
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

	log.Printf("Started rendering (%d samples)", samples)
	start := time.Now()

	frame.Render(pixels, pitch, objects, samples)

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
