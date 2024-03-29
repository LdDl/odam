# ODaM - Object Detection and Monitoring
[![GoDoc](https://godoc.org/github.com/LdDl/odam?status.svg)](https://godoc.org/github.com/LdDl/odam) [![Sourcegraph](https://sourcegraph.com/github.com/LdDl/odam/-/badge.svg)](https://sourcegraph.com/github.com/LdDl/odam?badge) [![Go Report Card](https://goreportcard.com/badge/github.com/LdDl/odam)](https://goreportcard.com/report/github.com/LdDl/odam) [![GitHub tag](https://img.shields.io/github/tag/LdDl/odam.svg)](https://github.com/LdDl/odam/releases)
# v0.9.0
ODaM is project aimed to do monitoring such as: pedestrian detection and counting, vehicle detection and counting, speed estimation of objects, sending detected objects to gRPC server for detailed analysis.



YOLOv4 + Kalman filter for tracking             |  YOLOv4 + simple centroid tracking
:-------------------------:|:-------------------------:
<img src="screenshots/yolov4-kalman.gif" width="640">  |  <img src="screenshots/yolov4-simple.gif" width="640">

YOLOv4 Tiny + Kalman filter for tracking             |  YOLOv4 Tiny + simple centroid tracking
:-------------------------:|:-------------------------:
<img src="screenshots/yolov4-tiny-kalman.gif" width="640">  |  <img src="screenshots/yolov4-tiny-simple.gif" width="640">

## Work in progress

We are working on this.

Not too fast, but it is what it is.

Former version (until [#21](https://github.com/LdDl/odam/pull/21)) of this repository were containing a lot of business logic via [go-darknet](https://github.com/LdDl/go-darknet#go-darknet-go-bindings-for-darknet-yolo-v4-yolo-v3). Current version depends on OpenCV's DNN module.

## Table of Contents
- [ODaM - Object Detection and Monitoring](#odam---object-detection-and-monitoring)
- [v0.9.0](#v090)
  - [Work in progress](#work-in-progress)
  - [Table of Contents](#table-of-contents)
  - [About](#about)
  - [QA section](#qa-section)
  - [Installation](#installation)
    - [notice: targeted for Linux users (no Windows/OSX instructions currenlty)](#notice-targeted-for-linux-users-no-windowsosx-instructions-currenlty)
  - [Usage](#usage)
    - [notice: targeted for Linux users (no Windows/OSX instructions currenlty)](#notice-targeted-for-linux-users-no-windowsosx-instructions-currenlty-1)
  - [Screenshots](#screenshots)
  - [Support](#support)
  - [Roadmap](#roadmap)
  - [Dependencies](#dependencies)
  - [License](#license)
  - [Developers](#developers)

## About
ODaM is tool for doing monitoring via Darknet's neural network called Yolo V4 (paper: https://arxiv.org/abs/2004.10934).

It's built on top of [GoCV](https://github.com/hybridgroup/gocv#gocv).

## QA section
> Who are you and what do you do?

There is info about me here: https://github.com/LdDl

You can have chat with me in Telegram/Gmail

> Is this library / software or even framework?

I think about it as software with library capabilities.

> What it capable of?

Not that much currently:

* Object detection via darknet: OpenCV::ddn module is used
* Object tracking via two possible techniques: Kalman tracking (filtering) or Centroid tracking;
* Sending data to dedicated gRPC server;
* MJPEG / imshow optional visual output;
* Speed estimation based of GIS calculations (via matching pixels to WGS84).

> Why Go?

Well, C++ is a killer in computer vision field and Python has a great battery included bindings for C++ code.

But I do no think that I'm ready to build gRPC/REST or any other web components of this software in C++ or Python (C++ is not that easy and Python...I just don't like Python syntax). That's why I prefer to stick with Go.

> Why did you pick JSON for configuration purposes instead of TOML/YAML/INI or any other well-suited formats?

1. Compared to TOML, JSON is not that 'human friendly', but still readable.
2. It is in standart Go's library.
3. Well, it is in standart Go's library.
4. You got the idea.

> Why bindings to Darknet instead of Opencv included stuff?

Sometimes you just do not need full OpenCV installation for object detection. I have such ANPR projet here: https://github.com/LdDl/license_plate_recognition
I guess when I'm done with stable core I might switch from Go's Darknet bindings to OpenCV one (since ODaM-project requires OpenCV installation obviously)

> What are your plans?

There is [ROADMAP.md](ROADMAP.md), but overall I am planning to extend capabilities of software: 
* Improve perfomance
* Implement some cool tracking techniques (e.g. [SORT](https://arxiv.org/abs/1602.00763))
* Do gRPC accepting microservice for enabling software to catch information from external devices/systems/microservices and etc. E.g: you want to send message 'there is red light on traffic light" to instance of software, then it would look like _grpcServer.Send('there is red light on traffic light')_. After that any captured object will have state with message above in it. So you can catch traffic offenders.
* Introduce convex polygon based calculations (same as virtual lines but for polygons)

> How to help you?

If you are here, then you are already helped a lot, since you noticed my existence :harold_face:

If you want to make PR for some undone features (algorithms mainly) I'll glad to take a look.

## Installation
### notice: targeted for Linux users (no Windows/OSX instructions currenlty)
**Need to enable CUDA (GPU) in every installation step where it's possible.**

1. Install CUDA (we recommend version 10.2)
    ```bash
    wget https://developer.download.nvidia.com/compute/cuda/repos/ubuntu1804/x86_64/cuda-ubuntu1804.pin
    sudo mv cuda-ubuntu1804.pin /etc/apt/preferences.d/cuda-repository-pin-600
    wget http://developer.download.nvidia.com/compute/cuda/10.2/Prod/local_installers/cuda-repo-ubuntu1804-10-2-local-10.2.89-440.33.01_1.0-1_amd64.deb
    sudo dpkg -i cuda-repo-ubuntu1804-10-2-local-10.2.89-440.33.01_1.0-1_amd64.deb
    sudo apt-key add /var/cuda-repo-10-2-local-10.2.89-440.33.01/7fa2af80.pub
    sudo apt-get update
    sudo apt-get -y install cuda
    echo 'export PATH=/usr/local/cuda/bin:$PATH' >> ~/.bashrc
    echo 'export LD_LIBRARY_PATH=/usr/local/cuda/lib64:LD_LIBRARY_PATH'  >> ~/.bashrc
    source ~/.bashrc
    ```
2. Install cuDNN (we recommend version v7.6.5 (November 18th, 2019), for CUDA 10.2)
    Go to [NVIDIA's site](https://developer.nvidia.com/rdp/cudnn-download) and download *.deb package. After downloading *.deb package install it:
    ```bash
    sudo dpkg -i libcudnn7_7.6.5.32-1+cuda10.2_amd64.deb
    sudo dpkg -i libcudnn7-dev_7.6.5.32-1+cuda10.2_amd64.deb
    sudo dpkg -i libcudnn7-doc_7.6.5.32-1+cuda10.2_amd64.deb
    ```
    Do not forget to check if cuDNN installed properly:
    ```bash
    cp -r /usr/src/cudnn_samples_v7/ $HOME
    cd  $HOME/cudnn_samples_v7/mnistCUDNN
    make clean && make
    ./mnistCUDNN
    cd -
    ```
3. GoCV - [instructions link](https://github.com/hybridgroup/gocv#how-to-install).
4. Blob tracking library - [instructions link](https://github.com/LdDl/gocv-blob#installation)
5. If you want to use gRPC client-server model: gRPC - [instructions link](https://github.com/grpc/grpc-go#installation)

   You need to implement your gRPC server as following proto-file: https://github.com/LdDl/odam/blob/master/yolo_grpc.proto.

   If you need to rebuild *.pb.go file, call this is from project root folder:
   ```
   protoc -I . yolo_grpc.proto --go_out=plugins=grpc:.
   ```
   In case of my needs I need to detect license plates on vehicles and do OCR on server-side: you can take a look on https://github.com/LdDl/license_plate_recognition for gRPC server example

After steps above done:
```
go install github.com/LdDl/odam/cmd/odam
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
    * Download YOLO's weights, configuration file and *.names file. Your way may warry, but here is our script: [download_data.sh](cmd/odam/download_data_v4.sh)
        ```
        ./download_data_v4.sh
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
    "classes_settings": [ # classes settings (according to 'target_classes' in 'neural_network_settings')
        {
            "class_name": "car", # Corresponding class label
            "drawing_settings": {
                "bbox_settings": { # Setting for bounding boxes (detected objects)
                    "rgba": [255, 255, 0, 0], # Color of bounding box border
                    "thickness": 2 # Thickness as is
                },
                "centroid_settings": { # Setting for centroid of bounding boxes
                    "rgba": [255, 0, 0, 0], # Color of circle
                    "radius": 4, # Radius of circle
                    "thickness": 2 # Thickness as is
                },
                "text_settings": { # Setting for text above bounding boxes
                    "rgba": [0, 255, 0, 0], # Text color
                    "scale": 0.5, # Size of text
                    "thickness": 1, # Thickness as is
                    "font": "hershey_simplex" # Text font
                },
                "display_object_id": true # If you want to display object identifier
            }
        },
        {
            "class_name": "motorbike", # see "car" ref.
            "drawing_settings": {} # if propetry is empty, then default values are used
        },
        {
            "class_name": "bus", # see "car" ref.
            "drawing_settings": {} # if propetry is empty, then default values are used
        },
        {
            "class_name": "train", # see "car" ref.
            "drawing_settings": {} # if propetry is empty, then default values are used
        },
        {
            "class_name": "truck", # see "car" ref.
            "drawing_settings": {} # if propetry is empty, then default values are used
        }
    ],
    "tracker_settings": { # Tracked settings
        "tracker_type": "simple/kalman" # Use one of supported trackers. Simple tracker should fit realy simple scenes, while Kalman should be used with complicated scenes.
        "max_points_in_track": 150, # Restriction for maximum points in single track (>=1). Default value 10 (in case of value less than 1)
        "lines_settings":[
            {
                "line_id": 1, # Unique ID for line id (useful for 'client-server' model)
                "begin": [150, 800], # [X1,Y1], start point of line (usually, left side)
                "end": [1600, 800], # [X2,Y2], end point of line (usually, right side)
                "direction": "to_detector", # Direction of line (possible values: 'to_detector' and 'from_detector')
                "detect_classes": ["car", "motorbike", "bus", "train", "truck"], # What classes must be cropped (as detected objects) that were captured by detection line.
                "rgba": [255, 0, 0, 0], # Color of detection line
                "crop_mode": "crop" # When 'grpc_settings' field 'enable' is set to TRUE this option will be used for sending either cropped detected object (bbox==crop) or full image with bbox info to gRPC server-side application. Default is 'crop'
            }
        ],
        "speed_estimation_settings": { # Setting for speed estimation bas on GIS convertion between different spatial systems
            "enabled": false, # Enable this feature or not
            "mapper": [ # Map pixel coordinate to EPSG4326 coordinates
                # You should provide coordinates in correct order.
                # E.g. right bottom -> left bottom -> left top -> right top
                # Coordinates should match reduced_width and reduces_height attributes.
                {"image_coordinates": [640, 360], "epsg4326": [37.61891380882616, 54.20564268115055]},
                {"image_coordinates": [640, 0], "epsg4326": [37.61875545294513, 54.20546281228973]},
                {"image_coordinates": [0, 0], "epsg4326": [37.61903085447736, 54.20543126804313]},
                {"image_coordinates": [0, 360], "epsg4326": [37.61906183714973, 54.20562590237201]}
            ]
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

## Roadmap
Please see [ROADMAP.md](ROADMAP.md)

## Dependencies
* Bindings to [OpenCV](https://github.com/opencv/opencv) - [GoCV](https://github.com/hybridgroup/gocv#gocv). License is Apache-2.0
* MJPEG streaming via GoCV - [mjpeg](https://github.com/hybridgroup/mjpeg). No license currently
* Darknet (AlexeyAB's fork) - [darknet](https://github.com/AlexeyAB/darknet#yolo-v4-and-yolo-v3v2-for-windows-and-linux). License is YOLO LICENSE
* Tracking objects - [gocv-blob](https://github.com/LdDl/gocv-blob#gocv-blob). No license currently
* gRPC for doing "'client-server'" application - [grpc](https://github.com/grpc/grpc-go). License is Apache-2.0

## License
You can check it [here](LICENSE.md)

## Developers

LdDl https://github.com/LdDl

Pavel7824 https://github.com/Pavel7824

Former one: cpllbstr https://github.com/cpllbstr

