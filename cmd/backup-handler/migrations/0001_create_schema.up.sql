CREATE TABLE backups (
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    id varchar(72),
    name varchar(72),
    namespace varchar(60),
    is_running integer(1),
    pod_name varchar(60),
    container_name varchar(60),
    command varchar(256)
);