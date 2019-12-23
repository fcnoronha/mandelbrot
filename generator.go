package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"
)

// used to pass input parameters to functions
type Args struct {
	x1, x2, y1, y2, threshold float64
	w, h, nRoutines, nIter    int
	path                      string
}

func iterate(a Args, cReal, cImag float64) int {

	var zReal float64 = 0.0
	var zImag float64 = 0.0
	var iter int = 0

	squared := func(x, y float64) float64 {
		return (x * x) + (y * y)
	}

	for squared(zReal, zImag) <= a.threshold && iter < a.nIter {

		var nr float64 = (zReal * zReal) - (zImag * zImag) + cReal
		var ni float64 = (2.0 * zReal * zImag) + cImag
		zReal = nr
		zImag = ni
		iter++
	}
	return iter
}

func argInit() Args {

	var a Args
	flag.Float64Var(&a.x1, "x1", -2.0, "left position of real axis")
	flag.Float64Var(&a.x2, "x2", 1.0, "right position of real axis")
	flag.Float64Var(&a.y1, "y1", -1.5, "down position of imaginary axis")
	flag.Float64Var(&a.y2, "y2", 1.5, "up position of imaginary axis")
	flag.Float64Var(&a.threshold, "th", 4.0, "threshold of the function")

	flag.IntVar(&a.w, "w", 1000, "width in pixels of the image")
	flag.IntVar(&a.h, "h", 1000, "height in pixels of the image")
	flag.IntVar(&a.nIter, "ni", 1000, "maximum number of iterations for pixel")
	flag.IntVar(&a.nRoutines, "nr", 4, "number of go routines to be used")

	flag.StringVar(&a.path, "-p", "./", "path to the generated png image")

	flag.Parse()
	return a
}

func regionIterate(id int, a Args, c *[]int, rSize int, wg *sync.WaitGroup) {

	dx := (a.x2 - a.x1) / float64(a.w)
	dy := (a.y2 - a.y1) / float64(a.h)

	var start int = id * rSize
	var end int = (id + 1) * rSize
	if len(*c) < end {
		end = len(*c)
	}

	for i := start; i < end; i++ {

		var x float64 = (float64(i%a.w) * dx) + a.x1
		var y float64 = (float64(i/a.w) * dy) + a.y1
		(*c)[i] = iterate(a, x, y)
	}
	wg.Done()
}

func makeIterations(a Args) []int {

	calc := make([]int, a.w*a.h)
	rSize := (a.w * a.h) / a.nRoutines
	var wg sync.WaitGroup
	for r := 0; r < a.nRoutines; r++ {
		wg.Add(1)
		go regionIterate(r, a, &calc, rSize, &wg)
	}
	wg.Wait()
	return calc
}

func generateImage(a Args, calc []int) {

	img := image.NewRGBA(image.Rect(0, 0, a.w, a.h))

	for y := 0; y < a.h; y++ {
		for x := 0; x < a.w; x++ {
			col := uint8((float64(calc[y*a.w+x]) / float64(a.nIter)) * 0xff)
			img.Set(x, y, color.RGBA{col, col, col, 0xff})
		}
	}

	outputFile, err := os.Create(a.path + "mandelbrot.png")
	if err != nil {
		println("Could not save image")
		os.Exit(1)
	}

	png.Encode(outputFile, img)
	outputFile.Close()
}

func main() {

	args := argInit()
	calc := makeIterations(args)
	generateImage(args, calc)
}
