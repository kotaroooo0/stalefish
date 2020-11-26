create table settings(
    key text primary key,
    value text
);

create table documents(
    id integer primary key,
    title text not null,
    body text not null
);

create table tokens(
    id  integer primary key,
    token text not null,
    docs_count int not null,
    postings blob not null
);
