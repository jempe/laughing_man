package main

import (
	"fmt"
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

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed:")
			return
		}
		if img.Empty() {
			continue
		}

		buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(buf.GetBytes())
		buf.Close()
	}
}
