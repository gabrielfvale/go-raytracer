package main

import "fmt"

func main() {
	imageWidth, imageHeight := 256, 256

	fmt.Println("P3")
	fmt.Println(imageWidth, imageHeight)
	fmt.Println(255)

	for j := imageHeight - 1; j >= 0; j-- {
		for i := 0; i < imageWidth; i++ {
			r := float64(i) / float64(imageWidth-1)
			g := float64(j) / float64(imageHeight-1)
			b := 0.25

			ir := int(255.99 * r)
			ig := int(255.99 * g)
			ib := int(255.99 * b)
			fmt.Println(ir, ig, ib)
		}
	}
}
