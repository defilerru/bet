-- USE defiler_test;

DROP TABLE IF EXISTS bets;
DROP TABLE IF EXISTS predictions;
DROP TABLE IF EXISTS bets_users;

CREATE TABLE bets_users(
    id serial,
    username char(32),
    balance bigint unsigned default null,
    moderator boolean
) ENGINE = InnoDB;

CREATE TABLE predictions(
    id serial,
    created_by bigint unsigned not null,
    created_at timestamp not null DEFAULT CURRENT_TIMESTAMP,
    started_at timestamp null,
    finished_at timestamp null,
    error text default null,
    name char(32),
    option_1 char(16),
    option_2 char(16),
    opt1_won bool NOT NULL DEFAULT false,
    start_delay_seconds smallint unsigned not null,
    CONSTRAINT created_by_fk FOREIGN KEY(created_by) REFERENCES bets_users(id),
    PRIMARY KEY(id),
    KEY(created_at)
) engine = InnoDB DEFAULT CHARSET=utf8mb4;

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

INSERT INTO bets_users (id, balance, username, moderator) VALUES (41, 10000, "user41", false);
INSERT INTO bets_users (id, balance, username, moderator) VALUES (42, 10000, "user42", true);
INSERT INTO bets_users (id, balance, username, moderator) VALUES (43, 10000, "user43", false);
INSERT INTO bets_users (id, balance, username, moderator) VALUES (44, 10000, "user44", false);
