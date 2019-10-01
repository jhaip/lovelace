from helper2 import init, claim, retract, prehook, subscription, batch, MY_ID_STR, listen, check_server_connection, get_my_id_str
from graphics import Illumination
import numpy as np
import cv2
import logging

CAM_WIDTH = 1920
CAM_HEIGHT = 1080
projector_calibrations = {}
projection_matrixes = {}
DOTS_CAMERA_ID = 1
LASER_CAMERA_ID = 2
BLANK_SCORES = {"1": 0, "2": 0}
SCORES = BLANK_SCORES
PLAYER_REGIONS = {"1": None, "2": None}
MAX_SCORE = 100
# CAMERA 2 calibration:
# camera 2 has projector calibration TL ( 512 , 282 ) TR ( 1712 , 229 ) BR ( 1788 , 961 ) BL ( 483 , 941 ) @ 2

def project(calibration_id, x, y):
    global projection_matrixes
    x = float(x)
    y = float(y)
    if calibration_id not in projection_matrixes:
        logging.error("MISSING PROJECTION MATRIX FOR CALIBRATION {}".format(calibration_id))
        return (x, y)
    projection_matrix = projection_matrixes[calibration_id]
    pts = [(x, y)]
    dst = cv2.perspectiveTransform(
        np.array([np.float32(pts)]), projection_matrix)
    return (int(dst[0][0][0]), int(dst[0][0][1]))

def point_inside_polygon(x, y, poly):
    # Copied from http://www.ariel.com.au/a/python-point-int-poly.html
    n = len(poly)
    inside =False
    p1x,p1y = poly[0]
    for i in range(n+1):
        p2x,p2y = poly[i % n]
        if y > min(p1y,p2y):
            if y <= max(p1y,p2y):
                if x <= max(p1x,p2x):
                    if p1y != p2y:
                        xinters = (y-p1y)*(p2x-p1x)/(p2y-p1y)+p1x
                    if p1x == p2x or x <= xinters:
                        inside = not inside
        p1x,p1y = p2x,p2y
    return inside

@subscription(["$ $ camera $cameraId has projector calibration TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time"])
def sub_callback_calibration(results):
    global projector_calibrations, projection_matrixes, CAM_WIDTH, CAM_HEIGHT
    logging.info("sub_callback_calibration")
    logging.info(results)
    if results:
        for result in results:
            projector_calibration = [
                [result["x1"], result["y1"]],
                [result["x2"], result["y2"]],
                [result["x4"], result["y4"]],
                [result["x3"], result["y3"]] # notice the order is not clock-wise
            ]
            logging.info(projector_calibration)
            logging.error("RECAL PROJECTION MATRIX")
            pts1 = np.float32(projector_calibration)
            pts2 = np.float32(
                [[0, 0], [CAM_WIDTH, 0], [0, CAM_HEIGHT], [CAM_WIDTH, CAM_HEIGHT]])
            projection_matrix = cv2.getPerspectiveTransform(
                pts1, pts2)
            projector_calibrations[int(result["cameraId"])] = projector_calibration
            projection_matrixes[int(result["cameraId"])] = projection_matrix
            logging.error("RECAL PROJECTION MATRIX -- done")


def draw_score_on_region(result, percentage, r, g, b, subscription_id):
    polygon = [
        project(LASER_CAMERA_ID, result["x1"], result["y1"]),
        project(LASER_CAMERA_ID, result["x2"], result["y2"]),
        project(LASER_CAMERA_ID, result["x2"] + (result["x3"] - result["x2"])*percentage,
                                 result["y2"] + (result["y3"] - result["y2"])*percentage),
        project(LASER_CAMERA_ID, result["x1"] + (result["x4"] - result["x1"])*percentage,
                                 result["y1"] + (result["y4"] - result["y1"])*percentage),
    ]
    ill = Illumination()
    ill.fill(r, g, b, 100)
    ill.nostroke()
    ill.polygon(polygon)
    return ill.to_batch_claim(get_my_id_str(), subscription_id, "global")


def show_scores():
    global PLAYER_REGIONS, SCORES
    claims = []
    claims.append({"type": "retract", "fact": [
        ["id", get_my_id_str()],
        ["id", "1"],
        ["postfix", ""],
    ]})
    for player_id in SCORES:
        if player_id in PLAYER_REGIONS and PLAYER_REGIONS[player_id] is not None:
            r = 255
            g = 0
            b = 0
            if player_id == "2":
                r = 0
                g = 0
                b = 255
            score_percentage = (1.0 * SCORES[player_id]) / (1.0 * MAX_SCORE)
            draw_score_on_region(PLAYER_REGIONS[player_id], score_percentage, r, g, b, "1")
            PLAYER_REGIONS[player_id]
    batch(claims)


@subscription(["$ $ player $playerId scored @ $time"])
def sub_callback_player_scores(results):
    global SCORES, BLANK_SCORES
    if results:
        for result in results:
            result_player_id = str(result["playerId"])
            if result_player_id in SCORES:
                SCORES[result_player_id] += 1
                if SCORES[result_player_id] > MAX_SCORE:
                    SCORES = BLANK_SCORES
        show_scores()


@subscription(["$ $ region $regionId has name $name", "$ $ region $regionId at $x1 $y1 $x2 $y2 $x3 $y3 $x4 $y4"])
def sub_callback_toggle(results):
    global PLAYER_REGIONS
    if results:
        for result in results:
            if result["name"] == "p1health":
                PLAYER_REGIONS["1"] = result
            elif result["name"] == "p2health":
                PLAYER_REGIONS["2"] = result

init(__file__)