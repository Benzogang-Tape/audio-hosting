package main

import (
	"flag"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/config"
)

func main() {
	var (
		out    string
		indent int
	)

	flag.StringVar(&out, "o", "", "output file")
	flag.IntVar(&indent, "i", 2, "indent")
	flag.Parse()

	var file *os.File
	if out == "" {
		file = os.Stdout
	} else {
		dir := filepath.Dir(out)

		err := os.MkdirAll(dir, 0o700) //nolint:mnd
		if err != nil {
			panic(err)
		}

		file, err = os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	}

	mp := StructToMap[config.Config]()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(indent)

	err := encoder.Encode(mp)
	if err != nil {
		panic(err)
	}
}

func StructToMap[T any]() map[string]any {
	tp := reflect.TypeFor[T]()
	if tp.Kind() != reflect.Struct {
		panic("not a struct")
	}

	return structToMapTp(tp)
}

func structToMapTp(tp reflect.Type) map[string]any {
	fieldsNum := tp.NumField()
	res := make(map[string]any, fieldsNum)

	for i := range fieldsNum {
		f := tp.Field(i)
		name := f.Tag.Get("yaml")

		if name == "" {
			name = f.Name
		}

		if f.Type.Kind() == reflect.Map || f.Type.Kind() == reflect.Struct {
			res[name] = structToMapTp(f.Type)
			continue
		}

		value, ok := f.Tag.Lookup("e.g")
		if ok {
			res[name] = guessTypeAndParse(value)
			continue
		}

		value, ok = f.Tag.Lookup("env-default")
		if ok {
			res[name] = guessTypeAndParse(value)
			continue
		}

		res[name] = nil
	}

	return res
}

func guessTypeAndParse(value string) any {
	if value == "" {
		return nil
	}

	if parsedValue, err := strconv.ParseInt(value, 10, 64); err == nil {
		return parsedValue
	}

	if parsedValue, err := strconv.ParseFloat(value, 64); err == nil {
		return parsedValue
	}

	if parsedValue, err := strconv.ParseBool(value); err == nil {
		return parsedValue
	}

	return value
}
