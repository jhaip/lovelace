var longPollingActive = true;
var ignore_next_update = false;
var previousResultJSONString;
var CANVAS_WIDTH = 900;
var CANVAS_HEIGHT = 600;

function drawGraphics($rawCanvas, graphics) {
    var ctx = $rawCanvas.getContext('2d');

    ctx.fillStyle = '#000';
    ctx.fillRect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT);

    graphics.forEach(g => {
        let opt = g.options;
        if (g.type === "rectangle") {
            ctx.fillRect(opt.x, opt.y, opt.w, opt.h);
        } else if (g.type === "ellipse") {
            ctx.ellipse(opt.x, opt.y, opt.x, opt.w*0.5, opt.h*0.5, 0, 0, 2 * Math.PI);
        } else if (g.type === "text") {
            let lines = opt.text.split("\n");
            let lineHeight = ctx.measureText("X").height * 1.3;
            lines.forEach((line, i) => {
                ctx.fillText(line, opt.x, opt.y + i * lineHeight);
            });
        } else if (g.type === "fill") {
            if (typeof opt === "string") {
                ctx.fillStyle = opt;
            } else if (opt.length === 3) {
                ctx.fillStyle = `rgb(${opt[0]}, ${opt[1]}, ${opt[2]})`
            } else if (opt.length === 4) {
                ctx.fillStyle = `rgba(${opt[0]}, ${opt[1]}, ${opt[2]}, ${opt[3]})`
            }
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
//     [175, 156, 264, 61, 161, 279, 504, 330],
//     [
//         { "type": "fill", "options": "yellow" },
//         { "type": "rectangle", "options": { "x": 0, "y": 0, "w": CANVAS_WIDTH, "h": CANVAS_HEIGHT } },
//         { "type": "fill", "options": "green" },
//         { "type": "rectangle", "options": { "x": 10, "y": 10, "w": 60, "h": 60 } },
//         { "type": "rectangle", "options": { "x": 100, "y": 10, "w": 100, "h": 60 } },
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
