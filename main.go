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
	col "github.com/lucasb-eyer/go-colorful"
)

var (
	RGBColours = make([]color.RGBA, settings.MaxTypes)
	Images     = make([]*ebiten.Image, settings.MaxTypes)
	Attractors = make([]attract.AttractionFunction, settings.MaxTypes)
)

func init() {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		RecomputeImages(i)

		Attractors[i] = attract.DefaultAttractionFunc()
	}
}

func RecomputeImages(i int) {
	RecomputeColours()

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

func RecomputeColour(i int) {
	colour := col.Hsl(
		// equal spacing in hue based of Types
		360*float64(i)/float64(settings.Types),
		1,
		.7,
	)

	RGBColours[i] = color.RGBA{
		R: uint8(colour.R * 255),
		G: uint8(colour.G * 255),
		B: uint8(colour.B * 255),
		A: 255,
	}
}

func RecomputeColours() {
	for j := 0; j < settings.MaxTypes; j++ {
		RecomputeColour(j)
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

var clicks = map[string]int8{}
var presses = map[ebiten.Key]int8{}

var UI = map[[4]int][2]func(*Game){
	{settings.Width + 6, 4, settings.UIWidth - 10, 30}: {
		func(g *Game) {
			g.Setup()
			for i := 0; i < settings.MaxTypes; i++ {
				RecomputeImages(i)
			}
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
				RecomputeColours()
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
				RecomputeColours()
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

	{settings.Width + 8, 503, (settings.UIWidth - 16) / 2, 30}: {
		func(g *Game) {
			attract.AttractionMatrix = make([][]float64, settings.MaxTypes)
			for i := 0; i < settings.MaxTypes; i++ {
				attract.AttractionMatrix[i] = make([]float64, settings.MaxTypes)
			}
		}, func(g *Game) {},
	},
	{settings.Width + 4 + (settings.UIWidth)/2, 503, (settings.UIWidth - 16) / 2, 30}: {
		func(g *Game) {
			for i := range attract.AttractionMatrix {
				for j := range attract.AttractionMatrix[i] {
					attract.AttractionMatrix[i][j] = 2*rand.Float64() - 1
				}
			}
		}, func(g *Game) {},
	},
}

var Labels = map[[2]int]string{
	{settings.Width + 45, 10}:   "Random Environment",
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
	{settings.Width + 34, 510}:  "Clear",
	{settings.Width + 128, 510}: "Random",
}

type Game struct {
	particles       []Particle
	matrixEditorLoc [2]int
	darkTheme       bool
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

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			x, y := ebiten.CursorPosition()
			for i := range g.particles {
				g.particles[i].UpdateVelocity(Particle{
					X:    float64(x),
					Y:    float64(y),
					Type: g.particles[i].Type,
				}, attract.MouseAttraction)
			}
		}()
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
			presses[ebiten.KeyF11] = 1
		} else if ebiten.IsKeyPressed(ebiten.KeyUp) {
			presses[ebiten.KeyUp] = 1
		} else if ebiten.IsKeyPressed(ebiten.KeyDown) {
			presses[ebiten.KeyDown] = 1
		} else if ebiten.IsKeyPressed(ebiten.KeyLeft) {
			presses[ebiten.KeyLeft] = 1
		} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
			presses[ebiten.KeyRight] = 1
		} else if ebiten.IsKeyPressed(ebiten.KeyQ) {
			presses[ebiten.KeyQ] = 1
		} else if ebiten.IsKeyPressed(ebiten.KeyE) {
			presses[ebiten.KeyE] = 1
		} else if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			os.Exit(0)
		}

		for i, v := range presses {
			if v == 1 {
				if !ebiten.IsKeyPressed(i) {
					presses[i] = 0

					switch i {
					case ebiten.KeyUp:
						g.matrixEditorLoc[1]--
						g.matrixEditorLoc[1] = int(math.Max(0, float64(g.matrixEditorLoc[1])))
					case ebiten.KeyDown:
						g.matrixEditorLoc[1]++
						g.matrixEditorLoc[1] = int(math.Min(float64(settings.Types-1), float64(g.matrixEditorLoc[1])))
					case ebiten.KeyLeft:
						g.matrixEditorLoc[0]--
						g.matrixEditorLoc[0] = int(math.Max(0, float64(g.matrixEditorLoc[0])))
					case ebiten.KeyRight:
						g.matrixEditorLoc[0]++
						g.matrixEditorLoc[0] = int(math.Min(float64(settings.Types-1), float64(g.matrixEditorLoc[0])))
					case ebiten.KeyQ:
						attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] += 0.1
						if attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] > 1 {
							attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] = 1
						} else if attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] < -1 {
							attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] = -1
						}
					case ebiten.KeyE:
						attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] -= 0.1
						if attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] > 1 {
							attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] = 1
						} else if attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] < -1 {
							attract.AttractionMatrix[g.matrixEditorLoc[0]][g.matrixEditorLoc[1]] = -1
						}
					case ebiten.KeyF11:
						ebiten.SetFullscreen(!ebiten.IsFullscreen())
					}

					g.matrixEditorLoc = [2]int{
						int(math.Max(0, math.Min(float64(len(Attractors)-1), float64(g.matrixEditorLoc[0])))),
						int(math.Max(0, math.Min(float64(len(Attractors)-1), float64(g.matrixEditorLoc[1])))),
					}
				}
			}
		}
	}()

	wg.Wait()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Onclick Events
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

	if !g.darkTheme {
		screen.Fill(color.White)
	}

	// UI
	ebitenutil.DrawRect(screen, settings.Width+1, 0, 2, settings.Height, color.RGBA{100, 100, 100, 255})

	for i := range UI {
		ebitenutil.DrawRect(screen, float64(i[0]), float64(i[1]), float64(i[2]), float64(i[3]), color.RGBA{100, 100, 100, 255})
	}

	for k, v := range Labels {
		ebitenutil.DebugPrintAt(screen, v, k[0], k[1])
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %0.0f, TPS: %0.0f", ebiten.CurrentFPS(), ebiten.CurrentTPS()), settings.Width+12, settings.Height-22)

	for _, p := range g.particles {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)
		screen.DrawImage(Images[p.Type], op)
	}

	// Editor
	boxWidth := (settings.UIWidth - 40) / settings.Types
	for i := 0; i < settings.Types; i++ {
		// coloured headers
		ebitenutil.DrawRect(
			screen,
			float64(settings.Width+24+(i*boxWidth)), float64(314),
			float64(boxWidth), 10,
			RGBColours[i],
		)
		ebitenutil.DrawRect(
			screen,
			float64(settings.Width+8), float64(328+(i*boxWidth)),
			10, float64(boxWidth),
			RGBColours[i],
		)

		for j := 0; j < settings.Types; j++ {
			colour := color.RGBA{20, 20, 20, 255}
			if attract.AttractionMatrix[i][j] > 0 {
				colour = color.RGBA{0, uint8(255 * attract.AttractionMatrix[i][j]), 0, 255}
			} else if attract.AttractionMatrix[i][j] < 0 {
				colour = color.RGBA{uint8(255 * (1 - attract.AttractionMatrix[i][j])), 0, 0, 255}
			}

			ebitenutil.DrawRect(
				screen,
				float64(settings.Width+24+(i*boxWidth)), float64(328+(j*boxWidth)),
				float64(boxWidth), float64(boxWidth),
				colour,
			)
			if settings.Types < 7 {
				ebitenutil.DebugPrintAt(
					screen,
					fmt.Sprintf("%0.1f", attract.AttractionMatrix[i][j]),
					settings.Width+24+(i*boxWidth)+boxWidth/2-10,
					328+(j*boxWidth)+boxWidth/2-5,
				)
			}
		}
	}

	// add small white border to editor selection (editorLoc)
	ebitenutil.DrawRect(
		screen,
		float64(settings.Width+24+(g.matrixEditorLoc[0]*boxWidth)),
		float64(328+(g.matrixEditorLoc[1]*boxWidth)),
		float64(boxWidth), 4,
		color.RGBA{255, 255, 255, 255},
	)

	ebitenutil.DebugPrintAt(
		screen,
		"Esc: Exit\nF11: Toggle Fullscreen\nSome settings need a new\nenvironment before they update.\nSome update live.\nArrow keys to move editor\nselection. Q and E to change\nvalues. Click to interact.",
		settings.Width+6, settings.Height-200,
	)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return settings.Width + settings.UIWidth, settings.Height
}

func (g *Game) Setup() {
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
	case "point":
		for i := 0; i < settings.NParticles; i++ {
			g.particles = append(g.particles, Particle{
				X:    settings.Width/2 + rand.Float64() - .5,
				Y:    settings.Height/2 + rand.Float64() - .5,
				Type: int8(rand.Intn(settings.Types)),
			})
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
	ebiten.SetMaxTPS(250)

	game := Game{darkTheme: true}
	game.Setup()

	if err := ebiten.RunGame(&game); err != nil {
		panic(err)
	}
}
