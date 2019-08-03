from helper2 import init, subscription, batch, MY_ID_STR, check_server_connection, get_my_id_str
from imutils.video import WebcamVideoStream
import numpy as np
import cv2
import imutils
import os
import time

capture = WebcamVideoStream(src=1)
CAM_WIDTH = 640
CAM_HEIGHT = 480
capture.stream.set(cv2.CAP_PROP_FRAME_WIDTH, CAM_WIDTH)
capture.stream.set(cv2.CAP_PROP_FRAME_HEIGHT, CAM_HEIGHT)
capture.stream.set(cv2.CAP_PROP_GAIN, 1)
capture.stream.set(cv2.CAP_PROP_EXPOSURE, 1)
capture.stream.set(cv2.CAP_PROP_BRIGHTNESS, 1)
time.sleep(2)
capture.start()
time.sleep(2)

def detect():
    image = capture.read()
    image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    # image = imutils.resize(image, width=400)
    cv2.imshow("Original", image)
    threshold = 200
    ret, threshold_image = cv2.threshold(image, threshold, 255, cv2.THRESH_BINARY)
    cv2.imshow("Threshold", threshold_image)
    # TODO
    

def claim_tile_data(tile_data):
    currentTimeMs = int(round(time.time() * 1000))
    claims = [
        {"type": "retract", "fact": [["id", get_my_id_str()], ["id", "0"], ["postfix", ""]]},
        {"type": "retract", "fact": [["id", ""], ["id", ""],
            ["text", "wish"], ["text", "new"], ["text", "tiles"],
            ["text", "would"], ["text", "be"], ["text", "seen"]]}
    ]
    for datum in tile_data:
        if datum["tile"] != "":
            claims.append({"type": "claim", "fact": [
                ["id", get_my_id_str()],
                ["id", "0"],
                ["text", "tile"],
                ["text", datum["tile"]],
                ["text", "seen"],
                ["text", "at"],
                ["integer", str(datum["x"])],
                ["integer", str(datum["y"])],
                ["text", "@"],
                ["integer", str(currentTimeMs)]
            ]})
        else:
            claims.append({"type": "claim", "fact": [
                ["id", get_my_id_str()],
                ["id", "0"],
                ["text", "tile"],
                ["text", datum["best_tile"]],
                ["text", "maybe"],
                ["float", str(datum["score"])],
                ["text", "seen"],
                ["text", "at"],
                ["integer", str(datum["x"])],
                ["integer", str(datum["y"])],
                ["text", "@"],
                ["integer", str(currentTimeMs)]
            ]})
    batch(claims)


# @subscription(["$ $ wish new tiles would be seen"])
# def sub_callback(results):
#     if not results:
#         return
#     tile_data = detect()
#     claim_tile_data(tile_data)

# cv2.waitKey(0)
# init(__file__)
# init(__file__, skipListening=True)
tile_data = detect()
cv2.waitKey(0)
# while True:
#     tile_data = detect()
#     claim_tile_data(tile_data)
#     time.sleep(1)
