from client_helper import init, claim, retract, prehook, subscription
import sys
import time
from threading import Timer

MY_ID = sys.argv[1]
print(MY_ID)

def test_calls():
    # Repeated claim of Fox is out, shouldn't do anything
    claim("Fox is out")
    time.sleep(1.0)
    retract("$X Fox is out")
    time.sleep(1.0)
    # Claiming after the fact was retracted so this will trigger the subscriber
    claim("Fox is out")
    time.sleep(1.0)
    claim("Fox is out")
    time.sleep(1.0)

@prehook
def my_prehook():
    claim("Fox is out")
    t = Timer(1.0, test_calls)
    t.start()


@subscription(["$X Fox is out"])
def sub_callback(results):
    print("sub CALLBACK!")
    print(results)

init(MY_ID)
