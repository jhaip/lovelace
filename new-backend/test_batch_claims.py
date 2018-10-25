from client_helper import init, claim, retract, prehook, subscription, batch, MY_ID_STR
import sys
import time
MY_ID = sys.argv[1]
print(MY_ID)

N = 100
start = None

"""
1 : 2.74ms
2 : 2.94ms
4 : 3.92ms
8 : 10.6ms
16: 12.8ms
32: 26.6ms
64: 138.ms
"""

"""
1: 3.03ms
0.6ms sending subscription results x6
1.14ms parse x1
0.3ms notify subscribers queries x3
0.2ms claiming x2
extra might be parsing subscriber + travel time

goal:
0.1ms sending subscriber results
0 parse by doing it on the client side
0.1ms notify subscribers
0.1ms claiming (maybe higher for larger N)
extra stuff to send + receive multiple things for 1 loop: 0.2ms
~1 ms for 1 batch claim + 1 subscribers receiving batch once
"""

# @prehook
def my_prehook():
    global N, start, MY_ID_STR
    start = time.time()
    # now: 2 ms * N
    # desired: 2 ms
    # where is the other 1 ms from?

    # 0.056 ms * N <-- claim
    # 0.669 ms * N <-- parse
    # 0.110 ms * N*0.8 <-- retract
    # 0.060 ms * N*1.5 <-- notify subscribers
    for i in range(N):
        # ~ 2 ms per claim
        claim("dots 1036.000000 541.000000 color 77.000000 54.000000 34.000000 {}".format(
            i))


@prehook
def my_prehook2():
    global N, start, MY_ID_STR
    start = time.time()
    claims = []
    for i in range(N):
        claims.append({"type": "claim", "fact": [
                      ["source", MY_ID_STR], ["text", "dots"], ["float", "1036.000000"], ["float", "1036.000000"], ["text", "color"], ["float", "77.000000"], ["float", "77.000000"], ["float", "77.000000"], ["integer", str(i)]]})
    batch(claims)


@subscription(["$source dots $x $y color $r $g $b $t".format(MY_ID)])
def sub_callback(results):
    global N, start
    print(results)
    end = time.time()
    print("TIME: {} ms".format((end - start)*1000.0))
    print("AVG LOOP TIME: {} ms".format((end - start)*1000.0/N))
    # sys.exit()


init(MY_ID)
