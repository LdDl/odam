wget --output-document=coco.names https://raw.githubusercontent.com/AlexeyAB/darknet/master/data/coco.names
wget --output-document=yolov4.cfg https://raw.githubusercontent.com/AlexeyAB/darknet/master/cfg/yolov4.cfg
sed -i 's/batch=64/batch=1/g' yolov4-tiny.cfg
wget --output-document=yolov4.weights https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v3_optimal/yolov4.weights