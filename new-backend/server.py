from timeit import default_timer as timer
import logging
import sqlite3

logging.basicConfig(level=logging.INFO)

conn = sqlite3.connect('example.db')  # ':memory:'
c = conn.cursor()

def get_all_subscriptions():
    e = c.execute('''
    SELECT DISTINCT t.source, t2."subscription_id"
    FROM (
        SELECT factid, source
        FROM facts
        WHERE position = 0 AND value = 'subscription'
    ) AS t
    -- get the subscription_id at position 1
    INNER JOIN (
        SELECT factid, source, value as "subscription_id"
        FROM facts
        WHERE position = 1
    ) AS t2 ON t.factid = t2.factid AND t.source = t2.source
    ''')
    return e.fetchall()


def query_extract_variables_and_query_parts(source, subscription_id):
    # query = [
    #     [("VAR", "X"), ("STR", "sees"), ("STR", "a"), ("VAR", "Y")],
    #     [("VAR", "Y"), ("STR", "has"), ("VAR", "Z"), ("STR", "toes")],
    # ]
    # get X, Y, TYPE, VALUE
    e = c.execute('''
    SELECT t3."subscription_part", facts.position - 3, facts.value, facts.type
    FROM facts
    -- filter to only subscriptions by the value 'subscription' at position 0
    LEFT JOIN (
        SELECT factid, source
        FROM facts
        WHERE position = 0 AND value = 'subscription'
    ) AS t ON facts.factid = t.factid AND facts.source = t.source
    -- get the subscription_id at position 1
    LEFT JOIN (
        SELECT factid, source, value as "subscription_id"
        FROM facts
        WHERE position = 1
    ) AS t2 ON t.factid = t2.factid AND t.source = t2.source
    -- get the subscription_part at position 2
    LEFT JOIN (
        SELECT factid, source, value as "subscription_part"
        FROM facts
        WHERE position = 2
    ) AS t3 ON t.factid = t3.factid AND t.source = t3.source
    -- select on things that start with $ to indicate a variable...
    WHERE facts.position >= 3
      AND facts.source = '{}'
      AND t2."subscription_id" = '{}'
    '''.format(source, subscription_id))
    # for row in e:
    #     print(row)
    return e.fetchall()


def query_facts(query):
    # query = [
    #     [("VAR", "X"), ("STR", "sees"), ("STR", "a"), ("VAR", "Y")],
    #     [("VAR", "Y"), ("STR", "has"), ("VAR", "Z"), ("STR", "toes")],
    # ]
    variables = {}
    for ix, x in enumerate(query):
        for iy, y in enumerate(x):
            if y[0] == "variable":
                if y[1] in variables:
                    variables[y[1]]["equals"].append({"fact": ix, "position": iy})
                else:
                    variables[y[1]] = {"fact": ix, "position": iy, "equals": []}
    sql = "SELECT DISTINCT\n"
    for i, v in enumerate(variables.keys()):
        if i != 0:
            sql += ",\n"
        sql += 'facts{}_{}.value as "{}"'.format(variables[v]["fact"], variables[v]["position"], v)
    sql += '\nFROM\n'
    for ix, x in enumerate(query):
        for iy, y in enumerate(x):
            if ix != 0 or iy != 0:
                sql += ',\n'
            sql += 'facts as facts{}_{}'.format(ix, iy)
    sql += '\nWHERE\n'
    for ix, x in enumerate(query):
        for iy, y in enumerate(x):
            sql += 'facts{}_0.factid = facts{}_{}.factid AND\n'.format(ix, ix, iy)
            sql += 'facts{}_{}.position = {} AND\n'.format(ix, iy, iy)
            if y[0] != "variable":
                sql += "facts{}_{}.type = '{}' AND\n".format(ix, iy, y[0])
            if y[0] == "text":
                sql += 'facts{}_{}.value = {} AND\n'.format(ix, iy, "'{}'".format(y[1]))
            elif y[0] != "variable":
                sql += 'facts{}_{}.value = {} AND\n'.format(ix, iy, y[1])
    for v in variables.values():
        for k in v["equals"]:
            sql += 'facts{}_{}.value = facts{}_{}.value AND\n'.format(v["fact"], v["position"], k["fact"], k["position"])
    if sql[-4:] == 'AND\n':
        sql = sql[:-4]
    logging.debug(sql)
    e = c.execute(sql)
    # for row in e:
    #     print(row)
    return e.fetchall()


def get_facts_for_subscription(source, subscription_id):
    query = []
    r = query_extract_variables_and_query_parts(source, subscription_id)
    logging.debug("----")
    for row in r:
        if row[0] >= len(query):
            query.append([])
        query[row[0]].append((row[3], row[2]))
    logging.debug(query)
    logging.debug("----------")
    return query_facts(query)


def measure1000(f):
    start = timer()
    for i in range(1000):
        f()
    end = timer()
    print((end - start)*1000)


def print_results(rows):
    for row in rows:
        logging.info(row)


def send_subscription_results(source, subscription_id, results):
    if len(results) > 0:
        logging.info("WOULD SEND SUBSCRIPTION RESULTS TO {} ({})".format(source, subscription_id))
        print_results(results)
    else:
        logging.info("No results for SUBSCRIPTION {} ({})".format(source, subscription_id))


def update_all_subscriptions():
    subscriptions = get_all_subscriptions()
    for source, subscription_id in subscriptions:
        facts = get_facts_for_subscription(source, subscription_id)
        send_subscription_results(source, subscription_id, facts)

# def foo():
#     get_facts_for_subscription('source394', '2lj43lkj34')
# measure1000(foo)

# TODO: claim fact(s)
# TODO: retract fact(s)
# TODO: select
# TODO: add subscription (should be just a regular claim)
# TODO: how to notify of both assertions and retractions? Is this needed?
update_all_subscriptions()

conn.close()
