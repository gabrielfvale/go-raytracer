package main

import (
	"flag"
	"fmt"
	"unsafe"

	"github.com/gabrielfvale/go-raytracer/pkg/geom"
	"github.com/gabrielfvale/go-raytracer/pkg/tracer"
	"github.com/gabrielfvale/go-raytracer/pkg/util"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {

	var width int
	var samples int
	var output string

	flag.IntVar(&width, "w", 640, "")
	flag.IntVar(&samples, "s", 8, "")
	flag.StringVar(&output, "o", "", "")
	flag.Parse()

	aspect := 1.0
	height := int(float64(width) / aspect)

	// fmt.Println(width, height)

	matRed := tracer.DiffuseMaterial(tracer.NewColor(0.65, 0.05, 0.05))
	matGreen := tracer.DiffuseMaterial(tracer.NewColor(0.12, 0.45, 0.15))
	matWhite := tracer.DiffuseMaterial(tracer.NewColor(0.73, 0.73, 0.73))
	matLight := tracer.LightMaterial(tracer.NewColor(0.8, 0.8, 0.8), 1)
	matGlass := tracer.DielectricMaterial(1.53)
	matMirror := tracer.MetalicMaterial(tracer.NewColor(1, 1, 1), 1, 0)
	// matNormal := tracer.NormalMaterial()

	objects := []tracer.Hitable{
		tracer.NewAABB(geom.NewVec3(213, 548, 227), geom.NewVec3(343, 548.1, 332), matLight),
		tracer.NewAABB(geom.NewVec3(0, 0, 0), geom.NewVec3(555, 0.1, 555), matWhite),     // floor
		tracer.NewAABB(geom.NewVec3(0, 555, 0), geom.NewVec3(555, 555.1, 555), matWhite), // ceiling
		tracer.NewAABB(geom.NewVec3(0, 0, 555), geom.NewVec3(555, 555, 555.1), matWhite), // back wall
		tracer.NewAABB(geom.NewVec3(555, 0, 0), geom.NewVec3(555.1, 555, 555), matRed),   // left wall
		tracer.NewAABB(geom.NewVec3(0, 0, 0), geom.NewVec3(0.1, 555, 555), matGreen),     // right wall
		tracer.NewSphere(geom.NewVec3(278+110, 90, 227+120), 90, matMirror),
		tracer.NewSphere(geom.NewVec3(278-110, 90, 227-40), 90, matGlass),
	}

	cam := tracer.NewCamera(
		geom.NewVec3(278, 273, -800),
		geom.NewVec3(278, 278, 1),
		geom.NewVec3(0, 1, 0),
		40, aspect)

	globalMap := tracer.NewPhotonMap(50000)
	causticsMap := tracer.NewPhotonMap(50000)
	scene := tracer.NewScene(width, height, cam, objects, &globalMap, &causticsMap)

	if output != "" { // render to image
		bpp := int(unsafe.Sizeof(uint32(0)))
		pitch := bpp * height
		pixels := make([]uint8, width*pitch)
		scene.Render(pixels, pitch, samples)
		util.SaveToImage(output, width, height, pixels)
		return
	}

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

	scene.Render(pixels, pitch, samples)

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
