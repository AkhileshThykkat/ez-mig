# ez-mig
A UX-wraper around go migration tool for lazy people 


## Why
### Felt uneasy using go migrate std tool to do migration becuase of it's too explicit less dx workflow. So decided to create a wrapper to make a better dx exp.

## What else other than command shortening?
### Apart from short commands which we love, it also has session/db management to choose which db to apply the commands on. Allowing you to do migration from outside the targeted project, if you install it globally ( in case if that's you want ).



#### Pre-requisites
- Go installed
- GCC installed ( Just in case ) 
- Linux ( only support linux now, will be adding other platform support shortly.)
- Database installed ( PostgreSQL, MYSql ) will be adding other dbs support shortly.


## Installation

### From source

```bash
git clone https://github.com/AkhileshThykkat/ez-mig.git
cd ez-mig
go build
```

### From release

Download the latest release from the [releases page](https://github.com/AkhileshThykkat/ez-mig/releases).

#### Refer Commands.md for how to use.
