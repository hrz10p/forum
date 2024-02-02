PRAGMA foreign_keys = ON;


CREATE TABLE IF NOT EXISTS users (
                       id VARCHAR PRIMARY KEY,
                       username VARCHAR UNIQUE,
                       email VARCHAR UNIQUE,
                       password VARCHAR(60)
);

CREATE TABLE posts (
                       id INTEGER PRIMARY KEY AUTOINCREMENT,
                       title VARCHAR,
                       content VARCHAR,
                       UID VARCHAR,
                       FOREIGN KEY (UID) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE comments (
                          id INTEGER PRIMARY KEY AUTOINCREMENT,
                          uid VARCHAR,
                          post_id INTEGER,
                          content VARCHAR,
                          FOREIGN KEY (uid) REFERENCES users(id) ON DELETE CASCADE,
                          FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE TABLE categories (
                            id INTEGER PRIMARY KEY AUTOINCREMENT,
                            name VARCHAR
);

CREATE TABLE post_cats (
                           post_id INTEGER,
                           category_id INTEGER,
                           PRIMARY KEY (post_id, category_id),
                           FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
                           FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE sessions (
                          id VARCHAR,
                          uid VARCHAR,
                          expireTime DATE,
                          PRIMARY KEY (id, uid),
                          FOREIGN KEY (uid) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE posts_reactions (
                                 user_id VARCHAR,
                                 post_id INTEGER,
                                 sign INTEGER,
                                 PRIMARY KEY (user_id, post_id),
                                 FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
                                 FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE TABLE comments_reactions (
                                    user_id VARCHAR,
                                    comment_id INTEGER,
                                    sign INTEGER,
                                    PRIMARY KEY (user_id, comment_id),
                                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
                                    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
);

INSERT INTO categories (name) VALUES
                                  ('Category 1'),
                                  ('Category 2'),
                                  ('Category 3'),
                                  ('Category 4'),
                                  ('Category 5'),
                                  ('Category 6'),
                                  ('Category 7'),
                                  ('Category 8'),
                                  ('Category 9'),
                                  ('Category 10');

