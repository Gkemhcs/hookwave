CREATE TABLE users  (
    id uuid not null ,
    name varchar(25) not null ,
    email varchar(30),
    PRIMARY KEY(id),
    created_at timestamp default now(),
)

CREATE TABLE environments (
    id uuid not null ,
    name varchar(39) not null ,
    userid uuid ,
    created_at timestamp,
    updated_at timestamp,
    FORIEN KEY(userid) references users(id),
    PRIMARY KEY(id) ,

)

CREATE TABLE applications (
    id uuid not null ,
    name varchar(39) not null ,
    environmentid uuid ,
    created_at timestamp,
    updated_at timestamp,
    FORIEN KEY(environmentid) references environment(id),
    PRIMARY KEY(id) ,
)

CREATE TABLE endpointconfigs(
    id uuid not null ,
    headers JSONB,
    queryparams JSONB,
    url varchar(20),
    method varchar(10)
     PRIMARY KEY(id) ,
)
CREATE TABLE endpoints (
    id uuid not null ,
    name varchar(30),
    applicationId uuid not null,
    endpointconfigid uuid,
    created_at timestamp not null   ,
    updated_at tmestamp ,
    FORIEN KEY(environmentid) references environment(id),
     FORIEN KEY(endpointconfigid) references endpointconfigs(id),
    PRIMARY KEY(id) ,
    hmac_secret
)

CREATE TABLE messages(
    id uuid not null ,
    payload JSONB  ,
    event_type varchar(20),
    created_at timestamp default now(),
    applicationId uuid not null,
    FORIEN KEY(applicationid) references application(id),
    PRIMARY KEY(id) ,
    
)


CREATE TABLE deliveries(
    id uuid not null ,
    messageid uuid not null ,
    endpointid uuid not null ,
    created_at timestamp,
    updated_at timestamp ,
    status varchar(20),
    locked_at timestamp,
    next_attempt_at timestamp,
    attempts int,
    max_attempts int,

)


