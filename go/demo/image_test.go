package demo

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	// 创建一个图像大小为 100x100 像素的RGBA图像
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// 填充图像为蓝色
	fillBlue(img)

	fillPoint(img)

	// 保存图像到文件
	saveImage(img, "output.png")
}

func fillPoint(img *image.RGBA) {
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()
	for i := 0; i < 1000; i++ {
		x := rand.Intn(dx)
		y := rand.Intn(dy)
		img.Set(x, y, color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255))})
		time.Sleep(1 * time.Second)
		saveImage(img, "output.png")
	}
}

// fillBlue 填充图像为蓝色
func fillBlue(img *image.RGBA) {
	blue := color.RGBA{255, 255, 255, 255} // RGBA颜色，这里是纯蓝色
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, blue)
		}
	}
}

// saveImage 保存图像到文件
func saveImage(img image.Image, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 以PNG格式保存图像
	err = png.Encode(file, img)
	if err != nil {
		panic(err)
	}
}

func TestGif(t *testing.T) {
	// 创建一个帧的切片
	var frames []*image.Paletted
	// 创建一个颜色调色板
	palette := []color.Color{
		color.White,
		color.RGBA{255, 0, 0, 255}, // 红色
		color.RGBA{0, 255, 0, 255}, // 绿色
		color.RGBA{0, 0, 255, 255}, // 蓝色
	}

	// 在每一帧中填充不同的颜色
	for i := 0; i < 4; i++ {
		frame := image.NewPaletted(image.Rect(0, 0, 100, 100), palette)
		fillColor(frame, i)
		frames = append(frames, frame)
	}

	// 创建一个 GIF 图片，设置帧和帧之间的延迟（以10毫秒为单位）
	gifImage := &gif.GIF{
		Image:     frames,
		Delay:     []int{1, 1, 1, 1},
		LoopCount: 0, // 0 表示无限循环
	}

	// 将 GIF 图片保存到文件
	file, err := os.Create("output.gif")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = gif.EncodeAll(file, gifImage)
	if err != nil {
		panic(err)
	}
}

// fillColor 填充帧的颜色
func fillColor(img *image.Paletted, colorIndex int) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.SetColorIndex(x, y, uint8(colorIndex+1))
		}
	}
}
