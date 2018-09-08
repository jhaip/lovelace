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
const myId = (scriptName.split(".")[0]).split("__")[0]

const room = new Room()

const savedCalibrationLocation = __filename.replace(scriptName, 'files/projectorCalibration.txt')

fs.readFile(savedCalibrationLocation, 'utf8', function(err, contents) {
    if (err) return console.error(err);
    console.log("loaded initial calibration:")
    console.log(contents);
    room.assert(contents);
});

// Listen for calibration updates and save them
room.on(
  `$wisherId camera $cameraId has projector calibration TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $`,
  ({ wisherId, cameraId, x1, y1, x2, y2, x3, y3, x4, y4 }) => {
    const millis = (new Date()).getTime()
    const s = `#${wisherId || 0} camera ${cameraId || 0} has projector calibration TL (${x1 || 0}, ${y1 || 0}) TR (${x2 || 0}, ${y2 || 0}) BR (${x3 || 0}, ${y3 || 0}) BL (${x4 || 0}, ${y4 || 0}) @ ${millis}`
    fs.writeFile(savedCalibrationLocation, s, function(err) {
        if (err) return console.log(err)
        console.log("The file was saved!");
    });
  }
);
