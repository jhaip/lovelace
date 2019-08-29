from helper2 import init, subscription, batch, MY_ID_STR, check_server_connection, get_my_id_str
from imutils.video import WebcamVideoStream
import numpy as np
import cv2
import imutils
import os
import time

DEBUG = True

capture = WebcamVideoStream(src=0)
CAM_WIDTH = 1920
CAM_HEIGHT = 1080
capture.stream.set(cv2.CAP_PROP_FRAME_WIDTH, CAM_WIDTH)
capture.stream.set(cv2.CAP_PROP_FRAME_HEIGHT, CAM_HEIGHT)
time.sleep(2)
capture.start()
time.sleep(2)

PERSPECTIVE_CALIBRATION = np.array([
        (442, 0),
        (1488, 8),
        (1629, 1045),
        (288, 1041)], dtype = "float32")
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
CELL_WIDTH_PX = 174
CELL_HEIGHT_PX = 150
ORIGIN_X = 0
ORIGIN_Y = 0
CELL_X_PADDING_PX = 77
CELL_Y_PADDING_PX = 138
NUMBER_CELL_OFFSET_X_PX = 0
NUMBER_CELL_OFFSET_Y_PX = CELL_HEIGHT_PX + 10
NUMBER_CELL_WIDTH_PX = CELL_WIDTH_PX
NUMBER_CELL_HEIGHT_PX = 60
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
    threshold_image_arr = []
    threshold_number_image_arr = []
    erosion_number_image_arr = []
    contour_image_arr = []
    contour_number_image_arr = []
    final_image_arr = []
    final_number_image_arr = []
    tile_data = []
    KERNEL_SIZE = 7
    DILATIONS = 3
    kernel = np.ones((KERNEL_SIZE, KERNEL_SIZE), np.uint8)
    for ix in range(GRID_WIDTH_CELLS):
        for iy in range(GRID_HEIGHT_CELLS):
            x = ORIGIN_X + ix * (CELL_WIDTH_PX + CELL_X_PADDING_PX)
            y = ORIGIN_Y + iy * (CELL_HEIGHT_PX + CELL_Y_PADDING_PX)
            x2 = x + CELL_WIDTH_PX
            y2 = y + CELL_HEIGHT_PX
            if DEBUG:
                cv2.rectangle(warped, (x, y), (x2, y2), (0, 255, 0), 3)
            roi = warped_grey[y:y2, x:x2]
            color_roi = cv2.cvtColor(roi, cv2.COLOR_GRAY2BGR)
            # cv2.imshow("ROI", roi)
            threshold = np.median(roi) * 1.2 # threshold the roi a little bit above the median
            ret, threshold_image = cv2.threshold(roi, threshold, 255, cv2.THRESH_BINARY)
            threshold_image_arr.append(threshold_image)
            
            dilation = cv2.dilate(threshold_image, kernel, iterations=DILATIONS)
            cimg, contours, hierarchy = cv2.findContours(dilation, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
            biggest_contours = sorted(contours, key=cv2.contourArea, reverse=True)[:1]
            if len(biggest_contours) > 0:
                if cv2.contourArea(biggest_contours[0]) < 4000:
                    biggest_contours = []
                else:
                    rect = cv2.boundingRect(biggest_contours[0])
                    symmetry = min(rect[2], rect[3]) / max(rect[2], rect[3])
                    print(symmetry)
                    if symmetry < 0.75:
                        biggest_contours = []
            
            # print(contours)
            color_roi = cv2.drawContours(color_roi, biggest_contours, -1, (255, 0, 0), 3)
            contour_image_arr.append(color_roi)
            
            mask = np.zeros(roi.shape, np.uint8)
            cv2.drawContours(mask, biggest_contours, -1, (255), -1)
            masked_roi = cv2.bitwise_and(threshold_image, mask)
            if len(biggest_contours) > 0:
                cx,cy,cw,ch = cv2.boundingRect(biggest_contours[0])
                kernel_offset = (KERNEL_SIZE // 2) * DILATIONS
                cx += kernel_offset
                cy += kernel_offset
                cw -= kernel_offset*2
                ch -= kernel_offset*2
                final_image = masked_roi[cy:(cy+ch), cx:(cx+cw)]
                final_image = cv2.resize(final_image, (100, 100), interpolation=cv2.INTER_NEAREST)
                final_image_arr.append(final_image)
            
            tile_identifcation = {} # identify_tile(threshold_image)
            tile_identifcation["x"] = ix
            tile_identifcation["y"] = iy
            tile_data.append(tile_identifcation)
    for ix in range(GRID_WIDTH_CELLS):
        for iy in range(GRID_HEIGHT_CELLS):
            x = ORIGIN_X + ix * (CELL_WIDTH_PX + CELL_X_PADDING_PX) + NUMBER_CELL_OFFSET_X_PX
            y = ORIGIN_Y + iy * (CELL_HEIGHT_PX + CELL_Y_PADDING_PX) + NUMBER_CELL_OFFSET_Y_PX
            x2 = x + NUMBER_CELL_WIDTH_PX
            y2 = y + NUMBER_CELL_HEIGHT_PX
            if DEBUG:
                cv2.rectangle(warped, (x, y), (x2, y2), (0, 0, 255), 3)
            roi = warped_grey[y:y2, x:x2]
            color_roi = cv2.cvtColor(roi, cv2.COLOR_GRAY2BGR)
            # cv2.imshow("ROI", roi)
            threshold = np.median(roi) * 1.2 # threshold the roi a little bit above the median
            ret, threshold_image = cv2.threshold(roi, threshold, 255, cv2.THRESH_BINARY)
            threshold_number_image_arr.append(threshold_image)

            erode_kernel = np.ones((3, 3), np.uint8)
            erosion = cv2.erode(threshold_image, erode_kernel, iterations=1)
            erosion_number_image_arr.append(erosion)
            
            cimg, contours, hierarchy = cv2.findContours(threshold_image, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
            biggest_contours = sorted(contours, key=cv2.contourArea, reverse=True)[:1]
            # print(contours)
            color_roi = cv2.drawContours(color_roi, biggest_contours, -1, (255, 0, 0), 3)
            contour_number_image_arr.append(color_roi)
            
            mask = np.zeros(roi.shape, np.uint8)
            cv2.drawContours(mask, biggest_contours, -1, (255), -1)
            masked_roi = cv2.bitwise_and(threshold_image, mask)
            if len(biggest_contours) > 0:
                cx,cy,cw,ch = cv2.boundingRect(biggest_contours[0])
                final_image = masked_roi[cy:(cy+ch), cx:(cx+cw)]
                final_image = cv2.resize(final_image, (100, 100), interpolation=cv2.INTER_NEAREST)
                final_number_image_arr.append(final_image)
            # tile_identifcation = {} # identify_tile(threshold_image)
            # tile_identifcation["x"] = ix
            # tile_identifcation["y"] = iy
            # tile_data.append(tile_identifcation)
    if DEBUG:
        cv2.imshow("Original", image)
        cv2.imshow("Warped", warped)
        tiles = np.concatenate(threshold_image_arr, axis=1)
        tiles = imutils.resize(tiles, width=1600)
        cv2.imshow("tiles", tiles)
        contour_tiles = np.concatenate(contour_image_arr, axis=1)
        contour_tiles = imutils.resize(contour_tiles, width=1600)
        cv2.imshow("contour tiles", contour_tiles)
        final_tiles = np.concatenate(final_image_arr, axis=1)
        final_tiles = imutils.resize(final_tiles, width=1600)
        cv2.imshow("final tiles", final_tiles)
        threshold_number_tiles = np.concatenate(threshold_number_image_arr, axis=1)
        threshold_number_tiles = imutils.resize(threshold_number_tiles, width=1600)
        cv2.imshow("number theshold tiles", threshold_number_tiles)
        erosion_number_tiles = np.concatenate(erosion_number_image_arr, axis=1)
        erosion_number_tiles = imutils.resize(erosion_number_tiles, width=1600)
        cv2.imshow("number erosion tiles", erosion_number_tiles)
        cv2.imwrite('numbers.png', cv2.cvtColor(cv2.bitwise_not(erosion_number_image_arr[18]), cv2.COLOR_GRAY2BGR))

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

detect()
cv2.waitKey(0)
# init(__file__)
"""
init(__file__, skipListening=True)
while True:
    tile_data = detect()
    claim_tile_data(tile_data)
    time.sleep(1)
"""