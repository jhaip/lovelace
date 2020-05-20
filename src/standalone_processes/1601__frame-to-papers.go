// How to run:
//
// tracking [camera ID]
//
// 		go run ./cmd/haip-dots/main.go 0
//
// +build example

package main

import (
	"fmt"
	"image/color"
	"os"

	"gocv.io/x/gocv"
	// "gocv.io/x/gocv/contrib"
	// "gocv.io/x/gocv/features2d"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\ttracking [camera ID]")
		return
	}

	// parse args
	deviceID := os.Args[1]

	// open webcam
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("Tracking")
	defer window.Close()

	// create simple blob detector with parameters
	params := gocv.NewSimpleBlobDetectorParams()
	params.SetMinThreshold(50)
	params.SetMaxThreshold(230)
	params.SetFilterByCircularity(true)
	params.SetMinCircularity(0.5)
	params.SetFilterByArea(true)
	params.SetMinArea(9)
	params.SetFilterByInertia(false)
	bdp := gocv.NewSimpleBlobDetectorWithParams(params)
	defer bdp.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// read an initial image
	if ok := webcam.Read(&img); !ok {
		fmt.Printf("cannot read device %v\n", deviceID)
		return
	}

	// let the user mark a ROIs (projector corners) to track
	rects := gocv.SelectROIs("Tracking", img)
	if len(rects) != 4 {
		fmt.Printf("user cancelled roi selection or did not specify 4 corners\n")
		return
	}

	// color for the rect to draw
	blue := color.RGBA{0, 0, 255, 0}
	fmt.Printf("Start reading device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// detect blobs/keypoints
		kp := bdp.Detect(img)
		fmt.Printf("Keypoints detected: $v\n", len(kp))

		// draw the keypoints on the webcam image
		simpleKP := gocv.NewMat()
		gocv.DrawKeyPoints(img, kp, &simpleKP, blue, gocv.DrawDefault)

		// show the image in the window, and wait 10 millisecond
		window.IMShow(simpleKP)
		if window.WaitKey(10) >= 0 {
			break
		}
	}
}
