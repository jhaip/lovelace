from client_helper import init, claim, retract
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

def prehook():
    claim("Fox is out")
    t = Timer(1.0, test_calls)
    t.start()


def sub_callback(results):
    print("sub CALLBACK!")
    print(results)


subscriptions = [
    (["$X Fox is out"], sub_callback)
]

init(MY_ID, prehook, [], subscriptions)
