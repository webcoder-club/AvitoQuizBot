CREATE TABLE answers
(
  id            SERIAL NOT NULL
    CONSTRAINT answers_pkey
    PRIMARY KEY,
  user_id       INTEGER,
  speciality_id INTEGER,
  question_id   INTEGER,
  text          TEXT,
  created_at    TIMESTAMP
);

CREATE TABLE questions
(
  id            SERIAL NOT NULL
    CONSTRAINT questions_pkey
    PRIMARY KEY,
  speciality_id INTEGER,
  text          VARCHAR(200),
  answer_1      VARCHAR(50),
  answer_2      VARCHAR(50),
  answer_3      VARCHAR(50),
  answer_4      VARCHAR(50),
  correct       INTEGER
);

CREATE TABLE specialities
(
  id   SERIAL NOT NULL
    CONSTRAINT specialities_pkey
    PRIMARY KEY,
  name VARCHAR(50)
);

CREATE TABLE users
(
  id            SERIAL NOT NULL
    CONSTRAINT users_pkey
    PRIMARY KEY,
  tmid          INTEGER,
  user_name     VARCHAR(50),
  first_name    VARCHAR(50),
  last_name     VARCHAR(50),
  speciality_id INTEGER
);

