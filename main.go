package main

import (
	"fmt"
	"image/color"
	"life/attract"
	"life/settings"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

var (
	RGBColours = make([]color.RGBA, settings.MaxTypes)
	Images     = make([]*ebiten.Image, settings.MaxTypes)
	Attractors = make([]attract.AttractionFunction, settings.MaxTypes)
)

func init() {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		RGBColours[i] = color.RGBA{
			R: uint8(255 * math.Abs(math.Sin(float64(i)*2*math.Pi/float64(settings.Types)))),
			G: uint8(255 * math.Abs(math.Sin(float64(i)*2*math.Pi/float64(settings.Types)+2*math.Pi/3))),
			B: uint8(255 * math.Abs(math.Sin(float64(i)*2*math.Pi/float64(settings.Types)+4*math.Pi/3))),
			A: 255,
		}

		RecomputeImages(i)

		Attractors[i] = attract.DefaultAttractionFunc()
	}
}

func RecomputeImages(i int) {
	var err error
	Images[i], err = ebiten.NewImage(settings.ParticleSize, settings.ParticleSize, ebiten.FilterLinear)
	if err != nil {
		log.Fatal(err)
	}
	err = Images[i].Fill(RGBColours[i])
	if err != nil {
		log.Fatal(err)
	}
}

func dist(x1, y1, x2, y2 float64) float64 {
	// return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
	// manhattan distance is faster
	return math.Abs(x2-x1) + math.Abs(y2-y1)
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
		p.X = settings.Width
	}
	if p.X > settings.Width {
		p.X = 0
	}

	if p.Y < 0 {
		p.Y = settings.Height
	}
	if p.Y > settings.Height {
		p.Y = 0
	}

	// Update position
	p.X += p.Velocity[0] * settings.Speed
	p.Y += p.Velocity[1] * settings.Speed

	// Friction
	p.Velocity[0] *= settings.Friction
	p.Velocity[1] *= settings.Friction
}

type Game struct {
	particles []Particle
}

var clicks = map[string]int8{
	"types++": 0,
	"types--": 0,
}

