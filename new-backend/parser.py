import os
from arpeggio import *
from arpeggio import RegExMatch as _

# Escaped strings inspired from this:
# https://github.com/PhilippeSigaud/Pegged/blob/master/pegged/examples/strings.d

def whitespace():       return _('\s')
def postfix():          return '$+'
def wildcard():         return '$'
def variable():         return wildcard, _('\w+')
def postfixvariable():  return postfix, _('\w+')
def escapedchar():      return ('\\', ['"', '\'', '\\'])
def anychar():          return _('.')
def char():             return [escapedchar, anychar, whitespace]
def string():           return '"', ZeroOrMore(Not('"'), char), '"'
def value():            return  _('[^"\s]+')
def number():           return _('-?\d+((\.\d*)?((e|E)(\+|-)?\d+)?)?')
def id():               return '#', _('\d+')
def term():             return [postfixvariable, variable, postfix, wildcard, id, string, number, value, whitespace]
def fact():             return OneOrMore(term), EOF


class CalcVisitor(PTNodeVisitor):
    def visit_string(self, node, children):
        if self.debug:
            print("Expression {}".format(children))
            print("Node {}".format(node))
            print("---")
        return ('text', "".join(children))

    def visit_value(self, node, children):
        if self.debug:
            print("value Expression {}".format(children))
            print("value Node {}".format(node))
            print("---")
        return ('text', str(node))

    def visit_id(self, node, children):
        if self.debug:
            print("ID Expression {}".format(children))
            print("ID Node {}".format(node))
        return ('id', str(children[0]))

    def visit_term(self, node, children):
        if self.debug:
            print("term Expression {}".format(children))
            print("term Node {}".format(node))
        if len(children) is 0:
            return None
        return children

    def visit_whitespace(self, node, children):
        return None

    def visit_variable(self, node, children):
        return ('variable', str(children[1]))

    def visit_postfix(self, node, children):
        return ('postfix', '')

    def visit_postfixvariable(self, node, children):
        return ('postfix', str(children[1]))

    def visit_wildcard(self, node, children):
        return ('variable', '')

    def visit_number(self, node, children):
        if self.debug:
            print("NUMBER")
            print(node)
            print(type(node))
            print("---")
            print(children)
            print("----------")
        value = node.value
        try:
            return ('integer', int(value))
        except:
            pass
        try:
            return ('float', float(value))
        except:
            return None

    def visit_fact(self, node, children):
        return [x[0] for x in children if x != None]


def parse(fact_string, debug=False):
    parser = ParserPython(fact, debug=debug, skipws=False)
    parse_tree = parser.parse(fact_string)
    # parse_tree can now be analysed and transformed to some other form
    # using e.g. visitor support. See http://igordejanovic.net/Arpeggio/semantics/
    result = visit_parse_tree(parse_tree, CalcVisitor(debug=debug))
    if debug:
        print(result)
    return result

if __name__ == "__main__":
    # In debug mode dot (graphviz) files for parser model
    # and parse tree will be created for visualization.
    # Checkout current folder for .dot files.
    testdata = '#0 bird1 "has" 8 toes'
    testdata1 = '#0 bird1 "has" 8 nice toes 0.5 1. .99999 1.23e8'
    testdata2 = '$ $Animal "has" $+TheRest1'
    testdata3 = '#0 "This \\"is\\" a test" one "two" $ $X $+ $+Z'
    parse(testdata3, debug=True)
