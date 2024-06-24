-- name: ListBufos :many
select
    name,
    (
        select
            cast(COALESCE(avg(value), -1) as INTEGER)
        from
            viewer_bufovote
        where
            bufo_id = name
    ) as rating
from
    viewer_bufo
order by
    name;

-- name: GetBufo :one
select
    name,
    (
        select
            avg(value)
        from
            viewer_bufovote
        where
            bufo_id = name
    ) as rating
from
    viewer_bufo
where
    name = $1;

-- name: CreateVote :one
insert into
    viewer_bufovote (value, created, bufo_id)
values
    ($1, $2, $3) returning *;

-- name: CreateBufo :one
insert into
    viewer_bufo (name, created)
values
    ($1, $2)
on conflict(name) do update set name = EXCLUDED.name
returning *;