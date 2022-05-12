CREATE TABLE links(
    id bigserial primary key ,
    active_link text not null ,
    history_link text not null ,
)



CREATE TABLE admins(
                      id bigserial primary key ,
                      login text not null ,
                      password_hash text not null ,
)

