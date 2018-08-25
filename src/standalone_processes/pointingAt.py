import requests
import time
import sys
import os
import logging
import math
import json

scriptName = os.path.basename(__file__)
scriptNameNoExtension = os.path.splitext(scriptName)[0]
fileDir = os.path.dirname(os.path.realpath(__file__))
logPath = os.path.join(fileDir, 'logs/' + scriptNameNoExtension + '.log')
print(logPath)

logging.basicConfig(filename=logPath, level=logging.INFO)

URL = "http://localhost:3000/"

def say(fact):
    payload = {'facts': fact}
    return requests.post(URL + "assert", data=payload)

def retract(fact):
    payload = {'facts': fact}
    return requests.post(URL + "retract", data=payload)

def select(fact):
    payload = {'facts': fact}
    response = requests.post(URL + "select", data=payload)
    return response.json()

###########

def get_paper_center(c):
    x = (c[0]["x"] + c[1]["x"] + c[2]["x"] + c[3]["x"]) * 0.25
    y = (c[0]["y"] + c[1]["y"] + c[2]["y"] + c[3]["y"]) * 0.25
    return {"x": x, "y": y}

def move_along_vector(amount, vector):
    size = math.sqrt(vector["x"]**2 + vector["y"]**2)
    C = 1.0
    if size != 0:
        C = 1.0 * amount / size
    return {"x": C * vector["x"], "y": C * vector["y"]}

def add_vec(vec1, vec2):
    return {"x": vec1["x"] + vec2["x"], "y": vec1["y"] + vec2["y"]}

def diff_vec(vec1, vec2):
    return {"x": vec1["x"] - vec2["x"], "y": vec1["y"] - vec2["y"]}

def scale_vec(vec, scale):
    return {"x": vec["x"] * scale, "y": vec["y"] * scale}

def get_paper_wisker(corners, direction, length):
    center = get_paper_center(corners)
    segment = None
    if direction == 'right':
        segment = (corners[1], corners[2])
    elif direction == 'down':
        segment = (corners[2], corners[3])
    elif direction == 'left':
        segment = (corners[3], corners[0])
    else:
        segment = (corners[0], corners[1])
    segmentMiddle = add_vec(segment[1], scale_vec(diff_vec(segment[0], segment[1]), 0.5))
    wiskerEnd = add_vec(segmentMiddle, move_along_vector(length, diff_vec(segmentMiddle, center)))
    return (segmentMiddle, wiskerEnd)

# Adapted from https://stackoverflow.com/questions/9043805/test-if-two-lines-intersect-javascript-function
def intersects(v1,  v2,  v3,  v4):
    det = (v2["x"] - v1["x"]) * (v4["y"] - v3["y"]) - (v4["x"] - v3["x"]) * (v2["y"] - v1["y"])
    if det == 0:
        return False
    else:
        _lambda = ((v4["y"] - v3["y"]) * (v4["x"] - v1["x"]) + (v3["x"] - v4["x"]) * (v4["y"] - v1["y"])) / det
        gamma = ((v1["y"] - v2["y"]) * (v4["x"] - v1["x"]) + (v2["x"] - v1["x"]) * (v4["y"] - v1["y"])) / det
        return (0 < _lambda and _lambda < 1) and (0 < gamma and gamma < 1)

def get_paper_you_point_at(papers, you_id, WISKER_LENGTH):
    valid_papers = []
    my_paper = None
    for paper in papers:
        if len(paper["corners"]) == 4:
            if str(paper["id"]) == str(you_id):
                my_paper = paper["corners"]
            else:
                valid_papers.append(paper)
    if my_paper is not None and len(valid_papers) > 0:
        wisker = get_paper_wisker(my_paper, "up", WISKER_LENGTH)
        for paper in valid_papers:
            corners = paper["corners"]
            if intersects(wisker[0], wisker[1], corners[0], corners[1]) or \
               intersects(wisker[0], wisker[1], corners[1], corners[2]) or \
               intersects(wisker[0], wisker[1], corners[2], corners[3]) or \
               intersects(wisker[0], wisker[1], corners[3], corners[0]):
                return paper["id"]
    return None

#####

while True:
    result = select('camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time')
    if not result:
        continue
    papers = list(map(lambda p: ({
        "id": p["id"]["value"],
        "corners": [
            {"x": p["x1"]["value"], "y": p["y1"]["value"]},
            {"x": p["x2"]["value"], "y": p["y2"]["value"]},
            {"x": p["x3"]["value"], "y": p["y3"]["value"]},
            {"x": p["x4"]["value"], "y": p["y4"]["value"]}
        ]
    }), result))
    logging.info("--")

    WISKER_LENGTH = 150
    retract("paper $ is pointing at paper $")
    for paper in papers:
        other_paper = get_paper_you_point_at(papers, paper["id"], WISKER_LENGTH)
        logging.info("{} pointing at {}".format(paper["id"], other_paper))
        if other_paper is not None:
            say("paper {} is pointing at paper {}".format(paper["id"], other_paper))

    time.sleep(1)
