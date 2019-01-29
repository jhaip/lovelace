import re

s = """
import x;

when x is ($x, $y),
     $x was y,
     fox says 'hello':

  retract $ test $
  cleanup
  console.log(x);
  claim the fox is out!
  if (x > 0) {
    claim nice nice nice
  }
otherwise:
  // test
end
"""

specialParts = []

def matchFirst(prog):
    for m in prog.finditer(s):
        specialParts.append({"type": "when", "index": m.start(),
                            "length": len(m.group(1))})

def matchSecond(prog):
    for m in prog.finditer(s):
        specialParts.append({"type": "otherwise", "index": m.start() + len(m.group(1)),
                            "length": len(m.group(2))})

# whenOtherwiseFunc
prog = re.compile("(when [^:]*:)[\s\S]*otherwise:\n[\s\S]*$", re.MULTILINE)
matchFirst(prog)
prog = re.compile("(when [^:]*:[\s\S]*)(otherwise:)\n[\s\S]*$", re.MULTILINE)
matchSecond(prog)
# whenEndFunc
prog = re.compile("(when [^:]*:)[\s\S]*end\n", re.MULTILINE)
matchFirst(prog)
prog = re.compile("(when [^:]*:[\s\S]*)(end)\n", re.MULTILINE)
matchSecond(prog)
# whenFunc
prog = re.compile("(when [^:]*:)[\s\S]*$", re.MULTILINE)
matchFirst(prog)
# claimFunc
prog = re.compile("(claim [^\n]*)", re.MULTILINE)
matchFirst(prog)
# retractFunc
prog = re.compile("(retract [^\n]*)", re.MULTILINE)
matchFirst(prog)
# cleanupFunc
prog = re.compile("(cleanup\n)", re.MULTILINE)
matchFirst(prog)

print(specialParts)
specialParts.sort(key=lambda x: x["index"])
print(specialParts)

head = 0
chunks = []
for part in specialParts:
    if part["index"] >= head:
        chunks.append({"type": "normal", "text": s[slice(head, part["index"])]})
        chunks.append({"type": part["type"], "text": s[slice(
            part["index"], part["index"] + part["length"])]})
        head = part["index"] + part["length"]

if len(chunks) is 0:
    chunks.append({type: "normal", "text": s})

print(chunks)
