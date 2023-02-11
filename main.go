package main

import (
	"image/color"
	"life/attract"
	"life/settings"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten"
)

var (
	RGBColours = [settings.TYPES]color.RGBA{}
	Images     = [settings.TYPES]*ebiten.Image{}
	Attractors = [settings.TYPES]attract.AttractionFunction{}
)

func init() {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < settings.TYPES; i++ {
		// use a colour wheel (thanks copilot)
		RGBColours[i] = color.RGBA{
			R: uint8(255 * math.Abs(math.Sin(float64(i)*2*math.Pi/float64(settings.TYPES)))),
			G: uint8(255 * math.Abs(math.Sin(float64(i)*2*math.Pi/float64(settings.TYPES)+2*math.Pi/3))),
			B: uint8(255 * math.Abs(math.Sin(float64(i)*2*math.Pi/float64(settings.TYPES)+4*math.Pi/3))),
			A: 255,
		}

		var err error
		Images[i], err = ebiten.NewImage(settings.PARTICLE_SIZE, settings.PARTICLE_SIZE, ebiten.FilterNearest)
		if err != nil {
			log.Fatal(err)
		}
		Images[i].Fill(RGBColours[i])

		Attractors[i] = attract.DefaultAttractionFunc()
	}
}

func dist(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

type Particle struct {
	X, Y     float64
	Velocity [2]float64
	Type     int8
}

func (p *Particle) UpdateVelocity(other Particle, attract attract.AttractionFunction) {
	d := dist(p.X, p.Y, other.X, other.Y)
	p.Velocity[0] += (other.X - p.X) / d * attract(d, p.Type, other.Type)
	p.Velocity[1] += (other.Y - p.Y) / d * attract(d, p.Type, other.Type)
}

func (p *Particle) UpdatePosition() {
	// Teleport to other side of screen if out of bounds
	if p.X < 0 {
		p.X = settings.WIDTH
	}
	if p.X > settings.WIDTH {
		p.X = 0
	}

	if p.Y < 0 {
		p.Y = settings.HEIGHT
	}
	if p.Y > settings.HEIGHT {
		p.Y = 0
	}

	// Update position
	p.X += p.Velocity[0] * settings.SPEED
	p.Y += p.Velocity[1] * settings.SPEED

	// Friction
	p.Velocity[0] *= settings.FRICTION
	p.Velocity[1] *= settings.FRICTION
}

type Game struct {
	particles []Particle
}

func (g *Game) Update(screen *ebiten.Image) error {
	var wg sync.WaitGroup
	for i := range g.particles {
		wg.Add(1)
		go func(i int) {
			for j := range g.particles {
				if i == j {
					continue
				}
				g.particles[i].UpdateVelocity(g.particles[j], Attractors[g.particles[i].Type])
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := range g.particles {
		g.particles[i].UpdatePosition()
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		x = int(float64(x) / settings.SCALE)
		x = int(float64(y) / settings.SCALE)
		for i := range g.particles {
			g.particles[i].UpdateVelocity(Particle{X: float64(x), Y: float64(y)}, Attractors[g.particles[i].Type])
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, p := range g.particles {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)
		screen.DrawImage(Images[p.Type], op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return settings.WIDTH, settings.HEIGHT
}

func (g *Game) Setup() {
	switch settings.ARRANGEMENT {
	case "random":
		for i := 0; i < settings.N_PARTICLES; i++ {
			g.particles = append(g.particles, Particle{
				X:    rand.Float64() * settings.WIDTH,
				Y:    rand.Float64() * settings.HEIGHT,
				Type: int8(rand.Intn(settings.TYPES)),
			})
		}
	default:
		log.Fatal("Unknown arrangement.")
	}
}

func main() {
	ebiten.SetWindowSize(settings.WIDTH*settings.SCALE, settings.HEIGHT*settings.SCALE)
	ebiten.SetWindowTitle("Particle Life")
	ebiten.SetMaxTPS(300)

	game := Game{}
	game.Setup()

	if err := ebiten.RunGame(&game); err != nil {
		panic(err)
	}
}
