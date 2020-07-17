package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"mimg/pkg/img"
	"os"
	"path"
)

var confFile string
var srcDir string
var outDir string
var step int

func init() {
	flag.IntVar(&step, "step", 3, fmt.Sprintf("step value for partition"))
	flag.StringVar(&confFile, "c", "./", fmt.Sprintf("config file"))
	flag.StringVar(&srcDir, "s", "./", fmt.Sprintf("src file dir "))
	flag.StringVar(&outDir, "o", "./out", fmt.Sprintf("output dir"))
	flag.Parse()

	if srcDir == outDir {
		panic(fmt.Sprintf("%v same as %v", srcDir, outDir))
	}
}

// type Point struct {
// 	X float64
// 	Y float64
// }

func loadJSON(name string) (data map[string]image.Point, err error) {
	file, err := os.Open(name)
	if err != nil {
		return
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	return
}

func saveJSON(name string, data interface{}) (err error) {
	bts, err := json.Marshal(data)
	if err != nil {
		return
	}

	file, err := os.Create(name)
	if err != nil {
		return
	}
	defer file.Close()

	n, err := file.Write(bts)
	if err != nil || n < 1 {
		return fmt.Errorf("Save Json Error %v-%v", err, n)
	}

	return
}

func loadImgPoints(name string) (points map[image.Point]color.RGBA, err error) {
	reader, err := os.Open(name)
	if err != nil {
		return
	}

	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return
	}

	bounds := img.Bounds()
	if bounds.Empty() {
		err = errors.New("empty file")
		return
	}

	width := bounds.Dx()
	height := bounds.Dy()
	points = make(map[image.Point]color.RGBA)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			if r == 0 && g == 0 && b == 0 && a == 0 {
				continue
			}

			points[image.Point{X: x, Y: y}] = color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
		}
	}

	return
}

func log(file os.File, data string) {

}

func main() {
	data, err := loadJSON(confFile)
	if err != nil {
		panic(fmt.Sprintf("load conf[%v] err:%v", confFile, err))
	}

	ext := ".png"
	outputJSON := make(map[string]string)
	for srcName, srcPoint := range data {
		if srcPoint.X == 0 || srcPoint.Y == 0 {
			fmt.Printf("invalid point %v-%v\n", srcName, srcPoint)
			continue
		}

		points, err := loadImgPoints(path.Join(srcDir, srcName+ext))
		if err != nil {
			panic(fmt.Sprintf("load img point err %v", err))
		}

		fmt.Printf("split file[%v] %v-%v\n", srcName, len(points), srcPoint)

		src := img.NewImg(points)

		bounds := src.Boundary()
		srcLeft := srcPoint.X - bounds.Dx() * 10 / 2
		srcTop := srcPoint.Y - bounds.Dy() * 10 / 2

		//fmt.Printf("src df:%v %v-%v\n", srcName, bounds.Dx(), bounds.Dy())
		//fmt.Printf("src bound:%v %v-%v\n", srcName, bounds.Min, bounds.Max)
		//fmt.Printf("src point:%v %v-%v\n", srcName, srcPoint.X, srcPoint.Y)
		//fmt.Printf("src file:%v %v-%v\n", srcName, srcLeft, srcTop)

		// 0, 840
		subRegions := src.SplitPoints(step)
		for idx, region := range subRegions {
			itemName := fmt.Sprintf("%v-%v", srcName, idx)
			imgName := itemName + ext
			imgData := img.NewImg(region)
			n, err := imgData.Encode(path.Join(outDir, imgName), nil)
			if err != nil || n < 1 {
				panic(fmt.Sprintf("%v-%v %v-%v", idx, imgName, n, err))
			}

			bound := imgData.Boundary()
			x := (bound.Min.X + bound.Max.X) * 10 / 2
			y := (bound.Min.Y + bound.Max.Y) * 10 / 2
			outputJSON[itemName] = fmt.Sprintf("%d,%d", srcLeft+x, srcTop+y)
		}
	}

	jsonName := path.Base(confFile) + ".json"
	err = saveJSON(path.Join(outDir, jsonName), outputJSON)
	if err != nil {
		panic(fmt.Sprintf("save json err %v", err))
	}
}
