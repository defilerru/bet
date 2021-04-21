DROP TABLE IF EXISTS bets;
DROP TABLE IF EXISTS bets_users;
DROP TABLE IF EXISTS predictions;

CREATE TABLE predictions(
    id serial,
    created_at datetime,
    started_at datetime,
    finished_at datetime,
    error text default null,
    name char(64),
    option_1 char(32),
    option_2 char(32),
    winner ENUM('option_1', 'option_2'),
    PRIMARY KEY(id)
) engine = InnoDB;

CREATE TABLE bets_users(
    id serial,
    balance bigint unsigned default null
) ENGINE = InnoDB;

CREATE TABLE bets(
    user_id bigint unsigned not null,
    prediction_id bigint unsigned not null,
    amount bigint unsigned not null,
    closed boolean,
    `option` ENUM('option_1', 'option_2'),
    PRIMARY KEY(user_id, prediction_id),
    CONSTRAINT user_id_ft FOREIGN KEY(user_id) REFERENCES bets_users(id),
    CONSTRAINT pred_id_fk FOREIGN KEY(prediction_id) REFERENCES predictions(id)
) engine = InnoDB;