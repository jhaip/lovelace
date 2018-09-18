import sqlite3

conn = sqlite3.connect('example.db')  # ':memory:'
c = conn.cursor()

def init_table():
    c.execute('''CREATE TABLE IF NOT EXISTS facts (
      id INTEGER PRIMARY KEY,
      factid INTEGER,
      position INTEGER,
      value,
      type TEXT,
      source INTEGER
    )''')
    conn.commit()


def populate():
    source = 'source0'
    facts = [
        (0, 0, 'A', 'text', source),
        (0, 1, 'sees', 'text', source),
        (0, 2, 'a', 'text', source),
        (0, 3, 'bird', 'text', source),
        (1, 0, 'B', 'text', source),
        (1, 1, 'sees', 'text', source),
        (1, 2, 'a', 'text', source),
        (1, 3, 'snake', 'text', source),
        (2, 0, 'bird', 'text', source),
        (2, 1, 'has', 'text', source),
        (2, 2, 3, 'integer', source),
        (2, 3, 'toes', 'text', source),
        (3, 0, 'snake', 'text', source),
        (3, 1, 'has', 'text', source),
        (3, 2, 'no', 'text', source),
        (3, 3, 'toes', 'text', source),
    ]
    c.executemany('INSERT INTO facts (factid, position, value, type, source) VALUES (?,?,?,?,?)', facts)
    conn.commit()

def populate_subscriptions():
    source1 = 'source394'
    subscription_id1 = '2lj43lkj34'
    subscription_id2 = 'QOUERJKERO'
    # Subcription: V:X S:sees S:a V:Y, V:Y S:has V:Z S:toes
    # Subscription: V:X S:sees S:a S:snake
    facts = [
        (100, 0, 'subscription', 'text', source1),
        (100, 1, subscription_id1, 'text', source1),
        (100, 2, 0, 'integer', source1),
        (100, 3, 'X', 'variable', source1),
        (100, 4, 'sees', 'text', source1),
        (100, 5, 'a', 'text', source1),
        (100, 6, 'Y', 'variable', source1),
        (101, 0, 'subscription', 'text', source1),
        (101, 1, subscription_id1, 'text', source1),
        (101, 2, 1, 'integer', source1),
        (101, 3, 'Y', 'variable', source1),
        (101, 4, 'has', 'text', source1),
        (101, 5, 'Z', 'variable', source1),
        (101, 6, 'toes', 'text', source1),
        (102, 0, 'subscription', 'text', source1),
        (102, 1, subscription_id2, 'text', source1),
        (102, 2, 0, 'integer', source1),
        (102, 3, 'X', 'variable', source1),
        (102, 4, 'sees', 'text', source1),
        (102, 5, 'a', 'text', source1),
        (102, 6, 'snake', 'text', source1),
    ]
    c.executemany('INSERT INTO facts (factid, position, value, type, source) VALUES (?,?,?,?,?)', facts)
    conn.commit()


def init():
    init_table()
    populate()
    populate_subscriptions()


init()
