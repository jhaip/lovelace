from client_helper import init, claim, retract
MY_ID = 2

def prehook():
    pass

def select_callback(results):
    print("SELECT CALLBACK!")
    print(results)

selects = [
    (["$ $X has $Y toes"], select_callback)
]

init(MY_ID, prehook, selects)
