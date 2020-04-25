package img

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"testing"
)

func TestPartition(t *testing.T) {
	reader, err := os.Open("./19.png")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	//reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	img, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	if bounds.Empty() {
		return
	}

	fmt.Printf("before region %v\n", bounds.String())

	width := bounds.Dx()
	height := bounds.Dy()
	points := make(map[image.Point]color.RGBA)
	zero := 0
	f := 0
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			if r == 0 && g== 0 && b == 0 && a == 0{
				zero += 1
				continue
			}

			points[image.Point{X:x, Y:y}] = color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
		}
	}

	mImg := Img{Points:points}
	boundary := mImg.Boundary()

	fmt.Printf("after region %v-%v-%v-%v\n", boundary.String(), f, zero, len(points))

	regions := mImg.SplitPoints(2)
	if len(regions) < 2 {
		fmt.Printf("only one region\n")
		return
	}

	total := 0
	for idx, pts := range regions {
		img := Img{Points:pts}
		if img.Boundary().Empty() {
			fmt.Printf("%v is empty\n", idx)
			continue
		}

		b := img.Boundary()
		fmt.Printf("%d ----- %v\n", idx, b.String())
		total += len(pts)

		n, err := img.Encode(fmt.Sprintf("piece-%v.png", idx), nil)
		if err != nil || n < 1 {
			panic(fmt.Sprintf("encode %v-%v\n", err, n))
		}
	}

	fmt.Printf("%v-%v--%v\n", len(regions), len(points), total)
}