import os
from arpeggio import *
from arpeggio import RegExMatch as _

def whitespace():       return _('\s')
def postfix():          return '$+'
def wildcard():         return '$'
def variable():         return wildcard, _('\w+')
def postfixvariable():  return postfix, _('\w+')
def string():           return '"', _('((\\")|([^"]))*'), '"'
def value():            return  _('[^"\s]+')
def number():           return _('-?\d+((\.\d*)?((e|E)(\+|-)?\d+)?)?')
def id():               return '#', _('\d+')
def term():             return [postfixvariable, variable, postfix, wildcard, id, string, number, value, whitespace]
def fact():             return OneOrMore(term), EOF


def main(debug=False):
    parser = ParserPython(fact, debug=debug)
    testdata = '#0 bird1 "has" 8 toes'
    testdata1 = '#0 bird1 "has" 8 nice toes 0.5 1. .99999 1.23e8'
    testdata2 = '$ $Animal "has" $+TheRest1'
    testdata3 = '#0 "This \\"is\\" a test"'
    print(testdata3)
    parse_tree = parser.parse(testdata3)
    # parse_tree can now be analysed and transformed to some other form
    # using e.g. visitor support. See http://igordejanovic.net/Arpeggio/semantics/

if __name__ == "__main__":
    # In debug mode dot (graphviz) files for parser model
    # and parse tree will be created for visualization.
    # Checkout current folder for .dot files.
    main(debug=True)
