package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"net/http"
	"os"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

const MinimumArea = 3000

var (
	deviceID string
	err      error
	webcam   *gocv.VideoCapture
	stream   *mjpeg.Stream
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage:\n\tlaughing_man [video source]")
		return
	}

	deviceID := os.Args[1]

	webcam, err = gocv.OpenVideoCapture(deviceID)

	webcam.Set(gocv.VideoCaptureFPS, 4)

	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	stream = mjpeg.NewStream()

	go mjpegCapture()

	http.Handle("/", stream)
	log.Fatal(http.ListenAndServe("0.0.0.0:3000", nil))
}

func mjpegCapture() {
	img := gocv.NewMat()
	defer img.Close()

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	xmlFile := "classifiers/haarcascade_frontalface_default.xml"

	if !classifier.Load(xmlFile) {
		fmt.Printf("Error reading cascade file: %v\n", xmlFile)
		return
	}

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed:")
			return
		}
		if img.Empty() {
			continue
		}

		// detect faces
		rects := classifier.DetectMultiScaleWithParams(img, 1.5, 6, 0, image.Pt(0, 0), image.Pt(0, 0))
		fmt.Printf("found %d faces\n", len(rects))

		// draw a rectangle around each face on the original image,
		// along with text identifying as "Human"
		for _, r := range rects {
			//gocv.Rectangle(&img, r, blue, 3)

			center := image.Point{X: (r.Min.X + ((r.Max.X - r.Min.X) / 2)), Y: (r.Min.Y + ((r.Max.Y - r.Min.Y) / 2))}

			radius := (r.Max.X - r.Min.X) / 2

			thickness := -1

			gocv.Circle(&img, center, radius, blue, thickness)

			//size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
			//pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			//gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
		}

		buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(buf.GetBytes())
		buf.Close()
	}
}
