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
    * extend gRPC-client to send more attributes (track info)

* vehicle detection
    * detect vehicles
    * crop vehicle near detection line and prepare gRPC structure if needed
    * speed estimation

* added drawing options for tracker (conf.json)
* Check memory leaking
* github tags: godoc, go-report, tagnum, sourcegraph
* integration with go modules
* Integrate Kalman tracker
* Extend configuration of conf.json file.
    * Allow to configure draw methods for each type of detected objects
* Additional field 'targeted objects' (it's called 'detect_classes' actually) in [odam.VirtualLine](virtual_lines.go#11) struct. After it's done odam.VirtualLine will be able to detect e.g. only pedestrians or only motorbikes 

### W.I.P
* design: current BBoxes and text info on imshow()/mjpeg-server are...ugly
* provide video examples to show what this software capable of.
* gRPC
    * optional information about scaling source image
    * optional scaling track in pixel representation
* codebase improvements (design, optimizations, clarifications and etc.)
for example.
* Tracking in convex polygon: <<=== Current state (30.08.2021) Almost Done with convex polygons 
    * estimated time spent in polygon
    * estimated speed (via GIS 'mapper' technique)
    * objects filtering (same as with VirtualLine)
    * integrate into gRPC
    * convex polygons math
    * JSON configuration
    * store information about visited polygons somewhere
    * draw polygons
* Virtual lines
   * Split ID and other additional information for next paragraph
   * Make additional information optional for sending via gRPC. Sometimes reciever-side already knows everything about lines and just need its IDs.
* Move to full OpenCV (no [go-darknet](https://github.com/LdDl/go-darknet) is needed since OpenCV does stuff)

### Planned
* Stable core (need many tests as possible)
* Extend [conf.json](cmd/odam/conf.json) for such settings as: color of virtual lines, color of boxes and similar stuff.
* Front-end for editing [conf.json](cmd/odam/conf.json)
* Exend gRPC-client set of attributes, which must be send to gRPC-server
* Some kind of wiki
* Logo
* Contributing guidelines
* pedestrian detection
    * detect pedestrians
    * count pedestrians
    * speed estimation
* Implement SORT - https://arxiv.org/abs/1602.00763
* github tags: travis
* gRPC server-side for mutation and querying reference info
* REST server-side for mutation and querying reference info (may be by code wrapping gRPC-based code?)

### Continuous activity
* README
* Memory profiling
* [ODaM](cmd/odam) itself
* Roadmap itself
* conf.json features
