package attract

import (
	"life/settings"
	"log"
	"math/rand"
	"time"
)

var AttractionMatrix = [][]float64{}
var RadiusMatrix = [][]float64{}

func RandomizeAttractionMatrix() {
	switch settings.AttractionSelection {
	case "random":
		for j := 0; j < settings.MaxTypes; j++ {
			for i := 0; i < settings.MaxTypes; i++ {
				AttractionMatrix[i][j] = 2*settings.RandomFunc() - 1
			}
		}
	case "cluster":
		for j := 0; j < settings.MaxTypes; j++ {
			for i := 0; i < settings.MaxTypes; i++ {
				if i == j {
					AttractionMatrix[i][j] = 1
				} else {
					AttractionMatrix[i][j] = 0
				}
			}
		}
	default:
		log.Fatal("Invalid attraction type")
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < settings.MaxTypes; i++ {
		AttractionMatrix = append(AttractionMatrix, []float64{})
		RadiusMatrix = append(RadiusMatrix, []float64{})
		for j := 0; j < settings.MaxTypes; j++ {
			AttractionMatrix[i] = append(AttractionMatrix[i], 0)
			RadiusMatrix[i] = append(RadiusMatrix[i], 0)
		}
	}

	RandomizeAttractionMatrix()

	switch settings.RadiiSelection {
	case "random":
		for j := 0; j < settings.MaxTypes; j++ {
			for i := 0; i < settings.MaxTypes; i++ {
				RadiusMatrix[i][j] = 2*settings.RandomFunc() - 1
			}
		}

	case "equal":
		for j := 0; j < settings.MaxTypes; j++ {
			for i := 0; i < settings.MaxTypes; i++ {
				RadiusMatrix[i][j] = settings.MinRadius
			}
		}
	default:
		log.Fatal("Invalid radius type")
	}
}

type AttractionFunction func(float64, int8, int8) float64

var HalfRepelRadius = settings.RepelRadius / 2

func AbsoluteAttractionFunc() AttractionFunction {
	return func(d float64, t, ot int8) float64 {
		h := AttractionMatrix[t][ot]
		k := RadiusMatrix[t][ot]

		if d < settings.RepelRadius {
			return -(settings.RepelStrength / (d/settings.RepelRadius + 1)) + HalfRepelRadius
		}
		if d < settings.RepelRadius+2/k {
			return -h*(k*d-k*settings.RepelRadius-1) + h
		}
		return 0
	}
}

func ClusterAttractionFunc() AttractionFunction {
	return func(d float64, t int8, ot int8) float64 {
		if d < settings.RepelRadius {
			return -settings.RepelStrength / (d)
		}

		if t == ot {
			return .3 / d
		}
		return 0
	}
}

func SnakeAttractionFunc() AttractionFunction {
	return func(d float64, t int8, ot int8) float64 {
		if d < 10 {
			return -1 / (d)
		}

		if t == ot || t == ot+1 {
			return .5 / (d * 10)
		}

		return 0
	}
}

func DefaultAttractionFunc() AttractionFunction {
	return func(d float64, t int8, ot int8) float64 {
		if d < settings.RepelRadius {
			return -settings.RepelStrength / (d)
		}
		return AttractionMatrix[t][ot] / d
	}
}

func SimpleAttractionFunc() AttractionFunction {
	v := rand.NormFloat64()
	return func(d float64, t int8, ot int8) float64 {
		if d < settings.RepelRadius {
			return -settings.RepelStrength / (d)
		}
		return v / d
	}
}
