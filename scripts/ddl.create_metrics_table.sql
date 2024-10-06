drop table if exists metrics.metrics;

create table metrics.metrics
(
    load_id       int              not null,
    metric_type   text             not null,
    metric_name   text             not null,
    gauge_value   double precision null,
    counter_value int              null
);
