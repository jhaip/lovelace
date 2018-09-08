/* global LivingRoom, location */
const { canvas } = window
const hostname = location.hostname
const pathArray = location.pathname.split('/')
const htmlpath = pathArray[pathArray.length - 1]
const namespace = htmlpath.split('.')[0]

const room = new LivingRoom(`http://${hostname}:3000`)
const context = canvas.getContext('2d')

let texts = new Map()
let circles = new Map()
let lines = new Map()
let containedPapers = new Map()
let projectorCalibration = [
  0, 0,
  1920, 0,
  1920, 1080,
  0, 1080
];

const normToCoord = (n, s = canvas.height) => (Number.isInteger(n) ? n : n * s)

const extendLanguage = (language, storage, convertFunc) => {
  const update = ({ assertions, retractions }) => {
    retractions.forEach(x => storage.delete(JSON.stringify(x)));
    assertions.forEach(x => storage.set(
      JSON.stringify(x),
      Object.assign(convertFunc(x), {paper: x.paper})
    ));
    scheduleDraw();
  }
  language.forEach(statement => {
    room.subscribe(`${statement}`, update);
    room.subscribe(`${statement} on paper $paper`, update);
    room.subscribe(`${namespace}: ${statement}`, update);
    room.subscribe(`${namespace}: ${statement} on paper $paper`, update);
  });
}

// draw label Timon at (0.3, 0.3)
// draw centered label "Hello" at (0.5, 0.5) on paper 395
// `draw text "timon is cool" at (0.8, 0.8)`
// `draw small text "timon is cool" at (0.8, 0.8)`
extendLanguage(
  [
    `draw label $text at ($x, $y)`,
    `draw $centered label $text at ($x, $y)`,
    `draw text $text at ($x, $y)`,
    `draw $size text $text at ($x, $y)`,
    `draw $size text $text at ($x, $y) at angle $angle`
  ],
  texts,
  text => ({
    centered: text.centered || false,
    size: text.size,
    angle: text.angle,
    text: text.text || "",
    x: text.x || 0,
    y: text.y || 0
  })
);

// draw a (254, 254, 255) line from (0.01, 0.01) to (1.0, 1.0) on paper 395
extendLanguage(
  [`draw a ($r, $g, $b) line from ($x, $y) to ($xx, $yy)`],
  lines,
  line => ({
    r: line.r || 0,
    g: line.g || 0,
    b: line.b || 0,
    x: line.x || 0,
    y: line.y || 0,
    xx: line.xx || 0,
    yy: line.yy || 0
  })
);

//  `draw a (255, 12, 123) circle at (0.5, 0.6) with radius 0.1`
//  `draw a (255, 12, 123) halo around (0.5, 0.6) with radius 0.1`
extendLanguage(
  [`draw a ($fillR, $fillG, $fillB) circle at ($x, $y) with radius $radius`],
  circles,
  circle => ({
    x: circle.x || 0,
    y: circle.y || 0,
    fill: true,
    stroke: false,
    fillR: circle.fillR || 0,
    fillG: circle.fillG || 0,
    fillB: circle.fillB || 0,
    radius: circle.radius || 0
  })
);

extendLanguage(
  [`draw a ($strokeR, $strokeG, $strokeB) halo around ($x, $y) with radius $radius`],
  circles,
  circle => ({
    x: circle.x || 0,
    y: circle.y || 0,
    fill: false,
    stroke: true,
    strokeR: circle.strokeR || 0,
    strokeG: circle.strokeG || 0,
    strokeB: circle.strokeB || 0,
    radius: circle.radius || 0
  })
);

room.subscribe(
  `camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  ({ assertions }) => {
    if (!assertions || assertions.length === 0) {
      return;
    }
    containedPapers = new Map();
    assertions.forEach(p => {
      if (
        !isNaN(p.x1) && !isNaN(p.y1) &&
        !isNaN(p.x2) && !isNaN(p.y2) &&
        !isNaN(p.x3) && !isNaN(p.y3) &&
        !isNaN(p.x4) && !isNaN(p.y4)
      ) {
        containedPapers.set(p.id, {
          id: p.id,
          TL: {x: p.x1, y: p.y1},
          TR: {x: p.x2, y: p.y2},
          BR: {x: p.x3, y: p.y3},
          BL: {x: p.x4, y: p.y4}
        });
      }
    });
    scheduleDraw()
  }
)

room.subscribe(
  `camera $cameraId has projector calibration TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  ({ assertions }) => {
    if (!assertions || assertions.length === 0) {
      return;
    }
    assertions.forEach(a => {
      projectorCalibration = [
        a.x1, a.y1,
        a.x2, a.y2,
        a.x3, a.y3,
        a.x4, a.y4,
      ];
    })
  }
)

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

