const { room, myId } = require('../helper')(__filename);

room.on(`$ wish everything would be printed`, async (options) => {
  room.retract(`$ wish everything would be printed`)
  const papers = (await room.select(`$ $processName has paper ID $paperId`))
  console.log("papers:")
  console.log(papers);
  papers.forEach(p => {
    const processName = p.processName.word || p.processName.value;
    room.assert(`#${myId} wish paper ${p.paperId.value} at "${processName}" would be printed`)
  })
});
