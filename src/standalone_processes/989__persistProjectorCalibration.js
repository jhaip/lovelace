const fs = require('fs');
const { room, myId, scriptName, run } = require('../helper2')(__filename);

const savedCalibrationLocation = __filename.replace(scriptName, 'files/projectorCalibration.txt')

fs.readFile(savedCalibrationLocation, 'utf8', function(err, contents) {
  if (err) {
    room.assert(`camera 1 has projector calibration TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080) @ 1`)
    console.log("claiming default calibration because save file didn't exist")
    console.error(err);
  } else {
    console.log("loaded initial calibration:")
    console.log(contents);
    room.assert(contents);
  }
  // Listen for calibration updates and save them
  console.log("listening for changes to calibration")
  room.onGetSource('wisherId',
    `camera $cameraId has projector calibration TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $`,
    results => {
      results.forEach(({ wisherId, cameraId, x1, y1, x2, y2, x3, y3, x4, y4 }) => {
        const s = `camera ${cameraId || 0} has projector calibration TL (${x1 || 0}, ${y1 || 0}) TR (${x2 || 0}, ${y2 || 0}) BR (${x3 || 0}, ${y3 || 0}) BL (${x4 || 0}, ${y4 || 0}) @ 1`
        fs.writeFile(savedCalibrationLocation, s, function(err) {
          if (err) return console.log(err)
          console.log("The file was saved!");
        });
      })
    }
  );
  run();
});
