from client_helper import init, claim, retract
import sys
import time
MY_ID = sys.argv[1]
print(MY_ID)

N = 10
i = 0
start = None

def prehook():
    global start
    start = time.time()
    claim("Bird has 5 toes")
    claim("Man has 10 toes")

def sub_callback(results):
    global i, N, start
    print("sub CALLBACK!")
    print(results)
    print(i)
    i += 1
    if i >= N:
        end = time.time()
        print("TIME: {} ms".format((end - start)*1000.0))
        print("AVG LOOP TIME: {} ms".format((end - start)*1000.0/N))
        sys.exit()
    claim("Man has {} toes".format(i))

subscriptions = [
    (["$ $X has $Y toes"], sub_callback)
]

init(MY_ID, prehook, [], subscriptions)
