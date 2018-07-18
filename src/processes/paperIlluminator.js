// Draw animals on the table
module.exports = Room => {
  const room = new Room()
  let lastTime = -1;

  room.on(
    `camera $cameraId sees papers $papersString @ $time`,
    `paperIlluminator is active`,
    ({ cameraId, papersString, time }) => {
      console.error("got message")
      if (time > lastTime) {
        lastTime = time;
        console.error(papersString);
        console.error(papersString.replace(/'/g, '"'));
        const papers = JSON.parse(papersString.replace(/'/g, '"'));
        console.error(papers);

        room
          .retract(`table: draw a ($, $, $) circle at ($, $) with radius 0.005`);

        papers.forEach(paper => {
          paper["corners"].forEach(corner => {
            const x = corner["x"]
            const y = corner["y"]
            room
              .assert(`table: draw a (255, 0, 0) circle at (${x}, ${y}) with radius 0.005`)
          })
        });
      }
  })

  let dots = [];

  const updatePapers = ({ assertions, retractions }) => {
    if (!assertions) {
      room
        .retract(`table: draw a ($, $, $) circle at ($, $) with radius 0.005`);
    }
    assertions.forEach(A => {
      console.error(A);
      const time = A.time;
      const papersString = A.papersString;
      const cameraId = A.cameraId;

      if (time > lastTime) {
        lastTime = time;
        console.error(papersString);
        console.error(papersString.replace(/'/g, '"'));
        const papers = JSON.parse(papersString.replace(/'/g, '"'));
        console.error(papers);

        room
          .retract(`table: draw a ($, $, $) circle at ($, $) with radius 0.005`);

        papers.forEach(paper => {
          const x = paper["corners"][0]["x"] * 1.0 / CAM_WIDTH
          const y = paper["corners"][0]["y"] * 1.0 / CAM_HEIGHT
          room
            .assert(`table: draw a (255, 0, 0) circle at (${x}, ${y}) with radius 0.005`)
        });
      }
    })
  }

  room.subscribe(`camera $cameraId sees papers $papersString @ $time`, updatePapers)

  room.assert('paperIlluminator is active')
}
