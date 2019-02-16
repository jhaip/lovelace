const { room, MY_ID_STR, run } = require('../helper2')(__filename);

room.on(`$source wish I was labeled %text`, results => {
  room.cleanup()
  results.push({source: MY_ID_STR, text: "ADD: wish I was labeled..."})
  results.forEach(result => {
    let ill = room.newIllumination()
    let textLines = result.text.split(" ");
    let maxLineCharacters = 13;
    let nextLine = ""
    let lineOffset = 0;
    ill.translate(10,10)
    textLines.forEach(textLine => {
      if (nextLine.length + 1 + textLine.length >= maxLineCharacters) {
        ill.text(0, lineOffset, nextLine)
        nextLine = textLine;
        lineOffset += 30;
      } else {
        if (nextLine.length > 0) {
          nextLine += " "
        }
        nextLine += textLine;
      }
    });

    ill.text(0, lineOffset, nextLine)
    room.draw(ill, result.source)

  })
})

run();