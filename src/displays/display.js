/* global LivingRoom, location */

// This is a demo of subscribing to a server query.

// Draw a word
//  `draw label $name at ($x, $y)`
//  `draw label Timon at (0.3, 0.3)`

// Draw a centered word
//  `draw centered label $name at ($x, $y)`
//  `draw centered label Timon at (0.3, 0.3)`

// Draw a sentence
//  `draw text $text at ($x, $y)`
//  `draw text "timon is cool" at (0.8, 0.8)`

// Draw a tiny sentence
//  `draw small text $text at ($x, $y)`
//  `draw small text "timon is cool" at (0.8, 0.8)`

// Drawing a line
//  `draw a ($r, $g, $b) line from ($x, $y) to ($xx, $yy)`
//  `draw a (255, 255, 0) line from (0.3, 0.3) to (0.5, 0.5)`

// Drawing a circle:
//  `$name is a ($r, $g, $b) circle at ($x, $y) with radius $radius`
//  `draw a (255, 12, 123) circle at (0.5, 0.6) with radius 0.1`

// Drawing a halo:
//  `draw a ($r, $g, $b) halo around ($x, $y) with radius $radius`
//  `draw a (255, 12, 123) halo around (0.5, 0.6) with radius 0.1`
const PerspT = require('perspective-transform');
const { canvas } = window
const hostname = location.hostname
const pathArray = location.pathname.split('/')
const htmlpath = pathArray[pathArray.length - 1]
const namespace = htmlpath.split('.')[0]

const room = new LivingRoom(`http://${hostname}:3000`)
const context = canvas.getContext('2d')

let labels = new Map()
let texts = new Map()
let circles = new Map()
let halos = new Map()
let lines = new Map()
let containedPapers = new Map()
let projectorCalibration = [
  0, 0,
  1920, 0,
  1920, 1080,
  0, 1080
];

const normToCoord = (n, s = canvas.height) => (Number.isInteger(n) ? n : n * s)

const updateLabel = ({ assertions, retractions }) => {
  retractions.forEach(label => labels.delete(JSON.stringify(label)))

  assertions.forEach(label => {
    console.log("label:");
    console.log(label);
    labels.set(JSON.stringify(label), {
      centered: label.centered || false,
      label: label.name,
      x: label.x,
      y: label.y,
      paper: label.paper
    })
  })

  scheduleDraw()
}

const updateText = ({ assertions, retractions }) => {
  retractions.forEach(text => texts.delete(JSON.stringify(text)))

  assertions.forEach(text => {
    texts.set(JSON.stringify(text), {
      size: text.size,
      angle: text.angle,
      text: text.text,
      x: text.x,
      y: text.y
    })
  })
  scheduleDraw()
}

const updateLine = ({ retractions, assertions }) => {
  retractions.forEach(line => lines.delete(JSON.stringify(line)))

  assertions.forEach(line => {
    lines.set(JSON.stringify(line), {
      r: line.r || 0,
      g: line.g || 0,
      b: line.b || 0,
      x: line.x || 0,
      y: line.y || 0,
      xx: line.xx || 0,
      yy: line.yy || 0,
      paper: line.paper
    })
  })

  scheduleDraw()
}

const updateCircle = ({ assertions, retractions }) => {
  retractions.forEach(circle => circles.delete(JSON.stringify(circle)))

  assertions.forEach(circle => {
    circles.set(JSON.stringify(circle), {
      x: circle.x || 0,
      y: circle.y || 0,
      r: circle.r || 0,
      g: circle.g || 0,
      b: circle.b || 0,
      radius: circle.radius || 0
    })
  })

  scheduleDraw()
}

const updateHalo = ({ assertions, retractions }) => {
  retractions.forEach(halo => halos.delete(JSON.stringify(halo)))

  assertions.forEach(halo => {
    halos.set(JSON.stringify(halo), {
      x: halo.x,
      y: halo.y,
      r: halo.r,
      g: halo.g,
      b: halo.b,
      radius: halo.radius
    })
  })

  scheduleDraw()
}

const updateContainedPapers = ({ assertions }) => {
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
  // console.log(containedPapers)

  scheduleDraw()
}

