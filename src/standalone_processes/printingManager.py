import subprocess
import requests
import time

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
    print_wishes = select('wish $name would be printed')
    for wish in print_wishes:
        name = wish['name']['word']
        retract('wish {} would be printed'.format(name))
        if '.py' not in name and '.js' not in name:
            name += '.js'
        subprocess.call(['/usr/bin/lpr', 'src/standalone_processes/{}'.format(name)])
    time.sleep(1)
