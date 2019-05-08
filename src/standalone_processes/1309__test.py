import time
import logging
from helper2 import init, claim, retract, prehook, subscription, batch, get_my_id_pre_init, get_my_id_str

M = 1300
MY_ID = str(get_my_id_pre_init(__file__))
P = int(get_my_id_pre_init(__file__))-1

@subscription(["$ test client " + str(M) + " says $x @ $time"])
def sub_callback(results):
    currentTimeMs = int(round(time.time() * 1000))
    claims = []
    claims.append({"type": "retract", "fact": [
        ["id", get_my_id_str()],
        ["postfix", ""],
    ]})
    claims.append({"type": "claim", "fact": [
        ["text", get_my_id_str()],
        ["text", "test"],
        ["text", "client"],
        ["integer", MY_ID],
        ["text", "says"],
        ["integer", MY_ID],
        ["text", "@"],
        ["integer", str(currentTimeMs)]
    ]})
    batch(claims)

init(__file__)
