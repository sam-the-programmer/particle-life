package settings

import "math/rand"

const (
	// World Settings
	ARRANGEMENT = "random"
	N_PARTICLES = 300
	SCALE       = 1
	WIDTH       = 500
	HEIGHT      = 500

	// Particle Settings
	TYPES         = 10
	SPEED         = .001
	PARTICLE_SIZE = 3
	REPELRADIUS   = 0.5
	REPELSTRENGTH = 2

	ATTRACTION = "cluster"
	RADII      = "equal"

	MINRADIUS = 100
	MAXRADIUS = 200

	FRICTION = .999
)

var (
	RANDOMFUNC = rand.NormFloat64
)
