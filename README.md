# Liquibase Lock Guard

This little project solves an annoying little issue, if you run microservices using Liquibase, e.g. in Spring, in kubernetes.
Sometimes it happens, that during the start phase of a microservice, liquibase places a lock in the database, which never
gets unlocked. Consequently, the microservice keeps on hanging forever. The healing operation as well as the observation
is quite easy. It sometime might be a greater damage, if one of our services is crash-looping for a long time.

This guard solves the problem.

## TL;DR - How to use?

You can use the helm package in this project to directly install it in your cluster:

```bash
helm --namespace <NS_WHERE_YOUR_SERVICES_RUN> install liquibase-lock-guard helm/liquibase-lock-guard
```

Then a small little pod appears and watches the evil constellation and cleans it up for you.

## Anatomy of the problem

There are several reasons, why and how the deadlock appears. Sometimes it happens, when the service pod is killed for 
some reason shortly after placing the lock, or if multiple replicas of the same services run into a race-condition to the lock.

However, a typical liquibase managed database consists of a table called `databasechangeloglock`. This table holds the lock
to prevent actual race-conditions, it makes sense. But if the column `locked` is not set back to 'f' during start phase
all services will keep on hanging on that point forever. If you can determine, that the lock was granted more than an hour ago,
it usually means that liquibase caused a deadlock.

## Common Environment

This issue mainly occurs in applications generated and deployed by [JHipster](https://github.com/jhipster/generator-jhipster), 
but can also occur for any Spring application using liquibase
