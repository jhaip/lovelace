var longPollingActive = true;
var ignore_next_update = false;
var previousResultJSONString;
var CANVAS_WIDTH = 1920;
var CANVAS_HEIGHT = 1080;

function drawGraphics($rawCanvas, graphics) {
    var ctx = $rawCanvas.getContext('2d');

    ctx.fillStyle = '#000';
    ctx.fillRect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT);
    let fontSize = Math.max(Math.floor(CANVAS_WIDTH / 10), 1)
    ctx.font = `${fontSize}px Arial`;

    graphics.forEach(g => {
        let opt = g.options;
        if (g.type === "rectangle") {
            ctx.fillRect(opt.x, opt.y, opt.w, opt.h);
            ctx.strokeRect(opt.x, opt.y, opt.w, opt.h);
        } else if (g.type === "ellipse") {
            ctx.beginPath();
            ctx.ellipse(opt.x, opt.y, opt.x, opt.w*0.5, opt.h*0.5, 0, 0, 2 * Math.PI);
            ctx.fill();
            ctx.stroke();
        } else if (g.type === "text") {
            let lines = opt.text.split("\n");
            let lineHeight = fontSize * 1.3;
            lines.forEach((line, i) => {
                ctx.fillText(line, opt.x, opt.y + i * lineHeight);
            });
        } else if (g.type === "line") {
            ctx.beginPath();
            ctx.moveTo(opt[0], opt[1]);
            ctx.lineTo(opt[2], opt[3]);
            ctx.stroke();
        } else if (g.type === "polygon") {
            ctx.beginPath();
            ctx.moveTo(opt[0][0], opt[0][1]);
            for (let i = 1; i < opt.length; i += 1) {
                ctx.lineTo(opt[i][0], opt[i][1]);
            }
            ctx.fill();
            ctx.stroke();
        } else if (g.type === "fill" || g.type === "fontcolor") {
            if (typeof opt === "string") {
                ctx.fillStyle = opt;
            } else if (opt.length === 3) {
                ctx.fillStyle = `rgb(${opt[0]}, ${opt[1]}, ${opt[2]})`
            } else if (opt.length === 4) {
                ctx.fillStyle = `rgba(${opt[0]}, ${opt[1]}, ${opt[2]}, ${opt[3]})`
            }
        } else if (g.type === "stroke") {
            if (typeof opt === "string") {
                ctx.strokeStyle = opt;
            } else if (opt.length === 3) {
                ctx.strokeStyle = `rgb(${opt[0]}, ${opt[1]}, ${opt[2]})`
            } else if (opt.length === 4) {
                ctx.strokeStyle = `rgba(${opt[0]}, ${opt[1]}, ${opt[2]}, ${opt[3]})`
            }
        } else if (g.type === "nostroke") {
            ctx.strokeStyle = "rgba(1, 1, 1, 0)";
        } else if (g.type === "nofill") {
            ctx.fillStyle = "rgba(1, 1, 1, 0)";
        } else if (g.type === "strokewidth") {
            ctx.lineWidth = +opt;
        } else if (g.type === "fontsize") {
            fontSize = +opt;
            ctx.font = `${fontSize}px Arial`;
        } else if (g.type === "push") {
            ctx.save();
        } else if (g.type === "pop") {
            ctx.restore();
        } else if (g.type === "translate") {
            ctx.translate(+opt[0], +opt[1]);
        } else if (g.type === "rotate") {
            ctx.rotate(+opt);
        } else if (g.type === "scale") {
            ctx.scale(+opt[0], +opt[1]);
        } else {
            console.log(`unrecognized command:`)
            console.log(g);
        }
    });
}

function updatePerspectiveCanvas($rawCanvas, calibration) {
    // try to create a WebGL canvas (will fail if WebGL isn't supported)
    try {
        var canvas = fx.canvas();
    } catch (e) {
        alert(e);
        return;
    }

    var texture = canvas.texture($rawCanvas);
    const BASE_CALIBRATION = [0, 0, CANVAS_WIDTH, 0, CANVAS_WIDTH, CANVAS_HEIGHT, 0, CANVAS_HEIGHT];
    canvas
        .draw(texture)
        .perspective(
            BASE_CALIBRATION,
            calibration || BASE_CALIBRATION
        )
        .update();
    $rawCanvas.parentNode.insertBefore(canvas, $rawCanvas);
}

function update(calibration, graphics) {
    document.body.innerHTML = '';
    let $rawCanvas = document.createElement('canvas');
    $rawCanvas.setAttribute("class", "hide");
    $rawCanvas.setAttribute("width", CANVAS_WIDTH);
    $rawCanvas.setAttribute("height", CANVAS_HEIGHT);
    document.body.appendChild($rawCanvas);
    drawGraphics($rawCanvas, graphics);
    updatePerspectiveCanvas($rawCanvas, calibration);
}

// Test:
// update(
//     null,
//     [
//         { "type": "fill", "options": "yellow" },
//         { "type": "rectangle", "options": { "x": 0, "y": 0, "w": CANVAS_WIDTH, "h": CANVAS_HEIGHT } },
//         { "type": "fill", "options": "green" },
//         { "type": "rectangle", "options": { "x": 10, "y": 10, "w": 60, "h": 60 } },
//         { "type": "rectangle", "options": { "x": 100, "y": 10, "w": 100, "h": 60 } },
//         { "type": "text", "options": { "x": 200, "y": 200, "text": "Hello World!" } },
//     ]
// );

async function loop() {
    try {
        const response = await fetch('/status')
        const myJson = await response.json();
        const myJsonString = JSON.stringify(myJson)
        if (myJsonString !== previousResultJSONString) {
            if (ignore_next_update) {
                ignore_next_update = false;
            } else {
                update(myJson.calibration, myJson.graphics);
            }
        } else {
            console.log("ignoring update because nothing changed");
        }
        previousResultJSONString = myJsonString;
        if (longPollingActive) {
            setTimeout(function () {
                loop();
            }, 1000);
        }
    } catch (error) {
        console.error(error);
    }
}

loop();
