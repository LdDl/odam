# ODaM - Object Detection and Monitoring
# v0.1.0
ODaM is project aimed to do monitoring such as: pedestrian detection and counting, vehicle detection and counting, speed estimation of objects, sending detected objects to gRPC server for detailed analysis.

It's written on Go with a lot of [CGO](https://golang.org/cmd/cgo/).

# Work in progress
# DO NOT USE IT IN PRODUCTION UNTIL STABLE CORE

## Table of Contents
- [About](#about)
- [Installation](#installation)
- [Usage](#usage)
- [Support](#support)
- [ToDo](#todo)
- [Dependencies](#dependencies)
- [License](#license)

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

## Usage
### notice: targeted for Linux users (no Windows/OSX instructions currenlty)
@todo

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
* gRPC for doing "client-server" application - [grpc](https://github.com/grpc/grpc-go). License is Apache-2.0

## License
You can check it [here](LICENSE.md)