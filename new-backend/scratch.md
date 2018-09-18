// Subcribers = Query and a app to notify about it

// Facts
A sees a bird
S:A S:sees S:a S:bird
B sees a snake
S:A S:sees S:a S:snake
bird has 3 toes
S:bird S:has N:3 S:toes
snake has no toes
S:bird S:has S:no S:toes

// Query
$X sees a $Y,
$Y has $Z toes

bird has $x toes

snake has no toes

// Table definition
CREATE TABLE facts (
  id INTEGER PRIMARY KEY,
  factid INTEGER,
  postion INTEGER,
  value,
  source INTEGER
);

// insert
- how is the myfactid chosen? program keeps track of an autoincrementing ID?
INSERT INTO facts (factid, position, value, source)
VALUES
(myfactid, 0, mytokens[0], mysource),
(myfactid, 1, mytokens[1], mysource),
(myfactid, 2, mytokens[2], mysource);

//SELECT "bird has $x toes"
SELECT DISTINCT facts1_3.value
FROM
facts as facts1_1,
facts as facts1_2,
facts as facts1_3,
facts as facts1_4
WHERE
facts1_1.factid = facts1_2.factid AND
facts1_1.factid = facts1_3.factid AND
facts1_1.factid = facts1_4.factid AND
facts1_1.position = 0 AND
facts1_2.position = 1 AND
facts1_3.position = 2 AND
facts1_4.position = 3 AND
facts1_1.value = 'bird' AND
facts1_2.value = 'has' AND
facts1_4.value = 'toes'

//SELECT
// $X sees a $Y,
// $Y has $Z toes
SELECT DISTINCT
// List the variables
facts1_1.value as "x",
facts1_4.value as "y",
facts2_3.value as "z"
FROM
// one row for position and fact
facts as facts1_1,
facts as facts1_2,
facts as facts1_3,
facts as facts1_4,
facts as facts2_1,
facts as facts2_2,
facts as facts2_3,
facts as facts2_4
WHERE
// one row for each postion and fact
facts1_1.factid = facts1_2.factid AND
facts1_1.factid = facts1_3.factid AND
facts1_1.factid = facts1_4.factid AND
facts2_1.factid = facts2_2.factid AND
facts2_1.factid = facts2_3.factid AND
facts2_1.factid = facts2_4.factid AND
// one row for each postion and fact
facts1_1.position = 0 AND
facts1_2.position = 1 AND
facts1_3.position = 2 AND
facts1_4.position = 3 AND
facts2_1.position = 0 AND
facts2_2.position = 1 AND
facts2_3.position = 2 AND
facts2_4.position = 3 AND
// one row for each postion and fact
// BUT: skip variables
// if a variable is repeated, add a line for a previous value to be equal
facts1_2.value = 'sees' AND
facts1_3.value = 'a' AND
facts2_1.value = facts1_4.value AND
facts2_2.value = 'has' AND
facts2_4.value = 'toes'

// Query:
- need to identity which ones are variables
- and for each variable, if it's repeated: the fact # and position of others
that it's supposed to be equal to
Intermitent data strucutre built up by iterating through each fact and position:
variables: {
"x": {fact: 0, position: 0, equals: []},
"y": {fact: 0, position: 3, equals: [{fact: 1, position: 0}]},
"z": {fact: 1, position: 2, equals: []}
}
// Will need variables amp for everything so might as well keep it stored
// with the subscription
