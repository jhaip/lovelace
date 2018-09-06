const Room = require('@living-room/client-js')
const execFile = require('child_process').execFile;
const execFileSync = require('child_process').execFileSync;

const room = new Room()

/*** Start the program that can start all other programs ***/
console.error("pre--------DONE WITH INITIAL PROGRAM CODE")
const child0 = execFileSync(
  'node',
  [`src/standalone_processes/390__initialProgramCode.js`]
);
console.error("DONE WITH INITIAL PROGRAM CODE")
const child = execFile(
  'node',
  [`src/standalone_processes/1900__processManager.js`],
  (error, stdout, stderr) => {
    if (error) {
        console.error('stderr', stderr);
        console.error(error);
    }
    console.log('stdout', stdout);
});


/*** Start the programs that actually starts all boot programs ***/
room.assert(`camera 99 sees paper 1900 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
room.assert('wish "1900__processManager.js" would be running')
room.assert('wish "826__runSeenPapers.js" would be running')
// room.assert('wish "390__initialProgramCode.js" would be running')

/*** Claim that a (fake) camera can see all boot papers ***/
// Initial Program Code:
// room.assert(`camera 1 sees paper 390 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Printing Manager:
room.assert(`camera 99 sees paper 498 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Program Editor (may not be needed now?)
room.assert(`camera 99 sees paper 577 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Run Seen Papers
room.assert(`camera 99 sees paper 826 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Pointing At
room.assert(`camera 99 sees paper 277 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Paper Details
room.assert(`camera 99 sees paper 620 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Create New Paper
room.assert(`camera 99 sees paper 1459 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Process Manager
room.assert(`camera 99 sees paper 1800 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Print paper
room.assert(`camera 99 sees paper 1382 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Dots to papers
room.assert(`camera 99 sees paper 1900 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
/* TODO:
- keyboard.py
- frame-to-dots.py
- dots-to-papers.go
*/

/*** Initial boot values ***/
room.assert(`camera 1 has projector calibration TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080) @ 1`)

/*** DEBUG STUFF ***/
const sourceCode = `
import requests
import time
import json

URL = 'http://localhost:3000/'

def say(fact):
    payload = {'facts': fact}
    return requests.post(URL + 'assert', data=payload)

say('hello from python')
`;
const cleanSourceCode = sourceCode.replace(/\n/g, '\\n').replace(/"/g, String.fromCharCode(9787));
// room.assert(`wish a paper would be created in "py" with source code "${cleanSourceCode}" @ 1`)

// room.assert(`camera 99 sees paper 1924 at TL (100, 100) TR (1200, 100) BR (1200, 800) BL (100, 800) @ 1`)
// room.assert(`paper 1924 is pointing at paper 472`)  // comment out if pointingAt.py is running
