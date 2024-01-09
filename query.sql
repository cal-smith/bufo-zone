-- name: ListBufos :many
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
    viewer_bufo;

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
    name = ?;

-- name: CreateVote :one
insert into
    viewer_bufovote (value, created, bufo_id)
values
    (?, ?, ?) returning *;

-- name: CreateBufo :one
insert into
    viewer_bufo (name, created)
values
    (?, ?)
on conflict(name) do nothing
returning *;