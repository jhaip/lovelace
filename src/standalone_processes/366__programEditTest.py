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
import requests
import time
import json

URL = 'http://localhost:3000/'

def say(fact):
    payload = {'facts': fact}
    return requests.post(URL + 'assert', data=payload)

say('hello from python')
"""

sourceCodeStr = json.dumps(sourceCode)[1:-1]

print("Hello from pythonTest.py")
retract("wish testy.py has source code $sourceCode")
# No number or underscores in the process name, at least without quotes
say("wish testyPython.py has source code \"{}\"".format(sourceCodeStr))
