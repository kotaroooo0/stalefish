drop database if exists stalefish;
create database stalefish;
use stalefish;

drop table if exists documents;
create table documents (
    id integer not null auto_increment primary key,
    body text not null,
    token_count integer not null
);

drop table if exists tokens;
create table tokens (
    id integer not null auto_increment primary key,
    term varchar(512) not null unique
);

drop table if exists inverted_indexes;
create table inverted_indexes (
    token_id integer not null primary key,
    posting_list longblob not null
);

-- ユニットテスト用のDBを作成
drop database if exists stalefish_test;
create database stalefish_test;
use stalefish_test;

drop table if exists documents;
create table documents (
    id integer not null auto_increment primary key,
    body text not null,
    token_count integer not null
);

drop table if exists tokens;
create table tokens (
    id integer not null auto_increment primary key,
    term varchar(512) not null unique
);

drop table if exists inverted_indexes;
create table inverted_indexes (
    token_id integer not null primary key,
    posting_list longblob not null
);

