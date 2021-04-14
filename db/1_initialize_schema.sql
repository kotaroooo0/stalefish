drop database if exists stalefish;
create database stalefish;
use stalefish;

drop table if exists documents;
create table documents (
    id integer not null auto_increment primary key,
    body text not null
);

drop table if exists tokens;
create table tokens (
    id integer not null auto_increment primary key,
    term varchar(512) not null unique
);

drop table if exists inverted_indexes;
create table inverted_indexes (
    token_id integer not null primary key,
    posting_list blob not null,
    docs_count integer not null,
    positions_count integer not null
);
