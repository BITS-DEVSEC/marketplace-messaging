create table message (
    id bigserial primary key,
    chat_id bigserial not null,
    "from" bigserial not null,
    "to" bigserial not null,
    content text,
    time timestamptz not null default(now())
);

create index chat_index on message (chat_id);

alter table message add constraint chat_message_fk foreign key (chat_id) references chat (id) on delete cascade on update no action;
