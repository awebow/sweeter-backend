# Sweeter Backend
Sweeter Backend is the server of Sweeter using RESTful API written in Go.

# Installation
## Requirements
* MySQL or MariaDB
* Git
* Go

## Installation
1. Clone this repository.
```
$ git clone https://github.com/awebow/sweeter-backend
```

2. Move to following directory.
```
$ cd sweeter-backend/main
```

3. Install dependency packages.
```
$ go get -d ./...
```

4. Build
```
$ go build
```

5. Now you'll get executable binary.

## Configuration
### Create database tables
To create tables server needs, run this SQL.
```sql
CREATE TABLE `users` (
  `no` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `id` varchar(32) NOT NULL,
  `password` binary(32) NOT NULL,
  `name` varchar(100) NOT NULL,
  `picture` varchar(128) DEFAULT NULL,
  `register_at` datetime NOT NULL DEFAULT current_timestamp(),
  `withdraw_at` datetime DEFAULT NULL,
  PRIMARY KEY (`no`),
  UNIQUE KEY `users_UN` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8

CREATE TABLE `sweets` (
  `no` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `author` bigint(20) unsigned NOT NULL,
  `content` varchar(300) NOT NULL DEFAULT '',
  `sweet_at` datetime NOT NULL DEFAULT current_timestamp(),
  `delete_at` datetime DEFAULT NULL,
  PRIMARY KEY (`no`),
  KEY `tweets_users_FK` (`author`),
  CONSTRAINT `tweets_users_FK` FOREIGN KEY (`author`) REFERENCES `users` (`no`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8

CREATE TABLE `followings` (
  `user` bigint(20) unsigned NOT NULL,
  `target` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`user`,`target`),
  KEY `followings_user_IDX` (`user`) USING BTREE,
  KEY `followings_target_IDX` (`target`) USING BTREE,
  CONSTRAINT `followings_users_FK` FOREIGN KEY (`user`) REFERENCES `users` (`no`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `followings_users_FK_1` FOREIGN KEY (`target`) REFERENCES `users` (`no`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8
```

Finally, write config.json file on working directory.

Example:
```json
{
    "port": 8080,
    "signing_key": "alcks102dkascksa",
    "database": {
        "host": "localhost:7306",
        "name": "sweeter",
        "user": "sweeter",
        "password": "ds5s1qxl!"
    }
}
```