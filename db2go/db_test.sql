create table t0
(
    id                   int                not null
        primary key,
    c_tinyint            tinyint            null,
    c_smallint           smallint           null,
    c_mediumint          mediumint          null,
    c_int                int                null,
    c_bigint             bigint             null,
    c_tinyint_unsigned   tinyint unsigned   null,
    c_smallint_unsigned  smallint unsigned  null,
    c_mediumint_unsigned mediumint unsigned null,
    c_int_unsigned       int unsigned       null,
    c_bigint_unsigned    bigint unsigned    null,
    c_float              float              null,
    c_double             double             null,
    c_decimal            decimal(10, 5)     null,
    c_char               char(10)           null,
    c_varchar            varchar(20)        null,
    c_tinytext           tinytext           null,
    c_text               text               null,
    c_mediumtext         mediumtext         null,
    c_longtext           longtext           null,
    c_tinyblob           tinyblob           null,
    c_blob               blob               null,
    c_mediumblob         mediumblob         null,
    c_longblob           longblob           null,
    c_binary             binary(255)        null,
    c_time               time               null,
    c_timestamp          timestamp          null,
    c_date               date               null,
    c_datetime           datetime           null,
    c_year               year               null
);

create table t1
(
    id   int auto_increment
        primary key,
    name varchar(32) not null,
    constraint t1_name_uindex
        unique (name)
);

create table t2
(
    id   int auto_increment
        primary key,
    name varchar(32) null
);

create table t3
(
    id    int auto_increment
        primary key,
    t1_id int null,
    t2_id int null,
    constraint t3_t1_id_t2_id_uindex
        unique (t1_id, t2_id),
    constraint t3_t1_id_fk
        foreign key (t1_id) references t1 (id),
    constraint t3_t2_id_fk
        foreign key (t2_id) references t2 (id)
);

create table t4
(
    c1 int not null,
    c2 int not null,
    c3 int default 123 null,
    primary key (c1, c2)
);

