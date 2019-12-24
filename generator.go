package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"
)

// args is a struct that contains all the parameters passed through CL. Hence,
// it is an easy and organized way to make those variables accessible
type args struct {
	x1, x2, y1, y2, threshold float64
	w, h, nRoutines, nIter    int
	path                      string
}

// initialization of variables with default values or with arguments passed
// through CL. Return a populated args struct with everything set up
func argInit() args {

	var a args
	flag.Float64Var(&a.x1, "x1", -2.0, "left position of real axis")
	flag.Float64Var(&a.x2, "x2", 1.0, "right position of real axis")
	flag.Float64Var(&a.y1, "y1", -1.5, "down position of imaginary axis")
	flag.Float64Var(&a.y2, "y2", 1.5, "up position of imaginary axis")
	flag.Float64Var(&a.threshold, "th", 4.0, "squared threshold of the function")
	flag.IntVar(&a.w, "w", 1000, "width in pixels of the image")
	flag.IntVar(&a.h, "h", 1000, "height in pixels of the image")
	flag.IntVar(&a.nIter, "ni", 100, "maximum number of iterations for pixel")
	flag.IntVar(&a.nRoutines, "nr", 4, "number of go routines to be used")
	flag.StringVar(&a.path, "p", "./", "path to the generated png image")

	flag.Parse()
	return a
}

// receive a matrix corresponding to the number of iterations in each pixel and
// generate an image where each pixel 'brightness' is determined by the number
// of iterations in that point
func generateImage(a args, calc []int) {

	h := a.h
	w := a.w
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			index := (y * w) + x // 2D position into 1D
			brt := uint8((calc[index] * 0xff) / a.nIter)
			img.Set(x, y, color.RGBA{brt, brt, brt, 0xff})
		}
	}

	outputFile, err := os.Create(a.path + "mandelbrot.png")
	if err != nil {
		println("Error: could not save image")
		os.Exit(1)
	}
	png.Encode(outputFile, img)
	outputFile.Close()
}

// make calculations for each pixel, using the complex formula f(z) = z^2 + c.
// uses goroutines to distribute the work among cpu cores. Return an 1D array
// with the number of iterations in each pixel
func calculateSet(a args) []int {

	c := make([]int, a.w*a.h)
	rSize := (a.w * a.h) / a.nRoutines

	var wg sync.WaitGroup
	wg.Add(a.nRoutines)

	for r := 0; r < a.nRoutines; r++ {
		go func(id int, wg *sync.WaitGroup) {

			dx := (a.x2 - a.x1) / float64(a.w)
			dy := (a.y2 - a.y1) / float64(a.h)

			start := id * rSize
			end := (id + 1) * rSize
			if end > len(c) {
				end = len(c)
			}

			for i := start; i < end; i++ {
				// complex part of the equation
				cReal := (float64(i%a.w) * dx) + a.x1
				cImag := (float64(i/a.w) * dy) + a.y1
				zReal := 0.0
				zImag := 0.0
				iter := 0

				squared := func(x, y float64) float64 {
					return (x * x) + (y * y)
				}

				for squared(zReal, zImag) <= a.threshold && iter < a.nIter {
					newR := (zReal * zReal) - (zImag * zImag) + cReal
					newI := (2.0 * zReal * zImag) + cImag
					zReal = newR
					zImag = newI
					iter++
				}
				c[i] = iter
			}
			wg.Done()
		}(r, &wg)
	}
	wg.Wait()
	return c
}

func main() {
	a := argInit()
	c := calculateSet(a)
	generateImage(a, c)
}
