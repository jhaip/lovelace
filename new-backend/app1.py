from client_helper import init, claim, retract
MY_ID = 1

def prehook():
    claim("Bird has 5 toes")

def select_callback(results):
    print("SELECT CALLBACK!")
    print(results)

selects = [
    (["$ $X has 5 toes"], select_callback)
]

init(MY_ID, prehook, selects)
