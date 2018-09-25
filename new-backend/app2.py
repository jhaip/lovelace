from client_helper import init, claim, retract
MY_ID = 2

def prehook():
    claim("Bird has 5 toes")
    claim("Man has 10 toes")

def sub_callback(results):
    print("sub CALLBACK!")
    print(results)

subscriptions = [
    (["$ $X has $Y toes"], sub_callback)
]

init(MY_ID, prehook, [], subscriptions)
