import requests
import time
import json

URL = "http://localhost:3000/"

def say(fact):
    payload = {'facts': fact}
    return requests.post(URL + "assert", data=payload)

def retract(fact):
    payload = {'facts': fact}
    return requests.post(URL + "retract", data=payload)

def select(fact):
    payload = {'facts': fact}
    response = requests.post(URL + "select", data=payload)
    return response.json()

sourceCode = """
const Room = require('@living-room/client-js')
const room = new Room()

// comment

room.assert('hello from way inside programEditTest FROM PYTHON')
"""

sourceCodeStr = json.dumps(sourceCode)[1:-1]

print("Hello from pythonTest.py")
retract("wish testy.js has source code $sourceCode")
say("wish testy.js has source code \"{}\"".format(sourceCodeStr))
