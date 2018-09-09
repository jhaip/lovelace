const { room, myId } = require('../helper')(__filename);

// DRAW PAPERS
room.subscribe(
  `$ camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  ({ assertions }) => {
    if (!assertions || assertions.length === 0) {
      return;
    }
    console.log("ASSERTIONS:")
    console.log(assertions);
    const visibleIDs = assertions.map(paper => String(paper.id))
    room.retract(`#${myId} draw a (255, 255, 1) line from ($, $) to ($, $)`)
    room.retract(`#${myId} draw text $ at ($, $)`)
    assertions.forEach(p => {
      W = 1920.0;
      H = 1080.0;
      if (
        !isNaN(p.x1) && !isNaN(p.y1) &&
        !isNaN(p.x2) && !isNaN(p.y2) &&
        !isNaN(p.x3) && !isNaN(p.y3) &&
        !isNaN(p.x4) && !isNaN(p.y4)
      ) {
        const margin = 0.1;
        const low = 0.01 + margin;
        const high = 1.0 - margin;
        room.assert(`#${myId} draw a (255, 255, 1) line from (${low}, ${low}) to (${high}, ${low}) on paper ${p.id}`);
        room.assert(`#${myId} draw a (255, 255, 1) line from (${high}, ${low}) to (${high}, ${high}) on paper ${p.id}`);
        room.assert(`#${myId} draw a (255, 255, 1) line from (${high}, ${high}) to (${low}, ${high}) on paper ${p.id}`);
        room.assert(`#${myId} draw a (255, 255, 1) line from (${low}, ${high}) to (${low}, ${low}) on paper ${p.id}`);
        room.assert(`#${myId} draw centered label "What is ${p.id}" at (0.5, 0.5) on paper ${p.id}`);
      }
    })
  }
)

room.subscribe(
  `$ camera 1 sees no papers @ $time`,
  ({ assertions }) => {
    if (!assertions || assertions.length === 0) {
      return;
    }
    // this may not do anything because this program might not be running
    // when there are no papers. If it's not running then it can't retract it.
    room.retract(`#${myId} draw a (255, 255, 1) line from ($, $) to ($, $) on paper $`)
    room.retract(`#${myId} draw centered label $ at ($, $) on paper $`);
  }
)