var UI = map[[4]int][2]func(*Game){
	{settings.Width + 6, 4, settings.UIWidth - 10, 30}: {
		func(g *Game) {
			g.Setup()
		},
		func(g *Game) {},
	},

	// Change Types
	{settings.Width + 104, 36, 44, 30}: {
		func(g *Game) {
			clicks["types++"] = 1
		},
		func(g *Game) {
			if clicks["types++"] == 1 {
				settings.Types++
				settings.Types = int(math.Min(float64(settings.Types), float64(settings.MaxTypes)))
				clicks["types++"] = 0
				Labels[[2]int{settings.Width + 8, 42}] = fmt.Sprintf("Types: %d", settings.Types)
			}
		},
	},
	{settings.Width + 151, 36, 44, 30}: {
		func(g *Game) {
			clicks["types--"] = 1
		},
		func(g *Game) {
			if clicks["types--"] == 1 {
				clicks["types--"] = 0
				settings.Types--
				settings.Types = int(math.Max(float64(settings.Types), 1))
				Labels[[2]int{settings.Width + 8, 42}] = fmt.Sprintf("Types: %d", settings.Types)
			}
		},
	},

	// Change Number of Particles
	{settings.Width + 104, 76, 44, 30}: {
		func(g *Game) {
			clicks["particles++"] = 1
		},
		func(g *Game) {
			if clicks["particles++"] == 1 {
				settings.NParticles += 100
				settings.NParticles = int(math.Min(float64(settings.NParticles), float64(settings.MaxParticles)))
				clicks["particles++"] = 0
				Labels[[2]int{settings.Width + 8, 82}] = fmt.Sprintf("Particles: %d", settings.NParticles)
			}
		},
	},
	{settings.Width + 151, 76, 44, 30}: {
		func(g *Game) {
			clicks["particles--"] = 1
		},
		func(g *Game) {
			if clicks["particles--"] == 1 {
				clicks["particles--"] = 0
				settings.NParticles -= 100
				settings.NParticles = int(math.Max(float64(settings.NParticles), 1))
				Labels[[2]int{settings.Width + 8, 82}] = fmt.Sprintf("Particles: %d", settings.NParticles)
			}
		},
	},

	// Change Speed in increments of 0.01
	{settings.Width + 104, 116, 44, 30}: {
		func(g *Game) {
			clicks["speed++"] = 1
		},
		func(g *Game) {
			if clicks["speed++"] == 1 {
				settings.Speed += 0.01
				clicks["speed++"] = 0
				Labels[[2]int{settings.Width + 8, 122}] = fmt.Sprintf("Speed: %.2f", settings.Speed)
			}
		},
	},
	{settings.Width + 151, 116, 44, 30}: {
		func(g *Game) {
			clicks["speed--"] = 1
		},
		func(g *Game) {
			if clicks["speed--"] == 1 {
				clicks["speed--"] = 0
				settings.Speed -= 0.01
				settings.Speed = math.Max(settings.Speed, 0)
				Labels[[2]int{settings.Width + 8, 122}] = fmt.Sprintf("Speed: %.2f", settings.Speed)
			}
		},
	},

	// Change Friction in increments of 0.01
	{settings.Width + 104, 156, 44, 30}: {
		func(g *Game) {
			clicks["friction++"] = 1
		},
		func(g *Game) {
			if clicks["friction++"] == 1 {
				settings.Friction += 0.01
				settings.Friction = math.Min(settings.Friction, 1)
				clicks["friction++"] = 0
				Labels[[2]int{settings.Width + 8, 162}] = fmt.Sprintf("Friction: %.2f", settings.Friction)
			}
		},
	},
	{settings.Width + 151, 156, 44, 30}: {
		func(g *Game) {
			clicks["friction--"] = 1
		},
		func(g *Game) {
			if clicks["friction--"] == 1 {
				clicks["friction--"] = 0
				settings.Friction -= 0.01
				settings.Friction = math.Max(settings.Friction, 0)
				Labels[[2]int{settings.Width + 8, 162}] = fmt.Sprintf("Friction: %.2f", settings.Friction)
			}
		},
	},

	// RepelStength, in incrememnts of 0.1
	{settings.Width + 104, 196, 44, 30}: {
		func(g *Game) {
			clicks["repel++"] = 1
		},
		func(g *Game) {
			if clicks["repel++"] == 1 {
				settings.RepelStrength += 0.1
				clicks["repel++"] = 0
				Labels[[2]int{settings.Width + 8, 202}] = fmt.Sprintf("Repel: %.2f", settings.RepelStrength)
			}
		},
	},
	{settings.Width + 151, 196, 44, 30}: {
		func(g *Game) {
			clicks["repel--"] = 1
		},
		func(g *Game) {
			if clicks["repel--"] == 1 {
				clicks["repel--"] = 0
				settings.RepelStrength -= 0.1
				settings.RepelStrength = math.Max(settings.RepelStrength, 0)
				Labels[[2]int{settings.Width + 8, 202}] = fmt.Sprintf("Repel: %.2f", settings.RepelStrength)
			}
		},
	},

	// RepelRadius, in increments of 1
	{settings.Width + 104, 236, 44, 30}: {
		func(g *Game) {
			clicks["radius++"] = 1
		},
		func(g *Game) {
			if clicks["radius++"] == 1 {
				settings.RepelRadius += 1
				clicks["radius++"] = 0
				Labels[[2]int{settings.Width + 8, 242}] = fmt.Sprintf("Radius: %.2f", settings.RepelRadius)
			}
		},
	},
	{settings.Width + 151, 236, 44, 30}: {
		func(g *Game) {
			clicks["radius--"] = 1
		},
		func(g *Game) {
			if clicks["radius--"] == 1 {
				clicks["radius--"] = 0
				settings.RepelRadius -= 1
				settings.RepelRadius = math.Max(settings.RepelRadius, 0)
				Labels[[2]int{settings.Width + 8, 242}] = fmt.Sprintf("Radius: %.2f", settings.RepelRadius)
			}
		},
	},

	// Change Particle Size in increments of 1
	{settings.Width + 104, 276, 44, 30}: {
		func(g *Game) {
			clicks["size++"] = 1
		},
		func(g *Game) {
			if clicks["size++"] == 1 {
				settings.ParticleSize += 1
				clicks["size++"] = 0
				Labels[[2]int{settings.Width + 8, 282}] = fmt.Sprintf("Size: %d", settings.ParticleSize)
				for i := 0; i < 100; i++ {
					RecomputeImages(i)
				}
			}
		},
	},
	{settings.Width + 151, 276, 44, 30}: {
		func(g *Game) {
			clicks["size--"] = 1
		},
		func(g *Game) {
			if clicks["size--"] == 1 {
				clicks["size--"] = 0
				settings.ParticleSize -= 1
				settings.ParticleSize = int(math.Max(float64(settings.ParticleSize), 0))
				Labels[[2]int{settings.Width + 8, 282}] = fmt.Sprintf("Size: %d", settings.ParticleSize)
				for i := 0; i < 100; i++ {
					RecomputeImages(i)
				}
			}
		},
	},
}

