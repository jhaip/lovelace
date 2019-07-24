import time
import logging
import json
import uuid
import zmq
from helper2 import init, claim, retract, prehook, subscription, batch, get_my_id_str

proxy_context = zmq.Context()
proxy_client = None
proxy_connected = False
last_proxy_heartbeat = time.time()
health_check_delay_s = 10
PROXY_URL = "10.0.0.22"

def check_and_connect_proxy_server():
    global proxy_context, proxy_client, proxy_connected, PROXY_URL, last_proxy_heartbeat, health_check_delay_s
    if not proxy_connected or time.time() - last_proxy_heartbeat > health_check_delay_s:
        if not proxy_connected:
            logging.info("creating a new proxy_content")
            proxy_client = proxy_context.socket(zmq.DEALER)
            proxy_client.setsockopt(zmq.IDENTITY, get_my_id_str().encode())
            proxy_client.connect("tcp://{0}:5570".format(PROXY_URL))
            logging.info("connection established")
        else:
            logging.info("checking if proxy server is still alive")
        last_proxy_heartbeat = time.time()
        init_ping_id = str(uuid.uuid4())
        proxy_client.send_multipart([".....PING{}{}".format(get_my_id_str(), init_ping_id).encode()])
        proxy_connected = False
        proxy_server_timeout_s = 2
        poll_start_time = time.time()
        while time.time() - poll_start_time < proxy_server_timeout_s:
            try:
                raw_msg = proxy_client.recv_multipart(flags=zmq.NOBLOCK)
                proxy_connected = True
                break
            except zmq.Again:
                time.sleep(0.01)
        if not proxy_connected:
            logging.info("proxy server died, message dropped")
            proxy_client.disconnect("tcp://{0}:5570".format(PROXY_URL))
            logging.info("disconnected proxy_client")
            proxy_client.close()
            logging.info("closed proxy_client")
            return False
    return True

@subscription(["$ $ camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time"])
def sub_callback_papers(results):
    global proxy_client
    if not check_and_connect_proxy_server():
        return
    logging.info("proxying message")
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