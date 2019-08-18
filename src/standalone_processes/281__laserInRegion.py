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


@subscription(["$ $ laser seen at $x $y @ $t", "$ $ region $regionId at $x1 $y1 $x2 $y2 $x3 $y3 $x4 $y4"])
def sub_callback_laser_dots(results):
    claims = []
    claims.append({"type": "retract", "fact": [
        ["id", get_my_id_str()],
        ["id", "1"],
        ["postfix", ""],
    ]})
    for result in results:
        polygon = [
            project(LASER_CAMERA_ID, result["x1"], result["y1"]),
            project(LASER_CAMERA_ID, result["x2"], result["y2"]),
            project(LASER_CAMERA_ID, result["x3"], result["y3"]),
            project(LASER_CAMERA_ID, result["x4"], result["y4"]),
        ]
        laser_point = project(LASER_CAMERA_ID, result["x"], result["y"])
        inside = point_inside_polygon(laser_point[0], laser_point[1], polygon)
        if inside:
            logging.info("laser {} {} inside region {}".format(result["x"], result["y"], result["regionId"]))
            claims.append({"type": "claim", "fact": [
                ["id", get_my_id_str()],
                ["id", "1"],
                ["text", "laser"],
                ["text", "in"],
                ["text", "region"],
                ["text", str(result["regionId"])],
            ]})
        # else:
        #     logging.info("paper {} is not inside laser {} {}".format(result["paper"], result["x"], result["y"]))
    batch(claims)

init(__file__)