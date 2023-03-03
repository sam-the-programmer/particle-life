package settings

import (
	"math/rand"
)

const (
	// World Settings
	Arrangement  = "random"
	Scale        = 1.
	Width        = 1200
	Height       = 800
	MaxTypes     = 100
	MaxParticles = 10000

	// Optional Attraction Settings
	AttractionSelection = "random"
	RadiiSelection      = "random"

	MinRadius = 100
	MaxRadius = 200

	// UI Settings
	UIWidth = 200
)

var (
	// Changable Settings
	ParticleSize  = 2
	Friction      = .99
	RepelRadius   = 10.
	RepelStrength = 1.
	Speed         = .03
	Types         = 5
	NParticles    = 300

	// Randomization Settings
	RandomFunc = rand.Float64
)
