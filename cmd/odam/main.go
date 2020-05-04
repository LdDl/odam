package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"log"
	"net/http"
	"time"

	"github.com/LdDl/gocv-blob/blob"
	"github.com/LdDl/odam"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

var (
	settingsFile = flag.String("settings", "conf.json", "Path to application's settings")
)

var (
	window          *gocv.Window
	stream          *mjpeg.Stream
	imagesChannel   chan *odam.FrameData
	detectedChannel chan []odam.DetectedObject
	detected        []odam.DetectedObject
	allblobies      *blob.Blobies
)

func main() {

	// Settings
	flag.Parse()
	settings, err := odam.NewSettings(*settingsFile)
	if err != nil {
		log.Fatalln(err)
	}

	if settings.MjpegSettings.Enable {
		stream = mjpeg.NewStream()
		go func() {
			fmt.Printf("Starting MJPEG on http://localhost:%d\n", settings.MjpegSettings.Port)
			http.Handle("/", stream)
			log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", settings.MjpegSettings.Port), nil))
		}()
	}

	allblobies = blob.NewBlobiesDefaults()

	videoCapturer, err := gocv.OpenVideoCapture(settings.VideoSettings.Source)
	if err != nil {
		log.Fatalln(err)
	}
	if settings.MjpegSettings.ImshowEnable {
		fmt.Println("Press 'ESC' to stop imshow()")
		window = gocv.NewWindow("ODAM v0.1.0")
		window.ResizeWindow(settings.VideoSettings.ReducedWidth, settings.VideoSettings.ReducedHeight)
		defer window.Close()
	}
	imagesChannel = make(chan *odam.FrameData, 1)
	img := odam.NewFrameData()
	// Read first frame
	if ok := videoCapturer.Read(&img.ImgSource); !ok {
		log.Fatalf("Error cannot read video %v\n", settings.VideoSettings.Source)
	}
	gocv.Resize(img.ImgSource, &img.ImgScaled, image.Point{X: settings.VideoSettings.ReducedWidth, Y: settings.VideoSettings.ReducedHeight}, 0, 0, gocv.InterpolationDefault)

	// Read frames in a loop
	for {
		if ok := videoCapturer.Read(&img.ImgSource); !ok {
			fmt.Println("Can't read next frame, stop grabbing")
			break
		}
		if img.ImgSource.Empty() {
			fmt.Println("Empty frame has been detected. Sleep for 400 ms")
			time.Sleep(400 * time.Millisecond)
			continue
		}
		gocv.Resize(img.ImgSource, &img.ImgScaled, image.Point{X: settings.VideoSettings.ReducedWidth, Y: settings.VideoSettings.ReducedHeight}, 0, 0, gocv.InterpolationDefault)

		select {
		case detected = <-detectedChannel:
			processFrame(img)
		default:
			// show current frame without blocking, so do nothing here
		}

		if settings.MjpegSettings.ImshowEnable {
			for i := range settings.TrackerSettings.LinesSettings {
				settings.TrackerSettings.LinesSettings[i].VLine.Draw(&img.ImgScaled)
			}
			window.IMShow(img.ImgScaled)
			if window.WaitKey(1) == 27 {
				break
			}
		}
		if settings.MjpegSettings.Enable {
			buf, _ := gocv.IMEncode(".jpg", img.ImgScaled)
			stream.UpdateJPEG(buf)
		}

	}

	// Release memory
	img.Close()
	window.Close()

	// pprof
	if settings.PPROFSettings.Enable {
		var b bytes.Buffer
		// go run -tags matprofile main.go
		// gocv.MatProfile.WriteTo(&b, 1)
		fmt.Print(b.String())
	}
}

func processFrame(fd *odam.FrameData) {
	frame := odam.NewFrameData()
	fd.ImgSource.CopyTo(&frame.ImgSource)
	fd.ImgScaled.CopyTo(&frame.ImgScaled)
	imagesChannel <- frame
}

func performDetection() {
	for {
		frame := <-imagesChannel

		frame.Close()
	}
}
