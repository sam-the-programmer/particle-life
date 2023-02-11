package attract

import (
	"life/settings"
	"log"
)

func DefaultAttraction(d float64) float64 {
	return .5 / d
}

var AttractionMatrix = [settings.TYPES][settings.TYPES]float64{}
var RadiusMatrix = [settings.TYPES][settings.TYPES]float64{}

func init() {
	switch settings.ATTRACTION {
	case "random":
		for j := 0; j < settings.TYPES; j++ {
			for i := 0; i < settings.TYPES; i++ {
				AttractionMatrix[i][j] = 2*settings.RANDOMFUNC() - 1
			}
		}
	case "cluster":
		for j := 0; j < settings.TYPES; j++ {
			for i := 0; i < settings.TYPES; i++ {
				if i == j {
					AttractionMatrix[i][j] = -1
				} else {
					AttractionMatrix[i][j] = 1
				}
			}
		}
	default:
		log.Fatal("Invalid attraction type")
	}

	switch settings.RADII {
	case "random":
		for j := 0; j < settings.TYPES; j++ {
			for i := 0; i < settings.TYPES; i++ {
				RadiusMatrix[i][j] = 2*settings.RANDOMFUNC() - 1
			}
		}

	case "equal":
		for j := 0; j < settings.TYPES; j++ {
			for i := 0; i < settings.TYPES; i++ {
				RadiusMatrix[i][j] = settings.MINRADIUS
			}
		}
	default:
		log.Fatal("Invalid radius type")
	}
}

type AttractionFunction func(float64, int8, int8) float64

const HALFREPELRADIUS = settings.REPELRADIUS / 2

func BaseAttractionAbs(d float64, t, ot int8) float64 {
	h := AttractionMatrix[t][ot]
	k := RadiusMatrix[t][ot]

	if d < settings.REPELRADIUS {
		return -(settings.REPELSTRENGTH / (d/settings.REPELRADIUS + 1)) + HALFREPELRADIUS
	}
	if d < settings.REPELRADIUS+2/k {
		return -h*(k*d-k*settings.REPELRADIUS-1) + h
	}
	return 0
}

func RandomAttractionFunc() AttractionFunction {
	return func(d float64, t int8, ot int8) float64 {
		return BaseAttractionAbs(d, t, ot)
	}
}

func DefaultAttractionFunc() AttractionFunction {
	return func(d float64, t int8, ot int8) float64 {
		return AttractionMatrix[t][ot] / d
	}
}
