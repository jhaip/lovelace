import sqlite3

conn = sqlite3.connect('example.db')  # ':memory:'
c = conn.cursor()

def init_table(conn, c):
    c.execute('''CREATE TABLE IF NOT EXISTS facts (
      id INTEGER PRIMARY KEY,
      factid INTEGER,
      position INTEGER,
      value,
      type TEXT
    )''')
    conn.commit()


def populate(conn, c):
    source = 'source0'
    facts = [
        (0, 0, source, 'source'),
        (0, 1, 'A', 'text'),
        (0, 2, 'sees', 'text'),
        (0, 3, 'a', 'text'),
        (0, 4, 'bird', 'text'),

        (1, 0, source, 'source'),
        (1, 1, 'B', 'text'),
        (1, 2, 'sees', 'text'),
        (1, 3, 'a', 'text'),
        (1, 4, 'snake', 'text'),

        (2, 0, source, 'source'),
        (2, 1, 'bird', 'text'),
        (2, 2, 'has', 'text'),
        (2, 3, 3, 'integer'),
        (2, 4, 'toes', 'text'),

        (3, 0, source, 'source'),
        (3, 1, 'snake', 'text'),
        (3, 2, 'has', 'text'),
        (3, 3, 'no', 'text'),
        (3, 4, 'toes', 'text'),
    ]
    c.executemany('INSERT INTO facts (factid, position, value, type) VALUES (?,?,?,?)', facts)
    conn.commit()

def populate_subscriptions(conn, c):
    source1 = 'source0'
    subscription_id1 = '2lj43lkj34'
    subscription_id2 = 'QOUERJKERO'
    # Subcription: V:X S:sees S:a V:Y, V:Y S:has V:Z S:toes
    # Subscription: V:X S:sees S:a S:snake
    facts = [
        (100, 0, source1, 'source'),
        (100, 1, 'subscription', 'text'),
        (100, 2, subscription_id1, 'text'),
        (100, 3, 0, 'integer'),
        (100, 4, 'X', 'variable'),
        (100, 5, 'sees', 'text'),
        (100, 6, 'a', 'text'),
        (100, 7, 'Y', 'variable'),

        (101, 0, source1, 'source'),
        (101, 1, 'subscription', 'text'),
        (101, 2, subscription_id1, 'text'),
        (101, 3, 1, 'integer'),
        (101, 4, 'Y', 'variable'),
        (101, 5, 'has', 'text'),
        (101, 6, 'Z', 'variable'),
        (101, 7, 'toes', 'text'),

        (102, 0, source1, 'source'),
        (102, 1, 'subscription', 'text'),
        (102, 2, subscription_id2, 'text'),
        (102, 3, 0, 'integer'),
        (102, 4, 'X', 'variable'),
        (102, 5, 'sees', 'text'),
        (102, 6, 'a', 'text'),
        (102, 7, 'snake', 'text'),
    ]
    c.executemany('INSERT INTO facts (factid, position, value, type) VALUES (?,?,?,?)', facts)
    conn.commit()


def init(conn, c):
    init_table(conn, c)
    populate(conn, c)
    populate_subscriptions(conn, c)


if __name__ == "__main__":
    init(conn, c)
