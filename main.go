package main

import (
	"math"
	"math/rand/v2"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowWidth  = 800
	windowHeight = 600
	numStars     = 300
	centerX      = windowWidth / 2
	centerY      = windowHeight / 2
)

type position struct {
	x, y float64
}

// a struct to represent a star, with its position, speed and brightness
type Star struct {
	pos        position
	speed      position
	brightness byte
}

// support struct to keep track of all stars and update/draw them
type Stars struct {
	stars         []Star
	minWarpFactor float64
}

// generate a random float64 between min and max using a PCG random generator
func randFloat64(min, max float64) float64 {
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixMicro())))
	return min + (max-min)*r.Float64()
}

// clear the pixel buffer by setting all values to 0 (black)
func clearPixels(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

// set a pixel in the pixel buffer to a specific color (in this case, it's black and white so set everything to the same brightness	value)
// the buffer is a linear array of bytes, where each pixel is represented by 4 bytes (A, B, G, R), so we need to calculate the index of the pixel we want to set
// by multiplying the y coordinate by the width of the window and adding the x coordinate, then multiplying by 4 to get the index of the first byte of the pixel (A), and then setting the R, G and B values to the same brightness value. Note that
func setPixel(pixels []byte, x, y int, c byte) {
	index := (y*windowWidth + x) * 4
	if index > 0 && index < len(pixels)-4 {
		pixels[index] = c   // R
		pixels[index+1] = c // G
		pixels[index+2] = c // B
	}
}

func newStar() Star {
	angle := randFloat64(-math.Pi, math.Pi)
	speed := 255 * math.Pow(randFloat64(float64(0.3), float64(1.0)), 2)

	// calculate the direction of the star based on the angle
	dx := math.Cos(angle)
	dy := math.Sin(angle)

	d := rand.IntN(int(math.Round(float64(windowWidth)/8))) + 1

	// and then calculate the initial position of the star based on the center of the window and the direction
	pos := position{centerX + dx*float64(d), centerY + dy*float64(d)}
	// finally calculate the speed of the star based on the direction and the speed value we generated earlier
	speedPos := position{dx * speed, dy * speed}

	star := Star{
		pos:        pos,
		speed:      speedPos,
		brightness: 0,
	}
	return star
}

func (s *Stars) update(elapsed float32) {
	// for each star
	for i := range s.stars {
		star := &s.stars[i]
		// update the position of the star based on its speed and the elapsed time since the last update, multiplied by a warp factor to make the stars move faster
		star.pos.x += star.speed.x * s.minWarpFactor
		star.pos.y += star.speed.y * s.minWarpFactor

		// when a star goes out of the window bounds, we reset it to a new random position and speed, and set its brightness back to 0.
		if star.pos.x < 0 || star.pos.x >= windowWidth || star.pos.y < 0 || star.pos.y >= windowHeight {
			s.stars[i] = newStar()
		} else {
			// Otherwise, if the star is still within the bounds, we increase its brightness gradually until it reaches 255 (fully bright).
			if star.brightness < 255 {
				star.brightness += byte(5)
			}
		}
		// assign to write the star to the star tracker
		s.stars[i] = *star
	}
}

func (s *Stars) draw(pixels []byte) {
	// for every star
	for i := range s.stars {
		star := s.stars[i]
		// set the pixel in the pixel buffer using the current status of the star (position and brightness)
		setPixel(pixels, int(star.pos.x), int(star.pos.y), star.brightness)
	}
}

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}

	defer sdl.Quit()

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")
	window, err := sdl.CreateWindow("Starfield Simulation", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, windowWidth, windowHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)

	if err != nil {
		panic(err)
	}

	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, windowWidth, windowHeight)
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	var elapsed float32

	// initialize the pixel buffer, acting as the framebuffer we are going to write to
	pixels := make([]byte, windowWidth*windowHeight*4)
	starField := make([]Star, numStars)
	all := &Stars{minWarpFactor: 0.05}

	// populate the initial star field
	for range starField {
		all.stars = append(all.stars, newStar())
	}

	// initialize the star field by updating it a few times to give the stars some initial speed and brightness before we start rendering
	// skipping this step will result in a star field that starts with all stars black and getting brighter over the first few frames, which is weird but artistically interesting
	for range 2000 {
		all.update(0)
	}

	// initialize the pixel buffer to black before we start the main loop
	clearPixels(pixels)

	for {
		frameStart := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.KeyboardEvent:
				if e.Keysym.Sym == sdl.K_ESCAPE && e.Type == sdl.KEYDOWN {
					//fmt.Printf("Pressed %+v\n", e)
					return
				}
				if e.Keysym.Sym == sdl.K_UP && e.Type == sdl.KEYDOWN {
					all.minWarpFactor += 0.01
				}
				if e.Keysym.Sym == sdl.K_DOWN && e.Type == sdl.KEYDOWN {
					all.minWarpFactor -= 0.01
				}
			}
		}
		all.update(elapsed)
		all.draw(pixels)
		texture.Update(nil, unsafe.Pointer(&pixels[0]), windowWidth*4)
		renderer.Copy(texture, nil, nil)
		renderer.Present()
		clearPixels(pixels)
		elapsed = float32(time.Since(frameStart).Seconds() * 1000)
		if elapsed < 16 {
			sdl.Delay(uint32(16 - elapsed))
			elapsed = float32(time.Since(frameStart).Seconds() * 1000)
		}
	}
}
