const Room = require('@living-room/client-js')
const fs = require('fs');
const path = require('path');

const scriptName = path.basename(__filename);
const scriptNameNoExtension = path.parse(scriptName).name;
const logPath = __filename.replace(scriptName, 'logs/' + scriptNameNoExtension + '.log')
const access = fs.createWriteStream(logPath)
process.stdout.write = process.stderr.write = access.write.bind(access);
process.on('uncaughtException', function(err) {
  console.error((err && err.stack) ? err.stack : err);
})

const room = new Room()

const myId = 1924
let fontSize = 32;
let fontHeight = fontSize / 1080.0;
let lineHeight = 1.3 * fontHeight;
const origin = [0.0001, 0.0001 + lineHeight]
let charWidth = fontHeight * 0.38;
const cursorColor = `(255, 128, 2)`
let cursorPosition = [10, 10]
let currentWidth = 1;
let currentHeight = 1;
let editorWidthCharacters = 1;
let editorHeightCharacters = 1;
let windowPosition = [0, 0]
let currentTargetName;
let currentSourceCode = "";

const correctCursorPosition = () => {
  const lines = currentSourceCode.split("\n");
  cursorPosition[1] = Math.max(0,
    Math.min(
      cursorPosition[1],
      Math.max(0, lines.length - 1)
    )
  );
  cursorPosition[0] = Math.max(0,
    Math.min(
      cursorPosition[0],
      Math.max(0, lines[cursorPosition[1]].length)
    )
  );
}

const correctWindowPosition = () => {
  if (cursorPosition[1] < windowPosition[1]) {
    windowPosition[1] = cursorPosition[1];
  } else if (cursorPosition[1] >= windowPosition[1] + editorHeightCharacters) {
    windowPosition[1] = Math.max(0, cursorPosition[1] - editorHeightCharacters + 1);
  }
}

const insertChar = (char) => {
  console.log("inserting char", char);
  const index = getCursorIndex();
  console.log("index is", index)
  currentSourceCode = [
    currentSourceCode.slice(0, index),
    char,
    currentSourceCode.slice(index)
  ].join('');
  if (char === "\n") {
    cursorPosition = [0, cursorPosition[1] + 1];
  } else {
    cursorPosition[0] += 1;
  }
  render();
}

const deleteChar = () => {
  const index = getCursorIndex();
  if (index > 0) {
    if (cursorPosition[0] === 0) {
      cursorPosition[1] = Math.max(0, cursorPosition[1] - 1);
      const lines = currentSourceCode.split("\n");
      cursorPosition[0] = lines[cursorPosition[1]].length;
    } else {
      cursorPosition[0] -= 1;
    }
    currentSourceCode = [
      currentSourceCode.slice(0, index-1),
      currentSourceCode.slice(index)
    ].join('');
    render();
  }
}

const getCursorIndex = () => {
  const lines = currentSourceCode.split("\n");
  const linesBeforeCursor = lines.slice(0, cursorPosition[1])
  return linesBeforeCursor.reduce((acc, line) => acc + line.length + 1, 0) + cursorPosition[0]
}

const render = () => {
  correctCursorPosition();
  correctWindowPosition();
  room.retract(`draw $ text $ at ($, $) on paper ${myId}`)
  room.retract(`draw a ${cursorColor} line from ($, $) to ($, $) on paper ${myId}`)
  let lines = ["Point at something!"]
  if (currentTargetName) {
    lines = currentSourceCode.replace(new RegExp(String.fromCharCode(34), 'g'), String.fromCharCode(9787)).split("\n")
    console.error(lines)
  }
  editorWidthCharacters = 1000;
  editorHeightCharacters = Math.floor(currentHeight / (fontSize * 1.3 * 0.5));
  console.log("editor height", editorHeightCharacters);
  lines.slice(windowPosition[1], windowPosition[1] + editorHeightCharacters).forEach((lineRaw, i) => {
    const line = lineRaw.substring(0, editorWidthCharacters);
    room.assert(`draw "${fontSize}pt" text "${line}" at (${origin[0]}, ${origin[1] + i * lineHeight}) on paper ${myId}`)
  });
  room.assert(
    `draw a ${cursorColor} line from ` +
    `(${origin[0] + cursorPosition[0] * charWidth}, ${origin[1] + (cursorPosition[1] - windowPosition[1]) * lineHeight})` +
    ` to ` +
    `(${origin[0] + cursorPosition[0] * charWidth}, ${origin[1] + (cursorPosition[1] - windowPosition[1]) * lineHeight - fontHeight})` +
    ` on paper ${myId}`
  );
}

console.error("HEllo from text editor")

room.subscribe(
  `paper ${myId} is pointing at paper $targetId`,
  `$targetName has paper ID $targetId`,
  `$targetName has source code $sourceCode`,
  `paper ${myId} has width $myWidth height $myHeight angle $ at ($, $)`,
  ({assertions, retractions}) => {
    console.error("got stuff")
    console.error(assertions)
    console.error(retractions)
    if (false && retractions.length > 0) {
      room.assert(`draw "${fontSize}pt" text "Point at something!" at (${origin[0]}, ${origin[1]}) on paper ${myId}`)
      currentTargetName = undefined;
      currentSourceCode = "";
      cursorPosition = [0, 0];
      render();
    }
    assertions.forEach(({targetId, targetName, sourceCode, myWidth, myHeight}) => {
      if (currentTargetName !== targetName) {
        currentTargetName = targetName;
        currentSourceCode = sourceCode;
      }
      curentWidth = myWidth;
      currentHeight = myHeight;
      render();
    })
  }
)

room.on(
  `keyboard $ typed key $key @ $`,
  ({ key }) => {
    console.log("key", key);
    insertChar(key);
  }
)

room.on(
  `keyboard $ typed special key $specialKey @ $`,
  ({ specialKey }) => {
    console.log("special key", specialKey);
    const special_key_map = {
      "enter": "\n",
      "space": " ",
      "tab": "\t",
      "doublequote": String.fromCharCode(34)
    }
    if (!!special_key_map[specialKey]) {
      insertChar(special_key_map[specialKey])
    } else if (specialKey === "up") {
      cursorPosition[1] -= 1;
      render();
    } else if (specialKey === "right") {
      cursorPosition[0] += 1;
      render();
    } else if (specialKey === "down") {
      cursorPosition[1] += 1;
      render();
    } else if (specialKey === "left") {
      cursorPosition[0] -= 1;
      render();
    } else if (specialKey === "backspace") {
      deleteChar();
    } else if (specialKey === "C-s") {
      console.log("TODO save / create new paper with the current code")
      console.log(currentSourceCode);
      const cleanSourceCode = currentSourceCode.replace(/\n/g, '\\n').replace(/"/g, String.fromCharCode(9787))
      room.assert(`wish ${currentTargetName} has source code "${cleanSourceCode}"`);
    } else if (specialKey === "C-p") {
      console.log(`wish ${currentTargetName} would be printed`)
      room.assert(`wish ${currentTargetName} would be printed`);
    }
  }
)
