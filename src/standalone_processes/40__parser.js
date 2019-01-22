function parse(x) {
  const importPrefix = "import room;\n\n"
  const runPostfix = "\n\nrun();"
  const whenprefix = x.replace(/when ([^:]*):([\s\S]*$)/g, (match, p1, p2) => {
    const middle = p1.split(",").map(a => a.trim()).join(`",\n        "`)
    return `room.on("${middle}",\n        results => {` + p2 + "})"
  }).replace(/claim ([^\n]*)/g, (match, p1) => {
    return `room.assert("${p1}")`;
  }).replace(/retract ([^\n]*)/g, (match, p1) => {
    return `room.retract("${p1}")`;
  });
  return importPrefix + whenprefix + runPostfix;
}
