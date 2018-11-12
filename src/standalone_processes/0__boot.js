const execFile = require('child_process').execFile;
const execFileSync = require('child_process').execFileSync;
const { room, myId, scriptName, run } = require('../helper2')(__filename);

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
room.assert('wish', ["text", "1900__processManager.js"], 'would be running')
room.assert('wish', ["text", "826__runSeenPapers.js"], 'would be running')
// room.assert('wish "390__initialProgramCode.js" would be running')

/*** Initial boot values ***/
// now handled by 989:
// room.assert(`camera 1 has projector calibration TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080) @ 1`)

/*** Claim that a (fake) camera can see all boot papers ***/
// Initial Program Code:
// room.assert(`camera 1 sees paper 390 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Printing Manager:
// room.assert(`camera 99 sees paper 498 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
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
// Dots to papers
room.assert(`camera 99 sees paper 1800 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Print paper
room.assert(`camera 99 sees paper 1382 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Process Manager
room.assert(`camera 99 sees paper 1900 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Persist Projector Calibration
room.assert(`camera 99 sees paper 989 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Debug web viewer
room.assert(`camera 99 sees paper 10 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Latency measurement
room.assert(`camera 99 sees paper 11 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
//
/* TODO:
- keyboard.py
- frame-to-dots.py
*/

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

// room.assert(`camera 99 has projector calibration TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080) @ 1`)
// room.assert(`camera 99 sees paper 1013 at TL (100, 100) TR (1200, 100) BR (1200, 800) BL (100, 800) @ 1`)
// room.assert(`paper 1013 is pointing at paper 472`)  // comment out if pointingAt.py is running
// room.assert(`wish paper 498 at "498__printingManager.py" would be printed`)
// room.assert(`camera 99 sees paper 620 at TL (0, 0) TR (400, 0) BR (400, 200) BL (0, 200) @ 1`)
// room.assert(`camera 99 sees paper 472 at TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080) @ 1`)
// room.assert(`camera 99 sees paper 1013 at TL (0, 300) TR (400, 230) BR (430, 580) BL (30, 630) @ 1`)
// room.assertForOtherSource('472', `draw a (255, 255, 255) line from (0, 0) to (400, 400)`)
// room.assertForOtherSource('1013', `draw a (255, 255, 255) line from (0, 0) to (400, 400)`)
// room.assertForOtherSource('0472', [
// ["text", "draw"],
// ["text", "a"],
// ["text", "("],
// ["integer", "255"],
// ["text", ","],
// ["integer", "255"],
// ["text", ","],
// ["integer", "255"],
// ["text", ")"],
// ["text", "line"],
// ["text", "from"],
// ["text", "("],
// ["float", "0.000000"],
// ["text", ","],
// ["float", "0.000000"],
// ["text", ")"],
// ["text", "to"],
// ["text", "("],
// ["float", "400.000000"],
// ["text", ","],
// ["float", "400.000000"],
// ["text", ")"]])
// room.assertForOtherSource('1013', [
//   ["text", "draw"],
//   ["text", "a"],
//   ["text", "("],
//   ["integer", "255"],
//   ["text", ","],
//   ["integer", "255"],
//   ["text", ","],
//   ["integer", "255"],
//   ["text", ")"],
//   ["text", "line"],
//   ["text", "from"],
//   ["text", "("],
//   ["float", "0.000000"],
//   ["text", ","],
//   ["float", "0.000000"],
//   ["text", ")"],
//   ["text", "to"],
//   ["text", "("],
//   ["float", "400.000000"],
//   ["text", ","],
//   ["float", "400.000000"],
//   ["text", ")"]])
// room.assert(`camera 99 sees paper 777 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)

run();