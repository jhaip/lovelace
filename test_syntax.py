import re

def parse_syntax(s):
    def matchFirst(s, specialParts, r):
        prog = re.compile(r, re.MULTILINE)
        for m in prog.finditer(s):
            return specialParts + [{"type": "when", "index": m.start(),
                                    "length": len(m.group(1))}]
    def matchSecond(s, specialParts, r):
        prog = re.compile(r, re.MULTILINE)
        for m in prog.finditer(s):
            return specialParts + [{"type": "otherwise", "index": m.start() + len(m.group(1)),
                                    "length": len(m.group(2))}]
    labels = []
    labels = matchFirst(s, labels, "(when [^:]*:)[\s\S]*otherwise:\n[\s\S]*$")
    labels = matchSecond(s, labels, "(when [^:]*:[\s\S]*)(otherwise:)\n[\s\S]*$")
    labels = matchFirst(s, labels, "(when [^:]*:)[\s\S]*end\n")
    labels = matchSecond(s, labels, "(when [^:]*:[\s\S]*)(end)\n")
    labels = matchFirst(s, labels, "(when [^:]*:)[\s\S]*$")
    labels = matchFirst(s, labels, "(claim [^\n]*)")
    labels = matchFirst(s, labels, "(retract [^\n]*)")
    labels = matchFirst(s, labels, "(cleanup\n)")
    labels.sort(key=lambda x: x["index"])
    head = 0
    chunks = []
    for part in labels:
        if part["index"] >= head:
            chunks.append({"type": "normal", "text": s[slice(head, part["index"])]})
            chunks.append({"type": part["type"], "text": s[slice(
                part["index"], part["index"] + part["length"])]})
            head = part["index"] + part["length"]
    if len(chunks) is 0:
        chunks.append({type: "normal", "text": s})
    return chunks


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
print(parse_syntax(s))

