from helper2 import init, claim, retract, prehook, subscription, batch, MY_ID_STR, listen, check_server_connection, get_my_id_str
from graphics import Illumination
import logging

is_selecting = True

@subscription(["$ $ laser in region $regionId", "$ $ region $regionId is toggleable"])
def sub_callback_calibration(results):
    global is_selecting
    claims = []
    if results:
        for result in results:
            if is_selecting:
                claims.append({"type": "claim", "fact": [
                        ["id", "0"],
                        ["id", "1"],
                        ["text", "region"],
                        ["text", str(result["regionId"])],
                        ["text", "is"],
                        ["text", "toggled"],
                    ]})
                logging.error("region {} is on".format(result["regionId"]))
            else:
                claims.append({"type": "retract", "fact": [
                        ["id", "0"],
                        ["id", "1"],
                        ["text", "region"],
                        ["text", str(result["regionId"])],
                        ["text", "is"],
                        ["text", "toggled"],
                    ]})
                logging.error("region {} is OFF".format(result["regionId"]))
        batch(claims)


@subscription(["#0054 $ keyboard $ typed special key $key @ $t"])
def sub_callback_keyboard(results):
    global is_selecting
    claims = []
    claims.append({"type": "retract", "fact": [
        ["id", get_my_id_str()],
        ["id", "2"],
        ["postfix", ""],
    ]})
    ill = Illumination()
    if results:
        key = str(results[0]["key"])
        if key == "up" and is_selecting:
            is_selecting = False
            ill.text(0, 0, "selecting")
        else:
            is_selecting = True
            ill.text(0, 0, "deselecting")
    claims.append(ill.to_batch_claim(get_my_id_str(), "2"))
    batch(claims)

init(__file__)