const updateProjectorCalibration = ({ assertions }) => {
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

// Query labels
room.subscribe(`draw label $name at ($x, $y)`, updateLabel)
room.subscribe(`${namespace}: draw label $name at ($x, $y)`, updateLabel)
room.subscribe(`draw $centered label $name at ($x, $y)`, updateLabel)
room.subscribe(`${namespace}: draw $centered label $name at ($x, $y)`, updateLabel)
room.subscribe(`draw $centered label $name at ($x, $y) on paper $paper`, updateLabel)
// draw centered label "Hello" at (0.5, 0.5) on paper 395

// Query text
room.subscribe(`draw text $text at ($x, $y)`, updateText)
room.subscribe(`${namespace}: draw text $text at ($x, $y)`, updateText)

// Query small text
room.subscribe(`draw $size text $text at ($x, $y)`, updateText)
room.subscribe(`${namespace}: draw $size text $text at ($x, $y)`, updateText)

room.subscribe(`draw $size text $text at ($x, $y) at angle $angle`, updateText)
room.subscribe(
  `${namespace}: draw $size text $text at ($x, $y) at angle $angle`,
  updateText
)

// Query lines
room.subscribe(
  `draw a ($r, $g, $b) line from ($x, $y) to ($xx, $yy)`,
  updateLine
)
room.subscribe(
  `${namespace}: draw a ($r, $g, $b) line from ($x, $y) to ($xx, $yy)`,
  updateLine
)
room.subscribe(
  `draw a ($r, $g, $b) line from ($x, $y) to ($xx, $yy) on paper $paper`,
  updateLine
)
// draw a (1, 1, 255) line from (0.01, 0.01) to (1.0, 1.0) on paper 395
// draw a (1, 1, 255) line from (0.2, 0.2) to (0.8, 0.2) on paper 395
// draw a (1, 1, 255) line from (0.8, 0.2) to (0.8, 0.8) on paper 395
// draw a (1, 1, 255) line from (0.8, 0.8) to (0.2, 0.8) on paper 395
// draw a (1, 1, 255) line from (0.2, 0.8) to (0.2, 0.2) on paper 395

// Query circles
room.subscribe(
  `draw a ($r, $g, $b) circle at ($x, $y) with radius $radius`,
  updateCircle
)
room.subscribe(
  `${namespace}: draw a ($r, $g, $b) circle at ($x, $y) with radius $radius`,
  updateCircle
)
// Query halos
room.subscribe(
  `draw a ($r, $g, $b) halo around ($x, $y) with radius $radius`,
  updateHalo
)
room.subscribe(
  `${namespace}: draw a ($r, $g, $b) halo around ($x, $y) with radius $radius`,
  updateHalo
)

// Query where papers are
room.subscribe(
  `camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  updateContainedPapers
)

room.subscribe(
  `camera $cameraId has projector calibration TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  updateProjectorCalibration
)

const add_vec = (vec1, vec2) =>
  ({"x": vec1["x"] + vec2["x"], "y": vec1["y"] + vec2["y"]})

const diff_vec = (vec1, vec2) =>
  ({"x": vec1["x"] - vec2["x"], "y": vec1["y"] - vec2["y"]})

const scale_vec = (vec, scale) =>
  ({"x": vec["x"] * scale, "y": vec["y"] * scale})

const vec_length = (vec) =>
  Math.sqrt(vec["x"] * vec["x"] + vec["y"] * vec["y"])

const paper_approximation = (paper, perspT) => {
  const perspTCorner = corner => {
    const pt = perspT.transform(corner.x, corner.y)
    return {x: pt[0], y: pt[1]};
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

  context.fillStyle = '#fff'
  context.font = `${40 * window.devicePixelRatio}px sans-serif`

  labels.forEach(({ label, x, y, centered, paper }) => {
    context.save()
    if (centered === 'centered') {
      context.textBaseline = `middle`
      context.textAlign = `center`
    }
    console.log("paper", paper)
    if (!!paper && containedPapers.has(paper)) {
      containerPaper = containedPapers.get(paper);
      const paperApprox = paper_approximation(containerPaper, perspT);
      context.translate(
        normToCoord(paperApprox.origin.x, canvas.width),
        normToCoord(paperApprox.origin.y)
      );
      context.rotate(paperApprox.angle_radians)
      context.fillText(label, x * paperApprox.width, y * paperApprox.height);
    } else {
      context.fillText(label, normToCoord(x, canvas.width), normToCoord(y))
    }
    context.restore()
  })

  texts.forEach(({ text, x, y, size, angle }) => {
    context.save()
    context.fillStyle = '#9999ff'
    if (size === 'small') {
      context.font = `${20 * window.devicePixelRatio}px sans-serif`
    }
    if (typeof angle !== 'undefined') {
      context.translate(normToCoord(x, canvas.width), normToCoord(y))
      context.rotate(-angle) // counterclockwise
      context.translate(-normToCoord(x, canvas.width), -normToCoord(y))
    }
    context.fillText(text, normToCoord(x, canvas.width), normToCoord(y))
    context.restore()
  })

  circles.forEach(({ x, y, r, g, b, radius }) => {
    context.save()
    context.fillStyle = `rgb(${r},${g},${b})`
    context.beginPath()
    context.ellipse(
      normToCoord(x, canvas.width),
      normToCoord(y),
      normToCoord(radius),
      normToCoord(radius),
      0,
      0,
      2 * Math.PI
    )
    context.fill()
    context.restore()
  })

  halos.forEach(({ x, y, r, g, b, radius }) => {
    context.save()
    context.strokeStyle = `rgb(${r},${g},${b})`
    context.fillStyle = `rgb(${r},${g},${b},0)`
    context.beginPath()
    context.ellipse(
      normToCoord(x, canvas.width),
      normToCoord(y),
      normToCoord(radius),
      normToCoord(radius),
      0,
      0,
      2 * Math.PI
    )
    context.stroke()
    context.restore()
  })

  lines.forEach(({ x, y, xx, yy, r, g, b, paper }) => {
    context.save()
    context.strokeStyle = `rgb(${r},${g},${b})`
    context.beginPath()
    // console.log("PAPER:")
    // console.log(paper);
    // console.log(containedPapers)
    if (!!paper && containedPapers.has(paper)) {
      containerPaper = containedPapers.get(paper);
      const paperApprox = paper_approximation(containerPaper, perspT);
      context.translate(
        normToCoord(paperApprox.origin.x, canvas.width),
        normToCoord(paperApprox.origin.y)
      );
      context.rotate(paperApprox.angle_radians)
      context.moveTo(x * paperApprox.width, y * paperApprox.height);
      context.lineTo(xx * paperApprox.width, yy * paperApprox.height);
    } else {
      context.moveTo(x * canvas.width, y * canvas.height)
      context.lineTo(xx * canvas.width, yy * canvas.height)
    }
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
