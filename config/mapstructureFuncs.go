package config

import (
	"fmt"
	"github.com/Pramod-Devireddy/go-exprtk"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strconv"
	"strings"
)

func MapstructureStringToRatio() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Float64 {
			return data, nil
		}

		raw := data.(string)
		exp := exprtk.NewExprtk()
		exp.SetExpression(raw)
		exp.CompileExpression()
		return exp.GetEvaluatedValue(), nil
	}
}

func MapstructureStringToColor() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(Color{}) {
			return data, nil
		}

		raw := data.(string)

		// Trying default values
		c, ok := Colors[raw]
		if ok {
			return c, nil
		}

		// Trying to decode string as a color
		strs := strings.Split(raw, ",")
		if len(strs) != 3 {
			return nil, fmt.Errorf("unable to decode color %s (format: r,g,b)", raw)
		}
		rgb := make([]uint8, 3, 3)
		for i := 0; i < 3; i++ {
			ui64, err := strconv.ParseUint(strs[i], 16, 8)
			if err != nil {
				return nil, fmt.Errorf("unable to decode color value %s", strs[0])
			}
			rgb[i] = uint8(ui64)
		}

		return Color{rgb[0], rgb[1], rgb[2]}, nil
	}
}

func MapstructureStringToOrientation() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(Orientation("")) {
			return data, nil
		}

		raw := data.(string)
		if raw != string(Portrait) && raw != string(Landscape) {
			return nil, fmt.Errorf("orientation can only be portrait or landscape")
		}

		return Orientation(raw), nil
	}
}
