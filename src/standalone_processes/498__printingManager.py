import subprocess
from helper import *
init(__file__)

while True:
    logging.info("checking for printing wishes")
    print_wishes = select('wish file $name would be printed')
    for wish in print_wishes:
        name = wish['name']['value']
        retract('wish file "{}" would be printed'.format(name), '$')
        logging.info("PRINTING:")
        logging.info(name)
        subprocess.call(['/usr/bin/lpr', name])
    time.sleep(1)

logging.info("exited --- ")
