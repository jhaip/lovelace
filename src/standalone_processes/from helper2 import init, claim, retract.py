from helper2 import init, claim, retract, prehook, subscription, batch, MY_ID_STR, listen, check_server_connection, get_my_id_str
import numpy as np
import cv2
import logging

# our choice of the size of the size of the calendar in display px
CALENDAR_WIDTH_IN_DISPLAY_PIXELS = 1280
CALENDAR_HEIGHT_IN_DISPLAY_PIXELS = 720
projection_matrixes = {}

def project(projection_matrix, x, y):
    pts = (float(x), float(y))
    if not projection_matrix:
        return pts
    dst = cv2.perspectiveTransform(
        np.array([np.float32([pts])]), projection_matrix)
    return (int(dst[0][0][0]), int(dst[0][0][1]))

@subscription(["$ $ camera $cameraId calibration for $display is $M1 $M2 $M3 $M4 $M5 $M6 $M7 $M8 $M9"])
def sub_callback_calibration_points(results):
    global projection_matrixes
    if results:
        for result in results:
            projection_matrixes[str(result["cameraId"])] = np.float32([
                [float(result["M1"]), float(result["M2"]), float(result["M3"])],
                [float(result["M4"]), float(result["M5"]), float(result["M6"])],
                [float(result["M7"]), float(result["M8"]), float(result["M9"])]])

@subscription(["$ $ region $id at $rx1 $ry1 $rx2 $ry2 $rx3 $ry3 $rx4 $ry4 on camera $cameraId",
               "$ $ region $id has name calendar"])
def sub_callback_calibration_points(results):
    global projection_matrixes
    claims = [{
        "type": "retract", "fact": [["id", get_my_id_str()], ["id", "0"], ["postfix", ""]]
    }]
    if results and len(results) > 0:
        for result in results:
            src = np.float32([
                [0, 0],
                [CALENDAR_WIDTH_IN_DISPLAY_PIXELS, 0],
                [0, CALENDAR_HEIGHT_IN_DISPLAY_PIXELS],
                [CALENDAR_WIDTH_IN_DISPLAY_PIXELS, CALENDAR_HEIGHT_IN_DISPLAY_PIXELS] # notice the order is not clock-wise
            ])
            calendar_camera_homography_matrix = projection_matrixes.get(str(result["cameraId"]))
            dst = np.float32([
                project(calendar_camera_homography_matrix, result["rx1"], result["ry1"]),
                project(calendar_camera_homography_matrix, result["rx2"], result["ry2"]),
                project(calendar_camera_homography_matrix, result["rx4"], result["ry4"]),
                project(calendar_camera_homography_matrix, result["rx3"], result["ry3"]) # notice the order is not clock-wise
            ])
            homography_matrix = cv2.getPerspectiveTransform(src, dst)
            claims.append({"type": "claim", "fact": [
                ["id", get_my_id_str()],
                ["id", "0"],
                ["text", "camera"],
                ["text", str(result["cam"])],
                ["text", "calibration"],
                ["text", "for"],
                ["text", str(result["display"])],
                ["text", "is"],
                ["float", str(homography_matrix[0][0])],
                ["float", str(homography_matrix[0][1])],
                ["float", str(homography_matrix[0][2])],
                ["float", str(homography_matrix[1][0])],
                ["float", str(homography_matrix[1][1])],
                ["float", str(homography_matrix[1][2])],
                ["float", str(homography_matrix[2][0])],
                ["float", str(homography_matrix[2][1])],
                ["float", str(homography_matrix[2][2])],
            ]})            
    batch(claims)


init(__file__)
