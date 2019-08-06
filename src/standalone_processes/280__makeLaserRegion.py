from helper2 import init, claim, retract, prehook, subscription, batch, MY_ID_STR, listen, check_server_connection, get_my_id_str
from graphics import Illumination
import numpy as np
import cv2
import uuid
import logging

MODE = "IDLE"
lastLastPosition = None
regionPoints = [None, None, None, None]
CAM_WIDTH = 1920
CAM_HEIGHT = 1080
projector_calibrations = {}
projection_matrixes = {}
LASER_CAMERA_ID = 2
# CAMERA 2 calibration:
# camera 2 has projector calibration TL(512, 282) TR(1712, 229) BR(1788, 961) BL(483, 941) @2

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
                [result["x3"], result["y3"]]# notice the order is not clock - wise
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


@subscription(["keyboard $ typed key \"1\" @ $t"])
def sub_callback_keyboard(results):
    global MODE, lastLastPosition, regionPoints
    if results:
        if MODE == "IDLE":
            MODE = "0"
        elif MODE == "0" && lastLastPosition != None:
            MODE = "1"
            regionPoints[0] = lastLastPosition
        elif MODE == "1" && lastLastPosition != None:
            MODE = "2"
            regionPoints[1] = lastLastPosition
        elif MODE == "2" && lastLastPosition != None:
            MODE = "3"
            regionPoints[2] = lastLastPosition
        elif MODE == "3" && lastLastPosition != None:
            MODE = "IDLE"
            regionPoints[3] = lastLastPosition
            claims = []
            claims.append({"type": "claim", "fact": [
                ["id", get_my_id_str()],
                ["id", "1"],
                ["text", "region"],
                ["text", str(uuid.uuid4())],
                ["text", "at"],
                ["integer", str(regionPoints[0][0])],
                ["integer", str(regionPoints[0][1])],
                ["integer", str(regionPoints[1][0])],
                ["integer", str(regionPoints[1][1])],
                ["integer", str(regionPoints[2][0])],
                ["integer", str(regionPoints[2][1])],
                ["integer", str(regionPoints[3][0])],
                ["integer", str(regionPoints[3][1])],
            ]})
            batch(claims)
            regionPoints = [None, None, None, None]
        logging.info("MODE: {}, region points: {}".format(MODE, regionPoints))


@subscription(["$ $ laser seen at $x $y @ $t"])
def sub_callback_laser_dots(results):
    global lastLastPosition, MODE
    claims = []
    claims.append({
        "type": "retract", "fact": [
            ["id", get_my_id_str()],
            ["id", "2"],
            ["postfix", ""],
        ]
    })
    if results and len(results) > 0:
        result = results[0]
        lastLastPosition = [result["x"], result["y"]]
        if MODE != "IDLE":
            ill = Illumination()
            ill.stroke(255, 0, 255, 128)
            ill.fill(255, 0, 255, 100)
            current_corner = int(MODE)
            poly = regionPoints[:current_corner] + [lastLastPosition]
            projected_poly = list(map(lambda p: project(LASER_CAMERA_ID, p[0], [1]), poly))
            ill.polygon(projected_poly)
            SIZE = 5
            for pt in projected_poly:
                ill.ellipse(pt[0] - SIZE, pt[1] - SIZE, SIZE * 2, SIZE * 2)
            claims.append(ill.to_batch_claim(get_my_id_str(), "2", "global"))
    else:
        lastLastPosition = None
    batch(claims)

init(__file__)