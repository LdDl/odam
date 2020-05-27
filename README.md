# ODaM - Object Detection and Monitoring
# v0.1.0
ODaM is project aimed to do monitoring such as: pedestrian detection and counting, vehicle detection and counting, speed estimation of objects, sending detected objects to gRPC server for detailed analysis.

It's written on Go with a lot of [CGO](https://golang.org/cmd/cgo/).

## Work in progress

## Table of Contents
- [About](#about)
- [Installation](#installation)
- [Usage](#usage)
- [Screenshots](#screenshots)
- [Support](#support)
- [ToDo](#todo)
- [Dependencies](#dependencies)
- [License](#license)
- [Devs](#developers)

## About
ODaM is tool for doing monitoring via Darknet's neural network called Yolo V4 (paper: https://arxiv.org/abs/2004.10934).

It's built on top of [go-darknet](https://github.com/LdDl/go-darknet#go-darknet-go-bindings-for-darknet-yolo-v4-yolo-v3) which uses [AlexeyAB's fork of Darknet](https://github.com/AlexeyAB/darknet/#yolo-v4-and-yolo-v3v2-for-windows-and-linux). For doing computer vision stuff and video reading [GoCV](https://github.com/hybridgroup/gocv#gocv) is used.

## Installation
### notice: targeted for Linux users (no Windows/OSX instructions currenlty)
**Highly recommended to enable CUDA (GPU) in every installation step if it possible.**

1. Darknet - follow this [link](https://github.com/AlexeyAB/darknet#how-to-compile-on-linux-using-make). Do not forget to build library:
    ```Makefile
    LIBSO=1
    ```
    And then move it to /usr folder:
    ```shell
    [sudo] cp libdarknet.so /usr/[local]/lib/libdarknet.so && sudo cp include/darknet.h /usr/[local]/include/darknet.h
    ```
2. Go bindings for Darknet - [link](https://github.com/LdDl/go-darknet#installation)
3. GoCV - [link](https://github.com/hybridgroup/gocv#how-to-install).
4. Blob tracking library - [link](https://github.com/LdDl/gocv-blob#installation)
5. gRPC - [link](https://github.com/grpc/grpc-go#installation)

After steps above done:
```
go get github.com/LdDl/odam
go install github.com/LdDl/odam
```
Check if executable available
```
odam -h
```
and you will see something like this:
```
Usage of ./odam:
-settings string
        Path to application's settings (default "conf.json")
```

## Usage
### notice: targeted for Linux users (no Windows/OSX instructions currenlty)

* Prepare neural network stuff
    * Download YOLO's weights, configuration file and *.names file. Your way may warry, but here is our script: [download_data.sh](cmd/odam/download_data.sh)
        ```
        ./download_data_v3.sh
        ```
    * Make sure there is link to *.names file in YOLO's configuration file:
        ```
        [yolo]
        mask = 0,1,2
        anchors = 10,13,  16,30,  33,23,  30,61,  62,45,  59,119,  116,90,  156,198,  373,326
        classes=80
        num=9
        jitter=.3
        ignore_thresh = .7
        truth_thresh = 1
        random=1
        names = coco.names # <<========= here is the link to 'coco.names' file
        ```
* Prepare configuration file for application. Example of file: [conf.json](cmd/odam/conf.json). Description of fields:
```Makefile
{
    "video_settings": { # Video input settings
        "source": "rtsp://127.0.0.1:554/h264", # Link to RTSP stream
        "width": 1920, # Width of image in video source
        "height": 1080, # Height of image in video source
        "reduced_width": 640, # Desired width of image (for imshow and MJPEG streaming, also reduces inference time (processing > accuracy) for neural network)
        "reduced_height": 360, # Desired height of image (for imshow and MJPEG streaming, also reduces inference time (processing > accuracy) for neural network)
        "camera_id": "f2abe45e-aad8-40a2-a3b7-0c610c0f3dda" # Unique ID for video source (useful for 'client-server' model)
    },
    "neural_network_settings": { # YOLO neural network settings
        "darknet_cfg": "yolov3.cfg", # Path to configuration.file
        "darknet_weights": "yolov3.weights", # Path to weights wile
        "darknet_classes": "coco.names", # Path to *.names file (labels of objects)
        "conf_threshold": 0.2, # Confidence threshold
        "nms_threshold": 0.4, # NMS threshold (postprocessing)
        "target_classes": ["car", "motorbike", "bus", "train", "truck"] # What classes you want to detect (if you want to use public dataset, but ignore some classes)
    },
    "cuda_settings":{ # CUDA settings, currently useless
        "enable": true # CUDA settings, currently useless
    },
    "mjpeg_settings":{ # MJPEG streaming settings
        "imshow_enable": false, # Do you want to enable imshow() feature (useful for testing purposes)
        "enable": true, # Do you want to enable this feature?
        "port": 35678 # Listening port fo connections
    },
    "grpc_settings": { # gRPC 'client-server' model settings
        "enable": true, # Do you want to enable this feature?
        "server_ip": "localhost", # gRPC server's IP
        "server_port": 50051 # gRPC server's listening port
    },
    "tracker_settings": { # Tracked settings
        "lines_settings":[
            {
                "line_id": 1, # Unique ID for line id (useful for 'client-server' model)
                "begin": [150, 800], # [X1,Y1], start point of line (usually, left side)
                "end": [1600, 800], # [X2,Y2], end point of line (usually, right side)
                "direction": "to_detector", # Direction of line (possible values: 'to_detector' and 'from_detector')
                "detect_classes": ["car", "motorbike", "bus", "train", "truck"], # What classes must be cropped (as detected objects) that were captured by detection line.
                "rgba": [255, 0, 0, 0] # Color of detection line
            }
        ],
        "draw_track_settings": { # Tracker drawing settings (for WOW effect in imshow() or MJPEG streaming)
            "bbox_settings": { # Setting for bounding boxes (detected objects)
                "rgba": [255, 255, 0, 0], # Color of bounding box border
                "thickness": 2 # Thickness as is
            },
            "centroid_settings": { Setting for centroid of bounding boxes
                "rgba": [255, 0, 0, 0], # Color of circle
                "radius": 4, # Radius of circle
                "thickness": 2 # Thickness as is
            },
            "text_settings": { Setting for text above bounding boxes
                "rgba": [0, 255, 0, 0], # Text color
                "scale": 1.2, # Size of text
                "thickness": 2, # Thickness as is
                "font": "hershey_plain"
            }
        }
    },
    "matpprof_settings": { # pprof for GoCV. Useful for debugging
        "enable": true # Do you want to enable this feature?
    }
}
```

* Run
    ```
    odam --settings=conf.json
    ```

## Screenshots
* gocv.Imshow() output:

    <img src="screenshots/imshow_screen_1.png" width="720">

* MJPEG streaming output:

    <img src="screenshots/mjpeg_screen_1.png" width="720">

## Support
If you have troubles or questions please [open an issue](https://github.com/LdDl/odam/issues/new).
Feel free to make PR's (we do not have contributing guidelines currently, but we will someday)

## ToDo
Please see [ROADMAP.md](ROADMAP.md)

## Dependencies
* Bindings to [OpenCV](https://github.com/opencv/opencv) - [GoCV](https://github.com/hybridgroup/gocv#gocv). License is Apache-2.0
* MJPEG streaming via GoCV - [mjpeg](https://github.com/hybridgroup/mjpeg). No license currently
* Darknet (AlexeyAB's fork) - [darknet](https://github.com/AlexeyAB/darknet#yolo-v4-and-yolo-v3v2-for-windows-and-linux). License is YOLO LICENSE
* Golang binding to darknet - [go-darknet]https://github.com/LdDl/go-darknet#go-darknet-go-bindings-for-darknet-yolo-v4-yolo-v3). No license currently
* Tracking objects - [gocv-blob](https://github.com/LdDl/gocv-blob#gocv-blob). No license currently
* gRPC for doing "'client-server'" application - [grpc](https://github.com/grpc/grpc-go). License is Apache-2.0

## License
You can check it [here](LICENSE.md)

## Developers

cpllbstr https://github.com/cpllbstr

LdDl https://github.com/LdDl

Pavel7824 https://github.com/Pavel7824
