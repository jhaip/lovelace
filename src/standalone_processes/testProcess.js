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
      if (
        !isNaN(p.x1) && !isNaN(p.y1) &&
        !isNaN(p.x2) && !isNaN(p.y2) &&
        !isNaN(p.x3) && !isNaN(p.y3) &&
        !isNaN(p.x4) && !isNaN(p.y4)
      ) {
        room.assert(`draw a (255, 255, 1) line from (${p.x1/W}, ${p.y1/H}) to (${p.x2/W}, ${p.y2/H})`)
        room.assert(`draw a (255, 255, 1) line from (${p.x2/W}, ${p.y2/H}) to (${p.x3/W}, ${p.y3/H})`)
        room.assert(`draw a (255, 255, 1) line from (${p.x3/W}, ${p.y3/H}) to (${p.x4/W}, ${p.y4/H})`)
        room.assert(`draw a (255, 255, 1) line from (${p.x4/W}, ${p.y4/H}) to (${p.x1/W}, ${p.y1/H})`)
        room.assert(`draw text "Paper ${p.id}" at (${p.x1/W}, ${p.y1/H})`)
      }
    })
  }
)

const move_along_vector = (amount, vector) => {
  const size = Math.sqrt(vector["x"] * vector["x"] + vector["y"] * vector["y"])
  let C = 1.0
  if (size != 0) {
    C = 1.0 * amount / size
  }
  return {"x": C * vector["x"], "y": C * vector["y"]}
}

const add_vec = (vec1, vec2) =>
  ({"x": vec1["x"] + vec2["x"], "y": vec1["y"] + vec2["y"]})

const diff_vec = (vec1, vec2) =>
  ({"x": vec1["x"] - vec2["x"], "y": vec1["y"] - vec2["y"]})

const scale_vec = (vec, scale) =>
  ({"x": vec["x"] * scale, "y": vec["y"] * scale})

/* Example claims:
draw a (1, 1, 255) line from (0.01, 0.01) to (1.0, 1.0) on paper 395
draw a (1, 1, 255) line from (0.2, 0.2) to (0.8, 0.2) on paper 395
draw a (1, 1, 255) line from (0.8, 0.2) to (0.8, 0.8) on paper 395
draw a (1, 1, 255) line from (0.8, 0.8) to (0.2, 0.8) on paper 395
draw a (1, 1, 255) line from (0.2, 0.8) to (0.2, 0.2) on paper 395
 */
room.subscribe(
  `camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  `draw a ($r, $g, $b) line from ($x, $y) to ($xx, $yy) on paper $id`,
  ({ assertions }) => {
    if (!assertions || assertions.length === 0) {
      return;
    }
    console.log("ADAPTING PAPER!!")
    console.log(assertions);
    room.retract(`draw a (255, 1, 255) line from ($, $) to ($, $)`)
    assertions.forEach(p => {
      W = 1920.0;
      H = 1080.0;
      if (
        !isNaN(p.x1) && !isNaN(p.y1) &&
        !isNaN(p.x2) && !isNaN(p.y2) &&
        !isNaN(p.x3) && !isNaN(p.y3) &&
        !isNaN(p.x4) && !isNaN(p.y4) &&
        !isNaN(p.x) && !isNaN(p.y) &&
        !isNaN(p.xx) && !isNaN(p.yy)
      ) {
        // TODO: define top and left
        const TL = {"x": p.x1, "y": p.y1};
        const TR = {"x": p.x2, "y": p.y2};
        const BL = {"x": p.x4, "y": p.y4};
        const top = diff_vec(TR, TL) // check that this is the right direction
        const left = diff_vec(BL, TL) // check that this is the right direction
        console.log(top);
        console.log(left);
        const projTR = add_vec(TL, top);
        const projBL = add_vec(TL, left);
        const P1 = {"x": p.x, "y": p.y};
        const P2 = {"x": p.xx, "y": p.yy};
        const projP1 = add_vec(add_vec(TL, scale_vec(top, p.x)), scale_vec(left, p.y))
        const projP2 = add_vec(add_vec(TL, scale_vec(top, p.xx)), scale_vec(left, p.yy))
        console.log("projected point:");
        console.log(projP1);
        console.log(projP2);
        room.assert(`draw a (255, 1, 255) line from (${projP1["x"]/W}, ${projP1["y"]/H}) to (${projP2["x"]/W}, ${projP2["y"]/H})`)
      }
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
