// Draw animals on the table
module.exports = Room => {
  const room = new Room()
  let lastTime = -1;

  room.on(
    `camera $cameraId sees dots $dotsString @ $time`,
    `dotIlluminator is active`,
    ({ cameraId, dotsString, time }) => {
      if (time > lastTime) {
        lastTime = time;
        console.error(dotsString);
        console.error(dotsString.replace(/'/g, '"'));
        const dots = JSON.parse(dotsString.replace(/'/g, '"'));
        console.error(dots);

        room
          .retract(`table: draw a ($, $, $) circle at ($, $) with radius 0.02`);

        dots.forEach(dot => {
          room
            .assert(`table: draw a (${dot.r}, ${dot.g}, ${dot.b}) circle at (${dot.x}, ${dot.y}) with radius 0.02`)
        });
      }
  })

  let dots = [];

  const updateDots = ({ assertions, retractions }) => {
    if (!assertions) {
      room
        .retract(`table: draw a ($, $, $) circle at ($, $) with radius 0.02`);
    }
    assertions.forEach(A => {
      console.error(A);
      const time = A.time;
      const dotsString = A.dotsString;
      const cameraId = A.cameraId;

      if (time > lastTime) {
        lastTime = time;
        console.error(dotsString);
        console.error(dotsString.replace(/'/g, '"'));
        const dots = JSON.parse(dotsString.replace(/'/g, '"'));
        console.error(dots);

        dots.forEach(dot => {
          room
            .assert(`table: draw a (${dot.r}, ${dot.g}, ${dot.b}) circle at (${dot.x}, ${dot.y}) with radius 0.02`)
        });
      }
    })
  }

  room.subscribe(`camera $cameraId sees dots $dotsString @ $time`, updateDots)

  room.assert('dotIlluminator is active')
}
