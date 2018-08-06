const Room = require('@living-room/client-js')
const fs = require('fs');
const path = require('path');

const scriptName = path.basename(__filename);
const scriptNameNoExtension = path.parse(scriptName).name;
const logPath = __filename.replace(scriptName, 'logs/' + scriptNameNoExtension + ".log")
const access = fs.createWriteStream(logPath)
process.stdout.write = process.stderr.write = access.write.bind(access);
process.on('uncaughtException', function(err) {
  console.error((err && err.stack) ? err.stack : err);
})

const room = new Room()

console.log("start testProcess")
room.retract(`hello from testProcess @ $`)

setInterval(() => {
  console.error("hello from testProcess", new Date())
  room
    .retract(`hello from testProcess @ $`)
    .assert(`hello from testProcess @ ${(new Date()).getTime()}`)
}, 1000)


// DRAW PAPERS
room.subscribe(
  `camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  ({ assertions }) => {
    if (!assertions || assertions.length === 0) {
      return;
    }
    console.log("ASSERTIONS:")
    console.log(assertions);
    const visibleIDs = assertions.map(paper => String(paper.id))
    room.retract(`draw a (255, 255, 1) line from ($, $) to ($, $)`)
    room.retract(`draw text $ at ($, $)`)
    assertions.forEach(p => {
      W = 1920.0;
      H = 1080.0;
      room.assert(`draw a (255, 255, 1) line from (${p.x1/W}, ${p.y1/H}) to (${p.x2/W}, ${p.y2/H})`)
      room.assert(`draw a (255, 255, 1) line from (${p.x2/W}, ${p.y2/H}) to (${p.x3/W}, ${p.y3/H})`)
      room.assert(`draw a (255, 255, 1) line from (${p.x3/W}, ${p.y3/H}) to (${p.x4/W}, ${p.y4/H})`)
      room.assert(`draw a (255, 255, 1) line from (${p.x4/W}, ${p.y4/H}) to (${p.x1/W}, ${p.y1/H})`)
      room.assert(`draw text "Paper ${p.id}" at (${p.x1/W}, ${p.y1/H})`)
    })
  }
)

room.subscribe(
  `camera 1 sees no papers @ $time`,
  ({ assertions }) => {
    if (!assertions || assertions.length === 0) {
      return;
    }
    // this may not do anything because this program might not be running
    // when there are no papers. If it's not running then it can't retract it.
    room.retract(`draw a (255, 255, 1) line from ($, $) to ($, $)`)
    room.retract(`draw text $ at ($, $)`);
  }
)