var Labels = map[[2]int]string{
	{settings.Width + 50, 10}:   "Reset Environment",
	{settings.Width + 8, 42}:    fmt.Sprintf("Types: %d", settings.Types),
	{settings.Width + 120, 42}:  "+",
	{settings.Width + 170, 42}:  "-",
	{settings.Width + 8, 82}:    fmt.Sprintf("Particles: %d", settings.NParticles),
	{settings.Width + 120, 82}:  "+",
	{settings.Width + 170, 82}:  "-",
	{settings.Width + 8, 122}:   fmt.Sprintf("Speed: %.2f", settings.Speed),
	{settings.Width + 120, 122}: "+",
	{settings.Width + 170, 122}: "-",
	{settings.Width + 8, 162}:   fmt.Sprintf("Friction: %.2f", settings.Friction),
	{settings.Width + 120, 162}: "+",
	{settings.Width + 170, 162}: "-",
	{settings.Width + 8, 202}:   fmt.Sprintf("Repel: %.2f", settings.RepelStrength),
	{settings.Width + 120, 202}: "+",
	{settings.Width + 170, 202}: "-",
	{settings.Width + 8, 242}:   fmt.Sprintf("Radius: %.2f", settings.RepelRadius),
	{settings.Width + 120, 242}: "+",
	{settings.Width + 170, 242}: "-",
	{settings.Width + 8, 282}:   fmt.Sprintf("Size: %d", settings.ParticleSize),
	{settings.Width + 120, 282}: "+",
	{settings.Width + 170, 282}: "-",
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

				// Allow for overflow to other side of screen
				g.particles[i].UpdateVelocity(Particle{
					X:    g.particles[j].X + settings.Width,
					Y:    g.particles[j].Y,
					Type: g.particles[j].Type,
				}, Attractors[g.particles[i].Type])
				g.particles[i].UpdateVelocity(Particle{
					X:    g.particles[j].X - settings.Width,
					Y:    g.particles[j].Y,
					Type: g.particles[j].Type,
				}, Attractors[g.particles[i].Type])
				g.particles[i].UpdateVelocity(Particle{
					X:    g.particles[j].X,
					Y:    g.particles[j].Y + settings.Height,
					Type: g.particles[j].Type,
				}, Attractors[g.particles[i].Type])
				g.particles[i].UpdateVelocity(Particle{
					X:    g.particles[j].X,
					Y:    g.particles[j].Y - settings.Height,
					Type: g.particles[j].Type,
				}, Attractors[g.particles[i].Type])

			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := range g.particles {
			g.particles[i].UpdatePosition()
		}
	}()

	go func() {
		defer wg.Done()
		if ebiten.IsKeyPressed(ebiten.KeyF11) {
			ebiten.SetFullscreen(true)
		} else if ebiten.IsKeyPressed(ebiten.KeyF12) {
			ebiten.SetFullscreen(false)
		} else if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			os.Exit(0)
		}
	}()

	wg.Wait()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Editor UI
	ebitenutil.DrawRect(screen, settings.Width+1, 0, 2, settings.Height, color.RGBA{100, 100, 100, 255})

	for i := range UI {
		ebitenutil.DrawRect(screen, float64(i[0]), float64(i[1]), float64(i[2]), float64(i[3]), color.RGBA{100, 100, 100, 255})
	}

	for k, v := range Labels {
		ebitenutil.DebugPrintAt(screen, v, k[0], k[1])
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %0.0f, TPS: %0.0f", ebiten.CurrentFPS(), ebiten.CurrentTPS()), settings.Width+12, settings.Height-22)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { // onclick events for ui
		defer wg.Done()
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			for k, v := range UI {
				if x >= k[0] && x <= k[0]+k[2] && y >= k[1] && y <= k[1]+k[3] {
					v[0](g)
				}
			}
		} else {
			for k := range UI {
				UI[k][1](g)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for _, p := range g.particles {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.X, p.Y)
			screen.DrawImage(Images[p.Type], op)
		}
	}()
	wg.Wait()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return settings.Width + settings.UIWidth, settings.Height
}

