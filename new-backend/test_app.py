from client_helper import init, claim, retract, prehook, subscription
import sys
import time
MY_ID = sys.argv[1]
print(MY_ID)

N = 100
i = 1
start = None

@prehook
def my_prehook():
    global start
    start = time.time()
    claim("Man {} has 0 toes".format(MY_ID))


@subscription(["$ $X {} has $Y toes".format(MY_ID)])
def sub_callback(results):
    global i, N, start
    print(i)
    print(results)
    i += 1
    if i >= N:
        end = time.time()
        print("TIME: {} ms".format((end - start)*1000.0))
        print("AVG LOOP TIME: {} ms".format((end - start)*1000.0/N))
        sys.exit()
    claim("Man {} has {} toes".format(MY_ID, i))

init(MY_ID)
