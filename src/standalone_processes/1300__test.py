import time
import logging
from helper2 import init, claim, retract, prehook, subscription, batch, get_my_id_pre_init, get_my_id_str

N = 10
MY_ID = str(get_my_id_pre_init())
F = int(get_my_id_pre_init()) + N
logging.error("testing with N = " + N)

@subscription(["$ test client " + MY_ID + " says $x @ $time1", "$ test client " + F + " says $y @ $time2"])
def sub_callback(results):
    currentTimeMs = int(round(time.time() * 1000))
    claims = []
    claims.append({"type": "retract", "fact": [
        ["id", get_my_id_str()],
        ["postfix", ""],
    ]})
    batch(claims)
    # const span = tracer.startSpan('1200-done', { childOf: room.wireCtx() });
    # const currentTimeMs = (new Date()).getTime()
    # console.error(`TEST IS DONE @ ${currentTimeMs}`)
    # console.log("elapsed time:", parseInt(results[0].time2) - parseInt(results[0].time1), "ms")
    # span.finish();

init(__file__)

# afterServerConnects(() => {
#     const span = tracer.startSpan('1200-claim', { childOf: room.wireCtx() });
#     span.log({ 'event': 'claim from #1200' });
#     const currentTimeMs = (new Date()).getTime()
#     room.assert(["text", "test"], ["text", "client"], ["integer", `${myId}`], ["text", "says"], ["integer", `${myId}`], ["text", "@"], ["integer", `${currentTimeMs}`]);
#     span.finish();
#     room.flush();
# })


