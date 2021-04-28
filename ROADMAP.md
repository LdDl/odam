## ROADMAP
New ideas, thoughts about needed features will be store in this file.

### Done
* Initial core
    * Integration with [go-darknet](https://github.com/LdDl/go-darknet)
    * Initial integration with [gRPC](https://grpc.io/docs/quickstart/go/)
    * Initial integration with [GoCV](https://github.com/hybridgroup/gocv/)
    * Initial integration with [GoCV MJPEG](https://github.com/hybridgroup/mjpeg)

* go-darknet
    * convert [gocv.Mat](https://github.com/hybridgroup/gocv/blob/master/core.go#L179) to [darknet.DarknetImage](https://github.com/LdDl/go-darknet/blob/master/image.go#L14)
    * init neural network from configuration
    * prepare *.sh scripts to download yolov4.cfg and yolov4.weights files (also yolov3 avaible)
    * detect only targeted classes

* GoCV
    * init gocv.VideoCapture
    * make separate goroutines for grabbing frames and feeding them to neural network
    * make MJPEG server avaible as option

* gRPC
    * inital gRPC-client from https://github.com/LdDl/license_plate_recognition
    * prepare gRPC-client structure
    * create "sending" function
    * make gRPC-client server avaible as option

* vehicle detection
    * detect vehicles
    * crop vehicle near detection line and prepare gRPC structure if needed
    * speed estimation

* added drawing options for tracker (conf.json)
* Check memory leaking
* github tags: godoc, go-report, tagnum, sourcegraph
* integration with go modules
* Integrate Kalman tracker

### W.I.P
* Extend configuration of conf.json file.
* design: current BBoxes and text info on imshow()/mjpeg-server are...ugly
* gRPC
    * extend gRPC-client to send more attributes (track info)

### Planned
* Stable core (need many tests as possible)
* Extend [conf.json](cmd/odam/conf.json) for such settings as: color of virtual lines, color of boxes and similar stuff.
* Front-end for editing [conf.json](cmd/odam/conf.json)
* Exend gRPC-client set of attributes, which must be send to gRPC-server
* Some kind of wiki
* Logo
* Contributing guidelines
* Additional field 'targeted objects' in [odam.VirtualLine](virtual_lines.go#11) struct. After it's done odam.VirtualLine will be able to detect only pedestrians or only motorbikes for example.
* pedestrian detection
    * detect pedestrians
    * count pedestrians
    * speed estimation
* github tags: travis

### Continuous activity
* README
* Memory profiling
* [ODaM](cmd/odam) itself
* Roadmap itself
* conf.json features
