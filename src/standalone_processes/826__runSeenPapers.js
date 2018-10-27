const { room, myId, scriptName, run } = require('../helper2')(__filename);

let knownPapers = [];

room.on(
  `$ $processName has paper ID $paperId`,
  results => {
    knownPapers = results;
    console.log("NEW knownPapers:")
    console.log(knownPapers);
  }
)

room.on(
  `$ camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  results => {
    room.cleanup();
    console.error("camera sees paper")
    console.error(results)
    const visibleIDs = results.map(paper => String(paper.id))
    console.log("knownPapers", knownPapers)
    console.log("visibleIDs", visibleIDs)
    const bootPapers = ["0", "498", "577", "826", "277", "620", "1459", "1800", "1382", "1900", "989"]

    knownPapers.forEach(paper => {
      const processName = paper.processName;
      const paperId = String(paper.paperId);
      if (visibleIDs.includes(paperId)) {
        console.error(`wish "${processName}" would be running`)
        room.assert(`wish`, ["text", processName], `would be running`);
      } else if (!bootPapers.includes(paperId)) {
        console.error(`RETRACT: wish "${processName}" would be running`)
        room.retract(`#${myId} wish`, ["text", processName], `would be running`);
      }
    });
  }
)

room.on(
  `$ camera 1 sees no papers @ $time`,
  results => {
    console.log("no papers, stopping all programs")
    knownPapers.forEach(paper => {
      room.retract(`#${myId} wish`, ["text", processName], `would be running`);
    });
  }
)