func (g *Game) Setup() {
	attract.RandomizeAttractionMatrix()
	g.particles = make([]Particle, 0, settings.NParticles)
	switch settings.Arrangement {
	case "random":
		for i := 0; i < settings.NParticles; i++ {
			g.particles = append(g.particles, Particle{
				X:    rand.Float64() * settings.Width,
				Y:    rand.Float64() * settings.Height,
				Type: int8(rand.Intn(settings.Types)),
			})
		}
	case "circle":
		for i := 0; i < settings.NParticles; i++ {
			angle := float64(i) * 2 * math.Pi / float64(settings.NParticles)
			g.particles = append(g.particles, Particle{
				X:    settings.Width/2 + math.Cos(angle)*settings.Width/2,
				Y:    settings.Height/2 + math.Sin(angle)*settings.Height/2,
				Type: int8(rand.Intn(settings.Types)),
			})
		}
	case "f_circle": // filled circle
		for i := 0; i < settings.NParticles; i++ {
			angle := float64(i) * 2 * math.Pi / float64(settings.NParticles)
			g.particles = append(g.particles, Particle{
				X:    settings.Width/2 + math.Cos(angle)*settings.Width/2 + 20*(rand.Float64()-.5),
				Y:    settings.Height/2 + math.Sin(angle)*settings.Height/2 + 20*(rand.Float64()-.5),
				Type: int8(rand.Intn(settings.Types)),
			})
		}
	case "concentric":
		for ring := 0; ring < settings.Types; ring++ {
			for i := 0; i < settings.NParticles/settings.Types; i++ {
				angle := float64(i) * 2 * math.Pi / float64(settings.NParticles/settings.Types)
				g.particles = append(g.particles, Particle{
					X:    float64(settings.Width/2+math.Cos(angle)*settings.Width/2*float64(ring)/float64(settings.Types)) + rand.Float64() - .5,
					Y:    float64(settings.Height/2+math.Sin(angle)*settings.Height/2*float64(ring)/float64(settings.Types)) + rand.Float64() - .5,
					Type: int8(ring),
				})
			}
		}
	case "line":
		for i := 0; i < settings.NParticles; i++ {
			g.particles = append(g.particles, Particle{
				X:    float64(i) * settings.Width / float64(settings.NParticles),
				Y:    settings.Height/2 + rand.Float64() - .5,
				Type: int8(rand.Intn(settings.Types)),
			})
		}
	case "grid":
		for x := 0; x < int(math.Sqrt(float64(settings.NParticles))); x++ {
			for y := 0; y < int(math.Sqrt(float64(settings.NParticles))); y++ {
				g.particles = append(g.particles, Particle{
					X:    float64(x)*settings.Width/math.Sqrt(float64(settings.NParticles)) + rand.Float64() - .5,
					Y:    float64(y)*settings.Height/math.Sqrt(float64(settings.NParticles)) + rand.Float64() - .5,
					Type: int8(rand.Intn(settings.Types)),
				})
			}
		}
	case "row":
		for t := 0; t < settings.Types; t++ {
			for i := 0; i < settings.NParticles/settings.Types; i++ {
				g.particles = append(g.particles, Particle{
					X:    (settings.Width/float64(settings.Types))*(float64(t)+rand.Float64()) - 20,
					Y:    settings.Height/2 + 20*(rand.Float64()-.5),
					Type: int8(t),
				})
			}
		}
	default:
		log.Fatal("Unknown arrangement.")
	}
}

func main() {
	ebiten.SetWindowSize(settings.Width*settings.Scale+settings.UIWidth, settings.Height*settings.Scale)
	ebiten.SetWindowTitle("Particle Life")
	ebiten.SetWindowResizable(true)
	ebiten.SetInitFocused(true)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetMaxTPS(250)

	game := Game{}
	game.Setup()

	if err := ebiten.RunGame(&game); err != nil {
		panic(err)
	}
}