async function draw (time) {
  // if the window is resized, change the canvas to fill the window
  canvas.width = canvas.clientWidth * window.devicePixelRatio
  canvas.height = canvas.clientHeight * window.devicePixelRatio

  const perspT = PerspT(projectorCalibration, [0, 0, 1.0, 0, 1.0, 1.0, 0, 1.0]);

  // clear the canvas
  context.clearRect(0, 0, canvas.width, canvas.height)

  texts.forEach(({ text, x, y, centered, size, angle, paper }) => {
    context.save()
    context.fillStyle = '#fff'
    if (centered === 'centered') {
      context.textBaseline = `middle`
      context.textAlign = `center`
    }
    // TODO:
    // if (typeof angle !== 'undefined') {
    //   context.translate(normToCoord(x, canvas.width), normToCoord(y))
    //   context.rotate(-angle) // counterclockwise
    //   context.translate(-normToCoord(x, canvas.width), -normToCoord(y))
    // }
    let width = canvas.width;
    let height = canvas.height;
    if (!!paper) {
      if (containedPapers.has(paper)) {
        containerPaper = containedPapers.get(paper);
        const paperApprox = paper_approximation(containerPaper, perspT, canvas.width);
        context.translate(paperApprox.origin.x, paperApprox.origin.y);
        context.rotate(paperApprox.angle_radians)
        width = paperApprox.width;
        height = paperApprox.height;
      } else {
        return; // paper details aren't known, so don't draw this.
      }
    }
    context.font = `${(40./1080.) * height}px monospace`
    if (size === 'small') {
      context.font = `${(20./1080.) * height}px monospace`
    }
    if (typeof size === "string" && size.includes("pt") && !isNaN(parseInt(size))) {
      context.font = `${height * parseInt(size) / 1080.}px monospace`
    }
    context.fillText(text, normToCoord(x, width), normToCoord(y, height))
    context.restore()
  })

  circles.forEach(({ x, y, fill, stroke, fillR, fillG, fillB, strokeR, strokeG, strokeB, radius, paper }) => {
    context.save()
    if (fill) {
      context.fillStyle = `rgb(${fillR},${fillG},${fillB})`
    }
    if (stroke) {
      context.strokeStyle = `rgb(${strokeR},${strokeG},${strokeB})`
    }
    let width = canvas.width;
    let height = canvas.height;
    if (!!paper) {
      if (containedPapers.has(paper)) {
        containerPaper = containedPapers.get(paper);
        const paperApprox = paper_approximation(containerPaper, perspT, canvas.width);
        context.translate(paperApprox.origin.x, paperApprox.origin.y);
        context.rotate(paperApprox.angle_radians)
        width = paperApprox.width;
        height = paperApprox.height;
      } else {
        return; // paper details aren't known, so don't draw this.
      }
    }
    context.beginPath()
    context.ellipse(
      normToCoord(x, width),
      normToCoord(y, height),
      normToCoord(radius, height),
      normToCoord(radius, height),
      0,
      0,
      2 * Math.PI
    )
    if (fill) {
      context.fill()
    }
    if (stroke) {
      context.stroke()
    }
    context.restore()
  })

  lines.forEach(({ x, y, xx, yy, r, g, b, paper }) => {
    context.save()
    context.strokeStyle = `rgb(${r},${g},${b})`
    context.lineWidth = 3;
    context.beginPath()
    let width = canvas.width;
    let height = canvas.height;
    if (!!paper) {
      if (containedPapers.has(paper)) {
        containerPaper = containedPapers.get(paper);
        const paperApprox = paper_approximation(containerPaper, perspT, canvas.width);
        context.translate(paperApprox.origin.x, paperApprox.origin.y);
        context.rotate(paperApprox.angle_radians)
        width = paperApprox.width;
        height = paperApprox.height;
      } else {
        return; // paper details aren't known, so don't draw this.
      }
    }
    context.moveTo(normToCoord(x, width), normToCoord(y, height))
    context.lineTo(normToCoord(xx, width), normToCoord(yy, height))
    context.stroke()
    context.restore()
  })
}

let drawAnimationFrame = null
function scheduleDraw () {
  if (drawAnimationFrame) return
  drawAnimationFrame = window.requestAnimationFrame(() => {
    drawAnimationFrame = null
    draw()
  })
}

canvas.onclick = async () => {
  const frames = await room.select(`time is $frame`)
  if (!frames.length) return
  const frame = frames[0].frame.value
  room
    .retract(`mouse clicked on frame $`)
    .assert(`mouse clicked on frame ${frame}`)
}

window.addEventListener('resize', draw)

draw()
