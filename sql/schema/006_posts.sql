-- +goose Up

create table posts (
                       id uuid primary key,
                       created_at timestamp not null,
                       updated_at timestamp not null,
                       title text not null,
                        description text,
                        published_at timestamp not null,
                        url text not null unique,
                        feed_id uuid references feeds(id) not null

);

-- +goose Down

drop table posts;