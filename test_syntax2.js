function parseWithStates(x) {
    const STATES = { "GLOBAL": "GLOBAL", "WHEN_QUERY_PARAMS": "WHEN_QUERY_PARAMS", "WHEN_TRUE": "WHEN_TRUE", "WHEN_OTHERWISE": "WHEN_OTHERWISE", "WHEN_NEW_RESULTS_QUERY_PARAMS": "WHEN_NEW_RESULTS_QUERY_PARAMS", "WHEN_NEW_RESULTS": "WHEN_NEW_RESULTS" }
    let STATE = STATES.GLOBAL;
    let WHEN_VARIABLES_CACHE = "";
    let OUTPUT = "";
    OUTPUT += "const { room, myId, run } = require('../helper2')(__filename);\n\n"

    const claimFunc = s => {
        const m = s.match(/^(\s*)claim (.+)\s*$/);
        if (m === null) return "";
        return `${m[1]}room.assert(\`${m[2]}\`)\n`;
    }
    const retractFunc = s => {
        const m = s.match(/^(\s*)retract (.+)\s*$/);
        if (m === null) return "";
        return `${m[1]}room.retractAll(\`${m[2]}\`)\n`;
    }
    const cleanupFunc = s => {
        const m = s.match(/^(\s*)cleanup\s*$/);
        if (m === null) return "";
        return `${m[1]}room.cleanup()\n`;
    }
    const getUniqueVariables = s => {
        const variables = s.match(/\$([a-zA-Z0-9]+)/g);
        if (variables) {
            return variables
                .map(x => x.slice(1))
                .filter((value, index, self) => self.indexOf(value) === index);
        }
        return [];
    }

    const lines = x.split("\n")
    for (let lineIndex = 0; lineIndex < lines.length; lineIndex += 1) {
        const prevOUTPUT = OUTPUT.slice();
        const line = lines[lineIndex];
        const isLastLine = lineIndex === lines.length - 1;
        if (STATE === STATES.GLOBAL) {
            OUTPUT += claimFunc(line);
            OUTPUT += retractFunc(line);
            OUTPUT += cleanupFunc(line);
            if (line.match(/^when new \$results of /)) {
                STATE = STATES.WHEN_NEW_RESULTS_QUERY_PARAMS;
                if (line.slice(-1) === ':') {
                    const query = line.match(/^when new \$results of (.+):$/)[1]
                    const variables = getUniqueVariables(query);
                    const bySourceQueryMatch = line.match(/^when new \$results of (.+) by \$source:$/)
                    if (bySourceQueryMatch) {
                        OUTPUT += `room.onGetSource('source', \`${bySourceQueryMatch[1]}\`,\n`
                    } else {
                        OUTPUT += `room.on(\`${query}\`,\n`
                    }
                    OUTPUT += `        results => {\n`
                    STATE = STATES.WHEN_NEW_RESULTS;
                } else if (line.slice(-1) === ',') {
                    const query = line.match(/^when new \$results of (.+),$/)[1]
                    OUTPUT += `room.on(\`${query}\`,\n`
                } else {
                    console.error("bad when query!")
                }
            } else if (line.match(/^when /)) {
                STATE = STATES.WHEN_QUERY_PARAMS;
                if (line.slice(-1) === ':') {
                    const query = line.match(/^when (.+):$/)[1]
                    const variables = getUniqueVariables(query);
                    const bySourceQueryMatch = line.match(/^when (.+) by \$source:$/)
                    if (bySourceQueryMatch) {
                        OUTPUT += `room.onGetSource('source', \`${bySourceQueryMatch[1]}\`,\n`
                    } else {
                        OUTPUT += `room.on(\`${query}\`,\n`
                    }
                    OUTPUT += `        results => {\n`
                    OUTPUT += `  subscriptionPrefix();\n`
                    OUTPUT += `  if (!!results) {\n`
                    OUTPUT += `    results.forEach(({ ${variables.join(", ")} }) => {\n`
                    STATE = STATES.WHEN_TRUE;
                } else if (line.slice(-1) === ',') {
                    const query = line.match(/^when (.+),$/)[1]
                    WHEN_VARIABLES_CACHE = query;
                    OUTPUT += `room.on(\`${query}\`,\n`
                } else {
                    console.error("bad when query!")
                }
            }
        } else if (STATE == STATES.WHEN_TRUE) {
            OUTPUT += claimFunc(line);
            OUTPUT += retractFunc(line);
            OUTPUT += cleanupFunc(line);
            if (line.match(/^otherwise:$/)) {
                STATE = STATES.WHEN_OTHERWISE;
                OUTPUT += "\n    });\n  } else {\n"
            }
            if (line.match(/^end$/) || isLastLine) {
                STATE = STATES.GLOBAL;
                OUTPUT += "\n    });\n";
                OUTPUT += "  }\n  subscriptionPostfix();\n})\n";
            }
        } else if (STATE == STATES.WHEN_OTHERWISE) {
            OUTPUT += claimFunc(line);
            OUTPUT += retractFunc(line);
            OUTPUT += cleanupFunc(line);
            if (line.match(/^end$/) || isLastLine) {
                STATE = STATES.GLOBAL;
                OUTPUT += "  }\n  subscriptionPostfix();\n})\n";
            }
        } else if (STATE == STATES.WHEN_NEW_RESULTS) {
            OUTPUT += claimFunc(line);
            OUTPUT += retractFunc(line);
            OUTPUT += cleanupFunc(line);
            if (line.match(/^end$/) || isLastLine) {
                STATE = STATES.GLOBAL;
                OUTPUT += "\n})\n";
            }
        } else if (STATE == STATES.WHEN_QUERY_PARAMS) {
            const m = line.match(/^\s*(.+),$/)
            if (m) {
                const query = m[1]
                WHEN_VARIABLES_CACHE += ' ' + query;
                OUTPUT += `        \`${query}\`,\n`
            } else {
                const m2 = line.match(/^\s*(.+):$/)
                if (m2) {
                    const query = m2[1]
                    WHEN_VARIABLES_CACHE += ' ' + query;
                    const variables = getUniqueVariables(WHEN_VARIABLES_CACHE)
                    OUTPUT += `        \`${query}\`,\n`
                    OUTPUT += `        results => {\n`
                    OUTPUT += `  subscriptionPrefix();\n`
                    OUTPUT += `  if (!!results) {\n`
                    OUTPUT += `    results.forEach(({ ${variables.join(", ")} }) => {\n`
                    STATE = STATES.WHEN_TRUE;
                } else {
                    console.error("BAD QUERY")
                }
            }
        } else if (STATE == STATES.WHEN_NEW_RESULTS_QUERY_PARAMS) {
            const m = line.match(/^\s*(.+),$/)
            if (m) {
                const query = m[1]
                OUTPUT += `        \`${query}\`,\n`
            } else {
                const m2 = line.match(/^\s*(.+):$/)
                if (m2) {
                    const query = m2[1]
                    OUTPUT += `        \`${query}\`,\n`
                    OUTPUT += `        results => {\n`
                    STATE = STATES.WHEN_NEW_RESULTS;
                } else {
                    console.error("BAD QUERY")
                }
            }
        }
        if (prevOUTPUT === OUTPUT) {
            OUTPUT += line + "\n";
        }
    }

    if (STATE === STATES.WHEN_QUERY_PARAMS || STATE === STATES.WHEN_NEW_RESULTS_QUERY_PARAMS) {
        console.error("NO END TO WHEN QUERY")
    }
    OUTPUT += "\n\nrun();\n"
    return OUTPUT;
}