from helper2 import init, subscription, batch, MY_ID_STR, check_server_connection, get_my_id_str
from imutils.video import WebcamVideoStream
import numpy as np
import cv2
import imutils
import os
import time

capture = WebcamVideoStream(src=0)
CAM_WIDTH = 1920
CAM_HEIGHT = 1080
capture.stream.set(cv2.CAP_PROP_FRAME_WIDTH, CAM_WIDTH)
capture.stream.set(cv2.CAP_PROP_FRAME_HEIGHT, CAM_HEIGHT)
time.sleep(2)
capture.start()
time.sleep(2)

PERSPECTIVE_CALIBRATION = np.array([
        (80*4.8, 4*4.8),
        (305*4.8, 1*4.8),
        (344*4.8, 224*4.8),
        (46*4.8, 225*4.8)], dtype = "float32")
PERSPECTIVE_IMAGE_WIDTH = 1430
PERSPECTIVE_IMAGE_HEIGHT = 1086
dst = np.array([
        [0, 0],
        [PERSPECTIVE_IMAGE_WIDTH - 1, 0],
        [PERSPECTIVE_IMAGE_WIDTH - 1, PERSPECTIVE_IMAGE_HEIGHT - 1],
        [0, PERSPECTIVE_IMAGE_HEIGHT - 1]], dtype = "float32")
PERSPECTIVE_MATRIX = cv2.getPerspectiveTransform(PERSPECTIVE_CALIBRATION, dst)
GRID_WIDTH_CELLS = 6
GRID_HEIGHT_CELLS = 4
CELL_WIDTH_PX = 155
CELL_HEIGHT_PX = 135
ORIGIN_X = 25
ORIGIN_Y = 10
CELL_X_PADDING_PX = 90
CELL_Y_PADDING_PX = 148
SAMPLE_IMAGE_MATCH_THRESHOLD = 0.8  # Value 0 (0% match) to 1.0 (100% match)
TILES = ["up", "down", "left", "right", "loopstart", "loopstop"]
SAMPLE_IMAGES = []
for name in TILES:
    sample_image = cv2.imread(os.path.join(os.path.dirname(__file__), 'files/cv_tiles/{}.png'.format(name)))
    sample_image = cv2.cvtColor(sample_image, cv2.COLOR_BGR2GRAY)
    SAMPLE_IMAGES.append(sample_image)

def identify_tile(image):
    best_score = None
    best_sample = ""
    for i in range(len(TILES)):
        sample_image_name = TILES[i]
        sample_image = SAMPLE_IMAGES[i]
        # Use the count of while pixels in the XOR of the given image and the sample image
        # as a measure for how difference the images are
        xor_image = cv2.bitwise_xor(sample_image, image)
        xor_sum = np.sum(xor_image == 255)
        # cv2.imshow(same_image_name, xor_image)
        percentage_correct = 1.0 - float(xor_sum) / float(CELL_WIDTH_PX * CELL_HEIGHT_PX)
        if best_score is None or percentage_correct > best_score:
            best_score = percentage_correct
            best_sample = sample_image_name
    output = ""
    if best_score > SAMPLE_IMAGE_MATCH_THRESHOLD:
        output = best_sample
    return {"tile": output, "best_tile": best_sample, "score": best_score}

def detect():
    image = capture.read()
    # image = imutils.resize(image, width=400)
    warped = cv2.warpPerspective(image, PERSPECTIVE_MATRIX, (PERSPECTIVE_IMAGE_WIDTH, PERSPECTIVE_IMAGE_HEIGHT))
    warped = cv2.flip( warped, -1 )  # flip both axes
    warped_grey = cv2.cvtColor(warped, cv2.COLOR_BGR2GRAY)
    # cv2.imshow("Original", image)
    # cv2.imshow("Warped", warped)
    threshold_image_arr = []
    tile_data = []
    for ix in range(GRID_WIDTH_CELLS):
        for iy in range(GRID_HEIGHT_CELLS):
            x = ORIGIN_X + ix * (CELL_WIDTH_PX + CELL_X_PADDING_PX)
            y = ORIGIN_Y + iy * (CELL_HEIGHT_PX + CELL_Y_PADDING_PX)
            x2 = x + CELL_WIDTH_PX
            y2 = y + CELL_HEIGHT_PX
            cv2.rectangle(warped, (x, y), (x2, y2), (0, 255, 0), 3)
            roi = warped_grey[y:y2, x:x2]
            # cv2.imshow("ROI", roi)
            th2 = cv2.adaptiveThreshold(roi, 255, cv2.ADAPTIVE_THRESH_MEAN_C, \
                cv2.THRESH_BINARY,CELL_WIDTH_PX + CELL_HEIGHT_PX+1, 2)
            th4 = cv2.adaptiveThreshold(roi, 255, cv2.ADAPTIVE_THRESH_MEAN_C, \
                cv2.THRESH_BINARY, 11, 2)
            threshold = np.median(roi) * 1.2 # threshold the roi a little bit above the median
            ret, threshold_image = cv2.threshold(roi, threshold, 255, cv2.THRESH_BINARY)
            threshold_image_arr.append(threshold_image)
            tile_identifcation = identify_tile(threshold_image)
            tile_identifcation["x"] = ix
            tile_identifcation["y"] = iy
            tile_data.append(tile_identifcation)
    # Show images for debugging:
    # tiles = np.concatenate(threshold_arr, axis=1)
    # tiles = imutils.resize(tiles, width=1000)
    # cv2.imshow("tiles", tiles)
    # cv2.imwrite('left.png', threshold_arr[0])
    # cv2.imwrite('right.png', threshold_arr[4])
    # cv2.imwrite('down.png', threshold_arr[8])
    # cv2.imwrite('up.png', threshold_arr[13])
    # cv2.imwrite('loopstart.png', threshold_arr[17])
    # cv2.imwrite('loopstop.png', threshold_arr[20])
    return tile_data

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
init(__file__, skipListening=True)
while True:
    tile_data = detect()
    claim_tile_data(tile_data)
    time.sleep(1)
