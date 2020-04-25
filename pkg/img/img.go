package img

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path"
)

type Img struct {
	Points map[image.Point] color.RGBA
}

func NewImg(points map[image.Point]color.RGBA) *Img {
	return &Img{Points: points}
}

func (m *Img) At(p image.Point) (color.RGBA, bool) {
	rgba, ok := m.Points[p]
	return rgba, ok
}

func (m *Img) Pos() image.Point {
	rect := m.Boundary()

	return image.Point{
		X: (rect.Min.X + rect.Max.X) / 2,
		Y: (rect.Min.Y + rect.Max.Y) / 2,
	}
}

func (m *Img) Boundary() image.Rectangle {
	if len(m.Points) < 1 {
		return image.Rectangle{}
	}

	maxX, minX, maxY, minY := 0, math.MaxInt32, 0, math.MaxInt32
	for pt, _ := range m.Points {
		if maxX < pt.X {
			maxX = pt.X
		}

		if minX > pt.X {
			minX = pt.X
		}

		if maxY < pt.Y {
			maxY = pt.Y
		}

		if minY > pt.Y {
			minY = pt.Y
		}
	}

	return image.Rectangle{
		Min: image.Point{X: minX, Y: minY},
		Max: image.Point{X: maxX + 1, Y: maxY + 1},
	}
}

//
func (m *Img) Encode(name string, opts *jpeg.Options) (int, error) {
	boundary := m.Boundary()
	if boundary.Empty() {
		return 0, fmt.Errorf("empty ")
	}

	imageData := image.NewRGBA(boundary)
	for k, rgba := range m.Points {
		imageData.SetRGBA(k.X, k.Y, rgba)
	}

	var b bytes.Buffer
	ext := path.Ext(name)
	switch ext {
	case ".png":
		err := png.Encode(&b, imageData)
		if err != nil {
			return 0, err
		}

	case ".jpg":
		err := jpeg.Encode(&b, imageData, opts)
		if err != nil {
			return 0, err
		}

	default:
		return 0, fmt.Errorf("inivalid %v", ext)
	}

	//
	f, err := os.Create(name)
	if err != nil {
		return 0, err
	}

	return f.Write(b.Bytes())
}

func (m *Img) SplitPoints(step int) []map[image.Point]color.RGBA {
	boundary := m.Boundary()
	if boundary.Empty() {
		return nil
	}

	//
	//fmt.Printf("-----split points------------%v:%v:%v:%v\n", left, top, right, bottom)
	// copy
	pts := make(map[image.Point]bool)
	for k := range m.Points {
		pts[k] = false
	}

	randPoint := func(pts map[image.Point]bool) (p image.Point, err error) {
		for k := range pts {
			return k, nil
		}

		return p, fmt.Errorf("finished")
	}

	regions := make([]map[image.Point]color.RGBA, 0, 8)
	for {
		pt, err := randPoint(pts)
		if err != nil {
			break
		}

		sub := bfsLoop(boundary, step, pt, pts)
		if len(sub) < 25 {
			fmt.Printf("-------- ignore little subregion %v\n", len(sub))
			delete(pts, pt)
			continue
		}

		r := make(map[image.Point]color.RGBA)
		for k := range sub {
			delete(pts, k)
			if rgba, ok := m.At(k); ok {
				r[k] = rgba
			}
		}

		regions = append(regions, r)
		//fmt.Printf("regin %v\n", len(r))
	}

	return regions
}

const (
	DirectionTop = 0
	DirectionTopRight = 1
	DirectionRight = 2
	DirectionRightBottom = 3
	DirectionBottom = 4
	DirectionLeftBottom = 5
	DirectionLeft = 6
	DirectionLeftTop = 7
	DirectionMax = 8
)

func pointRect(bounds image.Rectangle, pt image.Point, step int) image.Rectangle {
	region := image.Rectangle {
		Min: image.Point{
			X: pt.X - step,
			Y: pt.Y - step,
		},
		Max: image.Point{
			X: pt.X + step,
			Y: pt.Y + step,
		},
	}

	return bounds.Intersect(region)
}

func rectPoints(rect image.Rectangle, points map[image.Point]bool) (pts []image.Point) {
	pts = make([]image.Point, 0, rect.Dx() * rect.Dy())
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			p := image.Point{X:x, Y:y}
			if _, ok := points[p]; !ok {
				continue
			}

			pts = append(pts, p)
		}
	}

	return pts
}

func nextPoint(pt image.Point, direction, step int) (next image.Point, err error) {
	switch direction {
	case DirectionTop:
		next = image.Point{
			X: pt.X,
			Y: pt.Y + step * 2,
		}

	case DirectionTopRight:
		next = image.Point{
			X: pt.X + step * 2,
			Y: pt.Y + step * 2,
		}

	case DirectionRight:
		next = image.Point{
			X: pt.X + step * 2,
			Y: pt.Y,
		}

	case DirectionRightBottom:
		next = image.Point{
			X: pt.X + step * 2,
			Y: pt.Y - step * 2,
		}

	case DirectionBottom:
		next = image.Point{
			X: pt.X,
			Y: pt.Y - step * 2,
		}

	case DirectionLeftBottom:
		next = image.Point{
			X: pt.X - step * 2,
			Y: pt.Y - step * 2,
		}

	case DirectionLeft:
		next = image.Point{
			X: pt.X - step * 2,
			Y: pt.Y,
		}

	case DirectionLeftTop:
		next = image.Point{
			X: pt.X - step * 2,
			Y: pt.Y + step * 2,
		}

	default:
		err = errors.New("invalid direction")
		return
	}

	return
}

func bfsLoop(bounds image.Rectangle, step int, pt image.Point, points map[image.Point]bool) map[image.Point]bool{
	_, ok := points[pt]
	if !ok {
		return nil
	}

	region := make(map[image.Point]bool)
	expect := map[image.Point]bool {pt: true}

	rect := pointRect(bounds, pt, step)
	if !rect.Empty() {
		pts := rectPoints(rect, points)
		for _, item := range pts {
			region[item] = true
		}
	}

	stack := make([]image.Point, 0, 1024)
	stack = append(stack, pt)
	for {
		if len(stack) < 1 {
			break
		}

		head := stack[0]
		for i:=0; i < DirectionMax; i++ {
			nextPoint, err := nextPoint(head, i, step)
			if err != nil {
				continue
			}

			if _, ok := expect[nextPoint]; ok { // calc over
				continue
			}

			expect[nextPoint] = true

			interSect := pointRect(bounds, nextPoint, step)
			if interSect.Empty() {
				continue
			}

			nearPoints := rectPoints(interSect, points)
			if len(nearPoints) < 1 {
				continue
			}

			for _, item := range nearPoints {
				region[item] = true
			}

			stack = append(stack, nextPoint)
		}

		stack = stack[1:]
	}

	return region
}
