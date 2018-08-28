const PerspT = require('perspective-transform');
const Room = require('@living-room/client-js')
const fs = require('fs');
const path = require('path');
const readline = require('readline');

const scriptName = path.basename(__filename);
const scriptNameNoExtension = path.parse(scriptName).name;
const logPath = __filename.replace(scriptName, 'logs/' + scriptNameNoExtension + '.log')
const access = fs.createWriteStream(logPath)
process.stdout.write = process.stderr.write = access.write.bind(access);
process.on('uncaughtException', function(err) {
  console.error((err && err.stack) ? err.stack : err);
})

const room = new Room()

const add_vec = (vec1, vec2) =>
  ({"x": vec1["x"] + vec2["x"], "y": vec1["y"] + vec2["y"]})

const diff_vec = (vec1, vec2) =>
  ({"x": vec1["x"] - vec2["x"], "y": vec1["y"] - vec2["y"]})

const scale_vec = (vec, scale) =>
  ({"x": vec["x"] * scale, "y": vec["y"] * scale})

const vec_length = (vec) =>
  Math.sqrt(vec["x"] * vec["x"] + vec["y"] * vec["y"])

const paper_approximation = (paper, perspT, canvasWidth) => {
  const perspTCorner = corner => {
    const pt = perspT.transform(corner.x, corner.y)
    return {x: normToCoord(pt[0], canvasWidth), y: normToCoord(pt[1])};
  }
  const perspTL = perspTCorner(paper.TL);
  const perspTR = perspTCorner(paper.TR);
  const perspBR = perspTCorner(paper.BR);
  const perspBL = perspTCorner(paper.BL);
  const top = diff_vec(perspTR, perspTL);
  const left = diff_vec(perspBL, perspTL);
  const projTR = add_vec(perspTL, top);
  const projBL = add_vec(perspTL, left);
  const origin = perspTL;
  const width = vec_length(top);
  const height = vec_length(left);
  const angle_radians = Math.atan2(top.y, top.x);
  return {width, height, angle_radians, origin};
}

room.on(
  `camera $cameraId has projector calibration TL ($PCx1, $PCy1) TR ($PCx2, $PCy2) BR ($PCx3, $PCy3) BL ($PCx4, $PCy4) @ $`,
  `camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $`,
  (data) => {
    if (
      !isNaN(data.x1) && !isNaN(data.y1) &&
      !isNaN(data.x2) && !isNaN(data.y2) &&
      !isNaN(data.x3) && !isNaN(data.y3) &&
      !isNaN(data.x4) && !isNaN(data.y4)
    ) {
      const projectorCalibration = [
        data.PCx1, data.PCy1,
        data.PCx2, data.PCy2,
        data.PCx3, data.PCy3,
        data.PCx4, data.PCy4
      ]
      const perspT = PerspT(projectorCalibration, [0, 0, 1.0, 0, 1.0, 1.0, 0, 1.0]);
      const canvasWidth = 1920
      const paper = {
        id: data.id,
        TL: {x: data.x1, y: data.y1},
        TR: {x: data.x2, y: data.y2},
        BR: {x: data.x3, y: data.y3},
        BL: {x: data.x4, y: data.y4}
      };
      const paperApprox = paper_approximation(paper, perspT, canvasWidth);
      room.retract(`paper ${data.id} has width $ height $ angle $ at ($, $)`)
      room.assert(
        `paper ${data.id} has width ${paperApprox.width}` +
        ` height ${paperApprox.height}` +
        ` angle ${paperApprox.angle_radians}` +
        ` at (${paperApprox.origin.x}, ${paperApprox.origin.y})`
      );
    }
  }
)
