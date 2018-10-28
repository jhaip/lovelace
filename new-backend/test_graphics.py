from client_helper import init, claim, retract, prehook, subscription, batch, MY_ID_STR
import sys
import time
MY_ID = sys.argv[1]
print(MY_ID)

@prehook
def my_prehook():
    global MY_ID_STR
    # "$ camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time"
    claims = []
    paper_drawing_target_id = "1234"
    paper_width = "400.000000"
    paper_height = "400.000000"
    claims.append({"type": "retract", "fact": [
        ["source", MY_ID_STR],
        ["postfix", ""],
    ]})
    claims.append({"type": "claim", "fact": [
                    ["source", MY_ID_STR],
                    ["text", "camera"],
                    ["integer", "1"],
                    ["text", "sees"],
                    ["text", "paper"],
                    ["integer", paper_drawing_target_id],
                    ["text", "at"],
                    ["text", "TL"],
                    ["text", "("],
                    ["float", "0.000000"],
                    ["text", ","],
                    ["float", "0.000000"],
                    ["text", ")"],
                    ["text", "TR"],
                    ["text", "("],
                    ["float", paper_width],
                    ["text", ","],
                    ["float", "0.000000"],
                    ["text", ")"],
                    ["text", "BR"],
                    ["text", "("],
                    ["float", paper_width],
                    ["text", ","],
                    ["float", paper_height],
                    ["text", ")"],
                    ["text", "BL"],
                    ["text", "("],
                    ["float", "0.000000"],
                    ["text", ","],
                    ["float", paper_height],
                    ["text", ")"],
                    ["text", "@"],
                    ["integer", "999"]]})
    # claims.append({"type": "claim", "fact": [
    #     ["source", MY_ID_STR],
    #     ["text", "camera"],
    #     ["integer", "1"],
    #     ["text", "has"],
    #     ["text", "projector"],
    #     ["text", "calibration"],
    #     ["text", "TL"],
    #     ["text", "("],
    #     ["float", "1036.000000"],
    #     ["text", ","],
    #     ["float", "77.000000"],
    #     ["text", ")"],
    #     ["text", "TR"],
    #     ["text", "("],
    #     ["float", "1036.000000"],
    #     ["text", ","],
    #     ["float", "77.000000"],
    #     ["text", ")"],
    #     ["text", "BR"],
    #     ["text", "("],
    #     ["float", "1036.000000"],
    #     ["text", ","],
    #     ["float", "77.000000"],
    #     ["text", ")"],
    #     ["text", "BL"],
    #     ["text", "("],
    #     ["float", "1036.000000"],
    #     ["text", ","],
    #     ["float", "77.000000"],
    #     ["text", ")"],
    #     ["text", "@"],
    #     ["integer", "999"]]})
    claims.append({"type": "claim", "fact": [
        ["source", paper_drawing_target_id],
        ["text", "draw"],
        ["text", "a"],
        ["text", "("],
        ["integer", "255"],
        ["text", ","],
        ["integer", "255"],
        ["text", ","],
        ["integer", "255"],
        ["text", ")"],
        ["text", "line"],
        ["text", "from"],
        ["text", "("],
        ["float", "0.000000"],
        ["text", ","],
        ["float", "0.000000"],
        ["text", ")"],
        ["text", "to"],
        ["text", "("],
        ["float", "1000.000000"],
        ["text", ","],
        ["float", "800.000000"],
        ["text", ")"]]})
    claims.append({"type": "claim", "fact": [
        ["source", paper_drawing_target_id],
        ["text", "draw"],
        ["text", "centered"],
        ["text", "label"],
        ["text", "Hello World centered label"],
        ["text", "at"],
        ["text", "("],
        ["float", "20.000000"],
        ["text", ","],
        ["float", "20.000000"],
        ["text", ")"]]})
    claims.append({"type": "claim", "fact": [
        ["source", paper_drawing_target_id],
        ["text", "draw"],
        ["text", "16pt"],
        ["text", "text"],
        ["text", "Hello World text"],
        ["text", "at"],
        ["text", "("],
        ["float", "20.000000"],
        ["text", ","],
        ["float", "500.000000"],
        ["text", ")"]]})
    batch(claims)


init(MY_ID)
