from helper2 import init, claim, retract, prehook, subscription, batch, MY_ID_STR, listen, check_server_connection, get_my_id_str
from graphics import Illumination
import logging

is_selecting = True

@subscription(["$ $ laser in region $region", "$ $ region $region is toggleable"])
def sub_callback_calibration(results):
    global is_selecting
    claims = []
    claims.append({"type": "retract", "fact": [
        ["id", get_my_id_str()],
        ["id", "1"],
        ["postfix", ""],
    ]})
    if results:
        logging.error("Got results", is_selecting)
        for result in results:
            if is_selecting:
                claims.append({"type": "claim", "fact": [
                        ["id", get_my_id_str()],
                        ["id", "1"],
                        ["text", "region"],
                        ["text", str(result["regionId"])],
                        ["text", "is"],
                        ["text", "toggled"],
                    ]})
            else:
                claims.append({"type": "retract", "fact": [
                        ["id", get_my_id_str()],
                        ["id", "1"],
                        ["text", "region"],
                        ["text", str(result["regionId"])],
                        ["text", "is"],
                        ["text", "toggled"],
                    ]})
    batch(claims)


@subscription(["#0054 $ keyboard $ typed special key $key @ $t"])
def sub_callback_keyboard(results):
    global is_selecting
    if results:
        key = str(results[0]["key"])
        logging.error("checking key", key)
        if key == "up" and is_selecting:
            is_selecting = False
        else:
            is_selecting = True

init(__file__)