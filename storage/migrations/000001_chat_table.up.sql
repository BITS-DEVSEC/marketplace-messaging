create table chat (
    id bigserial primary key,
    participants bigint[] check(array_length(participants, 1) = 2) not null,
    created_at timestamptz not null default(now()),
    deleted_at timestamptz
);
