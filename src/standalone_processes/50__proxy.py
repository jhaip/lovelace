import time
import logging
import json
import uuid
import zmq
from helper2 import init, claim, retract, prehook, subscription, batch, get_my_id_str

proxy_context = zmq.Context()
proxy_client = proxy_context.socket(zmq.DEALER)
proxy_connected = False
PROXY_URL = "10.0.0.22"

@subscription(["$ $ camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time"])
def sub_callback_papers(results):
    if not proxy_connected:
        proxy_client.setsockopt(zmq.IDENTITY, get_my_id_str().encode())
        proxy_client.connect("tcp://{0}:5570".format(PROXY_URL))
        init_ping_id = str(uuid.uuid4())
        proxy_client.send_multipart([".....PING{}{}".format(get_my_id_str(), init_ping_id).encode()])
        # assume the first message recv'd will be the PING response
        raw_msg = client.recv_multipart(flags=0)  # Blocks until a message is received
        proxy_connected = True
        logging.info("connected to proxy!")
    claims = []
    claims.append({"type": "retract", "fact": [
        ["id", get_my_id_str()],
        ["id", "1"],
        ["postfix", ""],
    ]})
    for result in results:
        claims.append({"type": "claim", "fact": [
            ["id", get_my_id_str()],
            ["id", "1"],
            ["text", "camera"],
            ["integer", str(result["cameraId"])],
            ["text", "sees"],
            ["text", "paper"],
            ["integer", str(result["id"])],
            ["text", "at"],
            ["text", "TL"],
            ["text", "("],
            ["integer", str(result["x1"])],
            ["text", ","],
            ["integer", str(result["y1"])],
            ["text", ")"],
            ["text", "TR"],
            ["text", "("],
            ["integer", str(result["x2"])],
            ["text", ","],
            ["integer", str(result["y2"])],
            ["text", ")"],
            ["text", "BR"],
            ["text", "("],
            ["integer", str(result["x3"])],
            ["text", ","],
            ["integer", str(result["y3"])],
            ["text", ")"],
            ["text", "BL"],
            ["text", "("],
            ["integer", str(result["x4"])],
            ["text", ","],
            ["integer", str(result["y4"])],
            ["text", ")"],
            ["text", "@"],
            ["integer", str(result["time"])],
        ]})
    proxy_client.send_multipart(["....BATCH{}{}".format(
        get_my_id_str(), json.dumps(claims)).encode()], zmq.NOBLOCK)

init(__file__)