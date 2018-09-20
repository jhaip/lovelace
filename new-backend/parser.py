import os
from arpeggio import *
from arpeggio import RegExMatch as _

def whitespace():   return _('\s')
def string():       return '"', _('[^"]*'),'"'
def value():        return  _('[^"\s]+')
def number():       return _('-?\d+((\.\d*)?((e|E)(\+|-)?\d+)?)?')
def term():         return [string, number, value, whitespace]
def fact():         return OneOrMore(term), EOF


def main(debug=False):
    parser = ParserPython(fact, debug=debug)
    testdata = 'bird1 "has" 8 nice toes 0.5 1. .99999 1.23e8'
    parse_tree = parser.parse(testdata)
    # parse_tree can now be analysed and transformed to some other form
    # using e.g. visitor support. See http://igordejanovic.net/Arpeggio/semantics/

if __name__ == "__main__":
    # In debug mode dot (graphviz) files for parser model
    # and parse tree will be created for visualization.
    # Checkout current folder for .dot files.
    main(debug=True)
