import subprocess
import requests
import time
import sys
import os

scriptName = os.path.basename(__file__)
scriptNameNoExtension = os.path.splitext(scriptName)[0]
fileDir = os.path.dirname(os.path.realpath('__file__'))
logPath = os.path.join(fileDir, 'logs/' + scriptNameNoExtension + '.log')
print(logPath)

log_file = open(logPath, 'w')
sys.stdout = log_file
sys.stderr = log_file

print("printingManager started!")

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

while True:
    print("checking for printing wishes")
    print_wishes = select('wish $name would be printed')
    for wish in print_wishes:
        name = wish['name']['word']
        retract('wish {} would be printed'.format(name))
        if '.py' not in name and '.js' not in name:
            name += '.js'
        print("PRINTING:", name)
        subprocess.call(['/usr/bin/lpr', 'src/standalone_processes/{}'.format(name)])
    time.sleep(1)
