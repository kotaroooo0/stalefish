drop database if exists stalefish;
create database stalefish;
use stalefish;

drop table if exists settings;
create table settings (
    `key` text not null,
    `value` text not null
);

drop table if exists documents;
create table documents (
    id integer not null primary key,
    title text not null,
    body text not null
);

drop table if exists tokens;
create table tokens (
    id  integer not null primary key,
    token text not null,
    docs_count int not null,
    postings blob not null
);
