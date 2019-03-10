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
room.assert(`camera 99 sees paper 498 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Program Editor (may not be needed now?)
room.assert(`camera 99 sees paper 577 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Run Seen Papers
room.assert(`camera 99 sees paper 826 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// Pointing At
// room.assert(`camera 99 sees paper 277 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
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
// HTTP Client
room.assert(`camera 1 sees paper 20 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// JS compiler
room.assert(`camera 99 sees paper 40 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
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
// room.assert(`camera 99 sees paper 33 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 99 sees paper 34 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 99 sees paper 35 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)

room.assert(`camera 99 sees paper 45 at TL (0, 0) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
room.assert(`camera 99 sees paper 46 at TL (0, 0) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)

room.assert(`camera 99 sees paper 1100 at TL (0, 0) TR (1440, 0) BR (1440, 860) BL (0, 860) @ 1`)
let sPaperWidth = 200
let sPaperHeight = 300;
let sPaperHMargin = 45;
let sPaperVMargin = 20;
let sOriginX = 590;
let sOriginY = 40;
for (let x = 0; x < 3; x += 1) {
  for (let y = 0; y < 3; y += 1) {
    let idOffset = 2000 + x + y*3;
    let x1 = sOriginX + x * (sPaperWidth + sPaperHMargin);
    let x2 = sOriginX + x * (sPaperWidth + sPaperHMargin) + sPaperWidth;
    let y1 = sOriginY + y * (sPaperHeight + sPaperVMargin);
    let y2 = sOriginY + y * (sPaperHeight + sPaperVMargin) + sPaperHeight;
    // room.assert(`camera 1 sees paper ${idOffset} at TL (${x1}, ${y1}) TR (${x2}, ${y1}) BR (${x2}, ${y2}) BL (${x1}, ${y2}) @ 1`)
  }
}
// room.assert(`camera 1 sees paper 2000 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 2001 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 2002 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 2003 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 2004 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 2005 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 2006 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 2007 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 2008 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 2009 at TL (1, 1) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`camera 1 sees paper 1013 at TL (0, 0) TR (1440, 0) BR (1440, 860) BL (0, 860) @ 1`)
// room.assert(`camera 1 sees paper 648 at TL (0, 0) TR (2, 1) BR (2, 2) BL (1, 2) @ 1`)
// room.assert(`paper 1013 has width 1440 height 860 angle 0 at (0, 0)`)
// room.assert(`paper 1100 has width 1440 height 860 angle 0 at (0, 0)`)
// room.assert(`wish RGB light strand is color 0 50 0`)
// room.assert(`wish display 1700 only showed 1013`)
room.assert(`wish display 1701 only showed 2000 2001 2002 2003 2004 2005 2006 2007 2008 2009 2010 2011 5 6 1100`)

room.assert(`paper 2000 has RFID "f26a0c2e"`)
room.assert(`paper 2001 has RFID "f238222e"`)
room.assert(`paper 2002 has RFID "80616ea3"`)
room.assert(`paper 2003 has RFID "d07911a3"`)
room.assert(`paper 2004 has RFID "91b4d108"`)
room.assert(`paper 2005 has RFID "53825027"`)
room.assert(`paper 2006 has RFID "10af78a3"`)
room.assert(`paper 2007 has RFID "7341a727"`)
room.assert(`paper 2008 has RFID "2574c72d"`)
room.assert(`paper 2009 has RFID "b680cc21"`)
room.assert(`paper 2010 has RFID "737e9c27"`)
room.assert(`paper 2011 has RFID "4221cd24"`)
room.assert(`paper 5 has RFID "d01ff625"`)
room.assert(`paper 1100 has RFID "e21eef27"`)
room.assert(`paper 1013 has RFID "7bdbe359"`)

room.assert(`Photon400035001547343433313338 can flash photon Photon3c002f000e47343432313031`)

run();
