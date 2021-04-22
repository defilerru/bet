USE defiler_test;

DROP TABLE IF EXISTS bets;
DROP TABLE IF EXISTS bets_users;
DROP TABLE IF EXISTS predictions;

CREATE TABLE predictions(
    id serial,
    created_by bigint unsigned not null,
    created_at timestamp not null DEFAULT CURRENT_TIMESTAMP,
    started_at timestamp null,
    finished_at timestamp null,
    error text default null,
    name char(64),
    option_1 char(32),
    option_2 char(32),
    winner ENUM('option_1', 'option_2'),
    start_delay_seconds smallint unsigned not null,
    CONSTRAINT created_by_fk FOREIGN KEY(created_by) REFERENCES bets_users(id),
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
    on_first_option boolean not null,
    created_at timestamp not null default current_timestamp,
    PRIMARY KEY(user_id, prediction_id),
    CONSTRAINT user_id_ft FOREIGN KEY(user_id) REFERENCES bets_users(id),
    CONSTRAINT pred_id_fk FOREIGN KEY(prediction_id) REFERENCES predictions(id)
) engine = InnoDB;

INSERT INTO bets_users (id, balance) VALUES (41, 10000);
INSERT INTO bets_users (id, balance) VALUES (42, 10000);
INSERT INTO bets_users (id, balance) VALUES (43, 10000);
INSERT INTO bets_users (id, balance) VALUES (44, 10000);